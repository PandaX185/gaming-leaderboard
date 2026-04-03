import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
        normal_flow: {
            executor: 'constant-vus',
            vus: 15,
            duration: '5m',
            exec: 'normalFlowScenario',
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<500'],
        http_req_failed: ['rate<0.05'],
    }
};

export function setup() {
    const gameIds = [];
    const playerIds = [];
    const params = { headers: { 'Content-Type': 'application/json' } };

    for (let i = 0; i < 10; i++) {
        const payload = JSON.stringify({ name: `Game_${Date.now()}_${i}` });
        const res = http.post(`http://localhost:8080/api/v1/games`, payload, params);
        if (res.status === 201) {
            gameIds.push(res.json().id);
        }
    }

    for (let i = 0; i < 100; i++) {
        const payload = JSON.stringify({
            username: `player_${Date.now()}_${i}`,
            password: 'password123',
        });
        const res = http.post(`http://localhost:8080/api/v1/players`, payload, params);
        if (res.status === 201) {
            playerIds.push(res.json().id);
        }
    }

    const selectedGame = gameIds[0];
    console.log(`Selected test game: ${selectedGame}`);
    console.log("Populating initial scores...");
    for (let i = 0; i < playerIds.length; i++) {
        const player = playerIds[i];
        const scorePayload = JSON.stringify({
            game_id: selectedGame,
            score: Math.floor(Math.random() * 10000) + 1,
        });
        http.put(`http://localhost:8080/api/v1/players/${player}/score`, scorePayload, params);
    }
    console.log("Setup complete - scores populated");
    return { selectedGame, gameIds, playerIds };
}

export function normalFlowScenario(data) {
    const { selectedGame, playerIds } = data;
    if (!selectedGame || !playerIds || playerIds.length === 0) {
        console.error("Setup data missing!");
        return;
    }
    const game = selectedGame;
    const leaderboardRes = http.get(`http://localhost:8080/api/v1/games/${game}/scores?page=1&page_size=20&sort=score&order=desc`);
    if (leaderboardRes.status !== 200) {
        console.error(`Failed to fetch leaderboard for game ${game}: ${leaderboardRes.status}`);
        return;
    }
    const leaderboardData = leaderboardRes.json();
    const topPlayers = Array.isArray(leaderboardData.items) ? leaderboardData.items : [];
    if (topPlayers.length === 0) {
        console.warn(`No players found in leaderboard for game ${game}`);
        return;
    }
    const selectedPlayerData = topPlayers[Math.floor(Math.random() * topPlayers.length)];
    const player = selectedPlayerData.player_id;
    const scorePayload = JSON.stringify({
        game_id: game,
        score: Math.floor(Math.random() * 5000) + 1,
    });
    const params = { headers: { 'Content-Type': 'application/json' } };
    const scoreRes = http.put(`http://localhost:8080/api/v1/players/${player}/score`, scorePayload, params);
    check(scoreRes, {
        'score update 200 OK': (r) => r.status === 200,
        'score update not failed': (r) => r.status < 400,
    });
    const randomDelay = Math.floor(Math.random() * 7000) + 3000;
    sleep(randomDelay / 1000);
}
