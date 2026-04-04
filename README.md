# 🎮 Gaming Leaderboard

A real-time gaming leaderboard system built with Go, MongoDB, and Redis. Track player scores across multiple games with live WebSocket updates, a fully containerised stack, and built-in observability through Prometheus and Grafana.

---

## Features

- **Player & Game Management** – full CRUD operations with pagination
- **Asynchronous Score Updates** – queue-backed worker pool keeps the API non-blocking
- **Real-time Leaderboards** – WebSocket broadcast delivers live rank changes to all connected clients
- **Redis Caching** – leaderboard snapshots served from cache for low latency
- **Observability** – Prometheus metrics, Grafana dashboards, pprof profiling
- **Load Testing** – k6 script included for performance validation

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.26, Gin |
| Database | MongoDB (replica set) |
| Cache / Queue | Redis |
| Real-time | WebSocket (gorilla/websocket) |
| Monitoring | Prometheus + Grafana |
| Frontend | Vanilla JS + HTML/CSS |
| Containerisation | Docker + Docker Compose + Traefik |
| Load Testing | k6 |

---

## Architecture

```
HTTP Request
    │
    ▼
Handler ──► Service ──► Redis Queue
                              │
                         Worker Pool  (async, default 100 workers)
                              │
                         Repository ──► MongoDB / Redis
                              │
                         Redis Pub/Sub
                              │
                         WebSocket Hub ──► Connected Clients
```

- **Repository pattern** – separates data access from business logic  
- **Service layer** – orchestrates use-cases and validation  
- **Queue / Worker** – decouples score ingestion from persistence  
- **Hub / Spoke** – Redis pub/sub fans out updates to WebSocket clients across instances  

---

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/)

### Quickstart (recommended)

```bash
git clone https://github.com/PandaX185/gaming-leaderboard.git
cd gaming-leaderboard
docker-compose up --build
```

Traefik will expose the backend API on port `8080` and proxy requests to the Go app service.

| Service | URL |
|---|---|
| API | http://localhost:8080 |
| Frontend | http://localhost:3001 |
| Prometheus | http://localhost:9090 |
| Grafana | http://localhost:3000 (admin / admin) |
| Mongo Express | http://localhost:8081 (admin / password) |

---

### Local Development

**Prerequisites:** Go 1.26+, MongoDB, Redis

1. **Install dependencies**

   ```bash
   go mod download
   ```

2. **Create a `.env` file** in the project root

   ```env
   DB_URI=mongodb://localhost:27017
   DB_NAME=gaming_leaderboard
   REDIS_URI=redis://localhost:6379
   PORT=:8080
   QUEUE_TYPE=redis
   WORKER_COUNT=100
   ```

3. **Run the backend**

   ```bash
   go run main.go
   ```

4. **Run the frontend**

   ```bash
   cd frontend
   python3 -m http.server 3001
   # or: npm install && npm start
   ```

---

## API Reference

Base path: `/api/v1`

### Players

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/players` | Create a player |
| `GET` | `/players` | List players (paginated) |
| `GET` | `/players/:id` | Get player by ID |
| `PUT` | `/players/:id/score` | Submit a score update (async) |

### Games

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/games` | Create a game |
| `GET` | `/games` | List games (paginated) |
| `GET` | `/games/:id` | Get game details |
| `PUT` | `/games/:id` | Update game |
| `DELETE` | `/games/:id` | Delete game |
| `GET` | `/games/:id/scores` | Get leaderboard |

**Leaderboard query parameters:** `page`, `page_size`, `sort` (`score`\|`name`), `order` (`asc`\|`desc`)

### WebSocket

```
GET /games/:id/leaderboard/ws
```

**Message types received from server:**

```jsonc
// Initial snapshot on connect
{ "type": "leaderboard_snapshot", "leaderboard": [ ... ] }

// Live update when any player's score changes
{ "type": "score_update", "player_id": "...", "score": 1500, "rank": 3 }
```

### Monitoring

```
GET /metrics    Prometheus metrics endpoint
```

---

### Example Requests

```bash
# Create a player
curl -X POST http://localhost:8080/api/v1/players \
  -H "Content-Type: application/json" \
  -d '{"username":"player1","password":"secret123"}'

# Update a player's score
curl -X PUT http://localhost:8080/api/v1/players/{playerId}/score \
  -H "Content-Type: application/json" \
  -d '{"game_id":"{gameId}","score":1500}'

# Fetch leaderboard (top 20, sorted by score)
curl "http://localhost:8080/api/v1/games/{gameId}/scores?page=1&page_size=20&sort=score&order=desc"
```

---

## Load Testing

The repository ships with a [k6](https://k6.io/) load test that:

- Spins up 10 games and 100 players
- Runs 15 virtual users for 5 minutes
- Continuously fetches leaderboards and submits score updates

**Thresholds:** p95 latency < 500 ms, error rate < 5 %

```bash
k6 run scripts/test.js
```

---

## Monitoring

Once the stack is running via Docker Compose, Grafana is available at **http://localhost:3000** (credentials: `admin` / `admin`).

Pre-provisioned dashboards cover:

- API request rates and latency
- Worker pool throughput and queue depth
- MongoDB and Redis performance
