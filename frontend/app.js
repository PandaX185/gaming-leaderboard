const state = {
    ws: null,
    apiBase: "http://localhost:8080",
    gameId: "",
    players: new Map(),
    flashUntil: new Map(),
    ranks: new Map(),
    rankTrends: new Map(), // playerId -> 'up' | 'down'
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
    ui.gameId.value = gameId;
}

function connect() {
    const gameId = ui.gameId.value.trim();
    if (!gameId) {
        setStatus("Game ID is required", false);
        return;
    }


    disconnect();

    const apiBase = normalizeBase(ui.apiBaseUrl.value.trim());

    state.gameId = gameId;
    state.apiBase = apiBase;
    state.players.clear();
    state.flashUntil.clear();
    state.ranks.clear();
    state.rankTrends.clear();
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
            const score = Number(data.score);
            if (!playerId || Number.isNaN(score)) {
                return;
            }

            state.players.set(playerId, { playerId, score });
            state.flashUntil.set(playerId, Date.now() + 1200);
            renderLeaderboard();
            scheduleFlashCleanup(playerId);
            const displayName = playerId;
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
            const score = Number(item.score);
            if (!playerId || Number.isNaN(score)) {
                continue;
            }
            state.players.set(playerId, { playerId, score });
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
    state.ranks.clear();
    state.rankTrends.clear();

    console.log("Received leaderboard snapshot with", snapshot);
    if (Array.isArray(snapshot.leaderboard)) {
        for (const entry of snapshot.leaderboard) {
            const playerId = String(entry.player_id || "").trim();
            const score = Number(entry.score);
            if (!playerId || Number.isNaN(score)) {
                continue;
            }
            state.players.set(playerId, { playerId, score });
        }
    }

    renderLeaderboard();
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

    const newRanks = new Map();
    const rankChanges = new Map();

    rows.forEach((row, index) => {
        newRanks.set(row.playerId, index);
        const oldRank = state.ranks.get(row.playerId);

        if (oldRank !== undefined) {
            if (oldRank > index) {
                state.rankTrends.set(row.playerId, { dir: 'up', expires: now + 3000 });
                scheduleTrendCleanup(row.playerId);
            } else if (oldRank < index) {
                state.rankTrends.set(row.playerId, { dir: 'down', expires: now + 3000 });
                scheduleTrendCleanup(row.playerId);
            }
        }
    });
    state.ranks = newRanks;

    ui.leaderboardBody.innerHTML = rows
        .map((row, i) => {
            const isFlashing = (state.flashUntil.get(row.playerId) || 0) > now;

            let trend = '';
            let trendClass = '';
            const trendData = state.rankTrends.get(row.playerId);
            if (trendData && trendData.expires > now) {
                trendClass = trendData.dir === 'up' ? 'trend-up' : 'trend-down';
                trend = trendData.dir === 'up' ? ' ▲' : ' ▼';
            } else if (trendData) {
                state.rankTrends.delete(row.playerId);
            }

            const displayName = row.playerId;
            return `
      <tr class="${trendClass}">
        <td class="rank ${isFlashing ? "flash" : ""}">#${i + 1}<span class="trend-icon">${trend}</span></td>
        <td class="player">${escapeHtml(displayName)}</td>
        <td class="score ${isFlashing ? "flash" : ""}">${formatScore(row.score)}</td>
      </tr>
    `;
        })
        .join("");
}

function scheduleTrendCleanup(playerId) {
    setTimeout(() => {
        const data = state.rankTrends.get(playerId);
        if (data && data.expires <= Date.now()) {
            state.rankTrends.delete(playerId);
            renderLeaderboard();
        }
    }, 3100);
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

