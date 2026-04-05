import http from 'k6/http';
import { check } from 'k6';

export const options = {
    scenarios: {
        constant_rps: {
            executor: 'constant-arrival-rate',
            rate: 7000,
            timeUnit: '1s',
            duration: '1m',
            preAllocatedVUs: 700,
            maxVUs: 1300,
        },
    },

    thresholds: {
        http_req_duration: ['p(95)<500'],
        http_req_failed: ['rate<0.05'],
    },
};

export function setup() {
    const gameIds = [];
    const playerIds = [];
    const params = { headers: { 'Content-Type': 'application/json' } };

    for (let i = 0; i < 10; i++) {
        const res = http.post(`http://localhost/api/v1/games`, JSON.stringify({
            name: `Game_${Date.now()}_${i}`,
        }), params);

        if (res.status === 201) {
            gameIds.push(res.json().id);
        }
    }

    for (let i = 0; i < 100; i++) {
        const res = http.post(`http://localhost/api/v1/players`, JSON.stringify({
            username: `player_${Date.now()}_${i}`,
            password: 'password123',
        }), params);

        if (res.status === 201) {
            playerIds.push(res.json().id);
        }
    }

    return { gameIds, playerIds };
}

export default function (data) {
    const { gameIds, playerIds } = data;

    if (!gameIds.length || !playerIds.length) {
        console.error("Missing setup data");
        return;
    }

    const player = playerIds[Math.floor(Math.random() * playerIds.length)];
    const game = gameIds[Math.floor(Math.random() * gameIds.length)];

    const payload = JSON.stringify({
        player_id: player,
        game_id: game,
        score: Math.floor(Math.random() * 1000) + 1,
    });

    const res = http.put(`http://localhost/api/v1/scores`, payload, {
        headers: { 'Content-Type': 'application/json' },
    });

    check(res, {
        'status is 200': (r) => r.status === 200,
    });
}