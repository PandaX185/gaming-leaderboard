package realtime

import (
	"context"
	"gaming-leaderboard/internal/log"
	"gaming-leaderboard/internal/repository"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type ScoreUpdate struct {
	Type     string  `json:"type"`
	PlayerID string  `json:"player_id"`
	Score    float64 `json:"score"`
	Rank     int64   `json:"rank"`
}

type LeaderboardEntry struct {
	Rank     int64   `json:"rank"`
	PlayerID string  `json:"player_id"`
	Score    float64 `json:"score"`
}

type LeaderboardSnapshot struct {
	Type        string             `json:"type"`
	GameID      string             `json:"game_id"`
	Leaderboard []LeaderboardEntry `json:"leaderboard"`
	Timestamp   int64              `json:"timestamp"`
}

type wsClient struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (c *wsClient) writeJSON(payload any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return c.conn.WriteJSON(payload)
}

type LeaderboardHub struct {
	rdb *redis.Client

	upgrader websocket.Upgrader

	mu      sync.Mutex
	clients map[string]map[*wsClient]struct{}
	stops   map[string]context.CancelFunc
}

func NewLeaderboardHub(rdb *redis.Client) *LeaderboardHub {
	return &LeaderboardHub{
		rdb: rdb,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
		clients: make(map[string]map[*wsClient]struct{}),
		stops:   make(map[string]context.CancelFunc),
	}
}

func (h *LeaderboardHub) HandleGameWS(c *gin.Context) {
	gameIDStr := c.Param("id")
	if gameIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game id is required"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &wsClient{conn: conn}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	snapshot, snapErr := h.buildLeaderboardSnapshot(ctx, gameIDStr)
	cancel()

	if snapErr == nil {
		_ = client.writeJSON(snapshot)
	}

	h.addClient(gameIDStr, client)
	defer h.removeClient(gameIDStr, client)

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })

	processDone := make(chan struct{})
	defer close(processDone)
	go func() {
		ticker := time.NewTicker(54 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				client.mu.Lock()
				conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					client.mu.Unlock()
					return
				}
				client.mu.Unlock()
			case <-processDone:
				return
			}
		}
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *LeaderboardHub) addClient(gameID string, client *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clientsByGame, exists := h.clients[gameID]
	if !exists {
		clientsByGame = make(map[*wsClient]struct{})
		h.clients[gameID] = clientsByGame
	}

	clientsByGame[client] = struct{}{}

	if len(clientsByGame) == 1 {
		stream := repository.LeaderboardUpdatesStream(gameID)
		ctx, cancel := context.WithCancel(context.Background())
		h.stops[gameID] = cancel
		go h.consumeGameUpdates(ctx, gameID, stream)
	}
}

func (h *LeaderboardHub) removeClient(gameID string, client *wsClient) {
	client.conn.Close()

	h.mu.Lock()
	defer h.mu.Unlock()

	clientsByGame, exists := h.clients[gameID]
	if !exists {
		return
	}

	delete(clientsByGame, client)
	if len(clientsByGame) > 0 {
		return
	}

	delete(h.clients, gameID)

	if stop := h.stops[gameID]; stop != nil {
		stop()
		delete(h.stops, gameID)
	}
}

func (h *LeaderboardHub) consumeGameUpdates(ctx context.Context, gameIDStr string, stream string) {
	lastID := "$"
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		resp, err := h.rdb.XRead(ctx, &redis.XReadArgs{
			Streams: []string{stream, lastID},
			Count:   50,
			Block:   5 * time.Second,
		}).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Error("failed reading leaderboard updates stream for game %s: %v", gameIDStr, err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		for _, chunk := range resp {
			for _, msg := range chunk.Messages {
				lastID = msg.ID
				h.broadcastDeltaToGame(gameIDStr, msg.Values)
			}
		}
	}
}

func (h *LeaderboardHub) broadcastDeltaToGame(gameID string, values map[string]any) {
	clients := h.getClients(gameID)
	if len(clients) == 0 {
		return
	}

	update, ok := toScoreUpdate(values)
	if !ok {
		log.Warn("received malformed leaderboard delta for game %s: %#v", gameID, values)
		return
	}

	for _, client := range clients {
		if err := client.writeJSON(update); err != nil {
			h.removeClient(gameID, client)
		}
	}
}

func (h *LeaderboardHub) buildLeaderboardSnapshot(ctx context.Context, gameID string) (LeaderboardSnapshot, error) {
	key := repository.LeaderboardKey(gameID)
	rows, err := h.rdb.ZRevRangeWithScores(ctx, key, 0, 99).Result()
	if err != nil {
		return LeaderboardSnapshot{}, err
	}

	entries := make([]LeaderboardEntry, 0, len(rows))
	for i, row := range rows {
		playerIDStr, ok := row.Member.(string)
		if !ok {
			continue
		}

		entries = append(entries, LeaderboardEntry{
			Rank:     int64(i + 1),
			PlayerID: playerIDStr,
			Score:    row.Score,
		})
	}

	return LeaderboardSnapshot{
		Type:        "leaderboard_snapshot",
		GameID:      gameID,
		Leaderboard: entries,
		Timestamp:   time.Now().Unix(),
	}, nil
}

func (h *LeaderboardHub) getClients(gameID string) []*wsClient {
	h.mu.Lock()
	defer h.mu.Unlock()

	clientsByGame, exists := h.clients[gameID]
	if !exists {
		return nil
	}

	clients := make([]*wsClient, 0, len(clientsByGame))
	for client := range clientsByGame {
		clients = append(clients, client)
	}
	return clients
}

func toScoreUpdate(values map[string]any) (ScoreUpdate, bool) {
	playerIDStr, ok := asString(values["player_id"])
	if !ok {
		return ScoreUpdate{}, false
	}

	score, ok := asFloat64(values["score"])
	if !ok {
		return ScoreUpdate{}, false
	}

	rank, ok := asInt64(values["rank"])
	if !ok {
		return ScoreUpdate{}, false
	}

	updateType := "score_update"
	if rawType, hasType := values["type"]; hasType {
		if parsedType, ok := asString(rawType); ok && parsedType != "" {
			updateType = parsedType
		}
	}

	return ScoreUpdate{
		Type:     updateType,
		PlayerID: playerIDStr,
		Score:    score,
		Rank:     rank,
	}, true
}

func asString(v any) (string, bool) {
	switch t := v.(type) {
	case string:
		return t, true
	case []byte:
		return string(t), true
	default:
		return "", false
	}
}

func asFloat64(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case uint64:
		return float64(t), true
	case string:
		f, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	case []byte:
		f, err := strconv.ParseFloat(string(t), 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

func asInt64(v any) (int64, bool) {
	switch t := v.(type) {
	case int64:
		return t, true
	case int:
		return int64(t), true
	case uint64:
		return int64(t), true
	case float64:
		return int64(t), true
	case string:
		i, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return 0, false
		}
		return i, true
	case []byte:
		i, err := strconv.ParseInt(string(t), 10, 64)
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}
