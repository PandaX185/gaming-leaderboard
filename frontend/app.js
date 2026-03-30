const state = {
    ws: null,
    apiBase: "http://localhost:8080",
    gameId: "",
    players: new Map(),
    flashUntil: new Map(),
    playerNameCache: new Map(),
    pendingNameRequests: new Set(),
};

const ui = {
    apiBaseUrl: document.getElementById("apiBaseUrl"),
    gameSelect: document.getElementById("gameSelect"),
    gameId: document.getElementById("gameId"),
    connectBtn: document.getElementById("connectBtn"),
    disconnectBtn: document.getElementById("disconnectBtn"),
    status: document.getElementById("status"),
    eventMeta: document.getElementById("eventMeta"),
    leaderboardBody: document.getElementById("leaderboardBody"),
};

ui.connectBtn.addEventListener("click", connect);
ui.disconnectBtn.addEventListener("click", disconnect);
ui.gameSelect.addEventListener("change", onGameSelect);
ui.apiBaseUrl.addEventListener("change", () => loadGames(normalizeBase(ui.apiBaseUrl.value.trim())));

// Load games on page load
window.addEventListener("load", () => loadGames(normalizeBase(ui.apiBaseUrl.value.trim())));

async function loadGames(apiBase) {
    const url = `${apiBase}/api/v1/games?page=1&page_size=100`;

    try {
        const res = await fetch(url);
        if (!res.ok) {
            ui.gameSelect.innerHTML = '<option value="">Failed to load games</option>';
            return;
        }

        const payload = await res.json();
        const games = Array.isArray(payload.items) ? payload.items : [];

        if (games.length === 0) {
            ui.gameSelect.innerHTML = '<option value="">No games available</option>';
            return;
        }

        ui.gameSelect.innerHTML = '<option value="">Select a game...</option>' + games
            .map(game => `<option value="${game.id}">${game.name}</option>`)
            .join("");
    } catch (e) {
        console.error("Failed to load games:", e);
        ui.gameSelect.innerHTML = '<option value="">Error loading games</option>';
    }
}

function onGameSelect(event) {
    const gameId = event.target.value;
    if (gameId) {
        ui.gameId.value = gameId;
    }
}

function connect() {
    const gameId = ui.gameId.value.trim();
    if (!gameId) {
        setStatus("Game ID is required", false);
        return;
    }
    if (!isMongoObjectId(gameId)) {
        setStatus("Game ID must be a 24-char hex Mongo ObjectID", false);
        ui.eventMeta.textContent = "Use game.id from GET /api/v1/games, not game name.";
        return;
    }

    disconnect();

    const apiBase = normalizeBase(ui.apiBaseUrl.value.trim());

    state.gameId = gameId;
    state.apiBase = apiBase;
    state.players.clear();
    state.flashUntil.clear();
    state.playerNameCache.clear();
    state.pendingNameRequests.clear();
    renderLeaderboard();

    loadInitialLeaderboard(apiBase, gameId).finally(() => {
        openSocket(apiBase, gameId);
    });
}

function disconnect() {
    if (state.ws) {
        state.ws.onclose = null;
        state.ws.close();
        state.ws = null;
    }

    ui.connectBtn.disabled = false;
    ui.disconnectBtn.disabled = true;
    setStatus("Offline", false);
}

function openSocket(apiBase, gameId) {
    const wsBase = toWsBase(apiBase);
    const wsUrl = `${wsBase}/api/v1/games/${encodeURIComponent(gameId)}/leaderboard/ws`;

    setStatus(`Connecting: ${wsUrl}`, false);

    const ws = new WebSocket(wsUrl);
    state.ws = ws;

    ws.onopen = () => {
        setStatus("Online", true);
        ui.connectBtn.disabled = true;
        ui.disconnectBtn.disabled = false;
    };

    ws.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            if (data.type === "leaderboard_snapshot") {
                handleSnapshot(data);
                ui.eventMeta.textContent = `Loaded ${data.leaderboard.length} rows, waiting for live deltas...`;
                return;
            }
            if (data.type !== "score_update") {
                return;
            }

            const playerId = String(data.player_id || "").trim();
            const playerName = String(data.player_name || "").trim();
            const score = Number(data.score);
            if (!playerId || Number.isNaN(score)) {
                return;
            }

            state.players.set(playerId, { playerId, playerName, score });
            resolvePlayerNameIfMissing(playerId, playerName);
            state.flashUntil.set(playerId, Date.now() + 1200);
            renderLeaderboard();
            scheduleFlashCleanup(playerId);
            const displayName = playerName || playerId;
            ui.eventMeta.textContent = `Latest: ${displayName} => ${score} at ${new Date().toLocaleTimeString()}`;
        } catch {
            ui.eventMeta.textContent = "Received malformed message";
        }
    };

    ws.onerror = () => {
        setStatus("WebSocket error", false);
    };

    ws.onclose = () => {
        if (state.ws === ws) {
            state.ws = null;
            ui.connectBtn.disabled = false;
            ui.disconnectBtn.disabled = true;
            setStatus("Disconnected", false);
        }
    };
}

