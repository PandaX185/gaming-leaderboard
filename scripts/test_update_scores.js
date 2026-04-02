import http from 'k6/http';
import { check } from 'k6';

export const options = {
    scenarios: {
        update_scores: {
            executor: 'ramping-arrival-rate',
            startRate: 5,
            timeUnit: '1s',
            preAllocatedVUs: 500,
            maxVUs: 1000,
            stages: [
                { target: 1000, duration: '30s' },
                { target: 5000, duration: '60s' },
                { target: 10000, duration: '90s' },
            ],
            exec: 'updateScore',
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
        const res = http.post(`http://host.docker.internal:8080/api/v1/games`, payload, params);
        if (res.status === 201) {
            gameIds.push(res.json().id);
        }
    }

    for (let i = 0; i < 100; i++) {
        const payload = JSON.stringify({
            username: `player_${Date.now()}_${i}`,
            password: 'password123',
        });
        const res = http.post(`http://host.docker.internal:8080/api/v1/players`, payload, params);
        if (res.status === 201) {
            playerIds.push(res.json().id);
        }
    }
    return { gameIds, playerIds };
}

export function updateScore(data) {
    const { gameIds, playerIds } = data;
    if (!gameIds || !playerIds || gameIds.length === 0 || playerIds.length === 0) {
        console.error("Setup data missing!");
        return;
    }
    const player = playerIds[Math.floor(Math.random() * playerIds.length)];
    const game = gameIds[Math.floor(Math.random() * gameIds.length)];
    const scorePayload = JSON.stringify({
        game_id: game,
        score: Math.floor(Math.random() * 1000) + 1,
    });
    const params = { headers: { 'Content-Type': 'application/json' } };
    const scoreRes = http.put(`http://host.docker.internal:8080/api/v1/players/${player}/score`, scorePayload, params);
    check(scoreRes, { '200 OK': (r) => r.status === 200 });
}
