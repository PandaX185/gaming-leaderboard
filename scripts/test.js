import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
        create_players: {
            executor: 'ramping-arrival-rate',
            startRate: 10,
            timeUnit: '1s',
            preAllocatedVUs: 50,
            maxVUs: 500,
            stages: [
                { target: 100, duration: '30s' },
                { target: 500, duration: '1m' },
                { target: 1000, duration: '2m' },
            ],
            exec: 'createPlayer',
        },
        update_scores: {
            executor: 'ramping-arrival-rate',
            startRate: 5,
            timeUnit: '1s',
            preAllocatedVUs: 50,
            maxVUs: 500,
            stages: [
                { target: 50, duration: '30s' },
                { target: 200, duration: '1m' },
                { target: 500, duration: '2m' },
            ],
            exec: 'updateScore',
        },
    },
};

const gameIds = [
    '69c0e8a9382a53995762aacc',
    '69c0e21d01a336170cd3a5bd',
    '69c0e22901a336170cd3a5be',
];

export function createPlayer() {
    const payload = JSON.stringify({
        username: `user_${Date.now()}_${Math.random().toString(36).substring(7)}`,
        password: 'password123',
    });

    const params = { headers: { 'Content-Type': 'application/json' } };
    const res = http.post(`http://host.docker.internal:8080/api/v1/players`, payload, params);

    check(res, { '201 Created': (r) => r.status === 201 });

    sleep(0.1);
}

export function updateScore() {
    const playersRes = http.get(`http://host.docker.internal:8080/api/v1/players?page=1&page_size=40`);
    const players = playersRes.json().items;
    if (players.length === 0) return;

    const player = players[Math.floor(Math.random() * players.length)];

    const scorePayload = JSON.stringify({
        game_id: gameIds[Math.floor(Math.random() * gameIds.length)],
        score: Math.floor(Math.random() * 1000),
    });

    const params = { headers: { 'Content-Type': 'application/json' } };
    const scoreRes = http.put(`http://host.docker.internal:8080/api/v1/players/${player.id}/score`, scorePayload, params);

    check(scoreRes, { '200 OK': (r) => r.status === 200 });

    sleep(0.1);
}