async function loadInitialLeaderboard(apiBase, gameId) {
    const url = `${apiBase}/api/v1/games/${encodeURIComponent(gameId)}/scores?page=1&page_size=50&sort=score&order=desc`;

    try {
        const res = await fetch(url);
        if (!res.ok) {
            ui.eventMeta.textContent = `Initial fetch skipped (${res.status})`;
            return;
        }

        const payload = await res.json();
        const items = Array.isArray(payload.items) ? payload.items : [];

        for (const item of items) {
            const playerId = String(item.player_id || "").trim();
            const playerName = String(item.player_name || "").trim();
            const score = Number(item.score);
            if (!playerId || Number.isNaN(score)) {
                continue;
            }
            state.players.set(playerId, { playerId, playerName, score });
            resolvePlayerNameIfMissing(playerId, playerName);
        }

        renderLeaderboard();
        ui.eventMeta.textContent = `Loaded ${items.length} rows, waiting for live deltas...`;
    } catch {
        ui.eventMeta.textContent = "Initial fetch failed, waiting for live deltas...";
    }
}

function handleSnapshot(snapshot) {
    state.players.clear();
    state.flashUntil.clear();

    console.log("Received leaderboard snapshot with", snapshot);
    if (Array.isArray(snapshot.leaderboard)) {
        for (const entry of snapshot.leaderboard) {
            const playerId = String(entry.player_id || "").trim();
            const playerName = String(entry.player_name || "").trim();
            const score = Number(entry.score);
            if (!playerId || Number.isNaN(score)) {
                continue;
            }
            state.players.set(playerId, { playerId, playerName, score });
            resolvePlayerNameIfMissing(playerId, playerName);
        }
    }

    renderLeaderboard();
}

function resolvePlayerNameIfMissing(playerId, playerName) {
    if (playerName) {
        state.playerNameCache.set(playerId, playerName);
        return;
    }
    if (state.playerNameCache.has(playerId) || state.pendingNameRequests.has(playerId)) {
        const cached = state.playerNameCache.get(playerId);
        if (cached) {
            const current = state.players.get(playerId);
            if (current && !current.playerName) {
                state.players.set(playerId, { ...current, playerName: cached });
            }
        }
        return;
    }

    state.pendingNameRequests.add(playerId);
    const url = `${state.apiBase}/api/v1/players/${encodeURIComponent(playerId)}`;

    fetch(url)
        .then((res) => (res.ok ? res.json() : null))
        .then((data) => {
            const resolvedName = String(data?.username || "").trim();
            if (!resolvedName) {
                return;
            }
            state.playerNameCache.set(playerId, resolvedName);
            const current = state.players.get(playerId);
            if (current) {
                state.players.set(playerId, { ...current, playerName: resolvedName });
                renderLeaderboard();
            }
        })
        .finally(() => {
            state.pendingNameRequests.delete(playerId);
        });
}

function renderLeaderboard() {
    const now = Date.now();
    const rows = [...state.players.values()]
        .sort((a, b) => b.score - a.score || a.playerId.localeCompare(b.playerId))
        .slice(0, 50);

    if (rows.length === 0) {
        ui.leaderboardBody.innerHTML = `<tr><td colspan="3">No data yet</td></tr>`;
        return;
    }

    ui.leaderboardBody.innerHTML = rows
        .map((row, i) => {
            const isFlashing = (state.flashUntil.get(row.playerId) || 0) > now;
            const displayName = row.playerName || row.playerId;
            return `
      <tr>
        <td class="rank ${isFlashing ? "flash" : ""}">#${i + 1}</td>
        <td class="player">${escapeHtml(displayName)}</td>
        <td class="score ${isFlashing ? "flash" : ""}">${formatScore(row.score)}</td>
      </tr>
    `;
        })
        .join("");
}

function scheduleFlashCleanup(playerId) {
    setTimeout(() => {
        const expiresAt = state.flashUntil.get(playerId) || 0;
        if (expiresAt <= Date.now()) {
            state.flashUntil.delete(playerId);
            renderLeaderboard();
        }
    }, 1300);
}

function normalizeBase(value) {
    return value.replace(/\/$/, "") || "http://localhost:8080";
}

function toWsBase(apiBase) {
    if (apiBase.startsWith("https://")) {
        return `wss://${apiBase.slice("https://".length)}`;
    }
    if (apiBase.startsWith("http://")) {
        return `ws://${apiBase.slice("http://".length)}`;
    }
    return `ws://${apiBase}`;
}

function formatScore(n) {
    return Number(n).toLocaleString();
}

function setStatus(msg, online) {
    ui.status.textContent = msg;
    ui.status.classList.toggle("online", online);
    ui.status.classList.toggle("offline", !online);
}

function escapeHtml(value) {
    return value
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#39;");
}

function isMongoObjectId(value) {
    return /^[a-fA-F0-9]{24}$/.test(value);
}