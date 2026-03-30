package handler

import (
	"gaming-leaderboard/internal/realtime"

	"github.com/gin-gonic/gin"
)

type LeaderboardWSHandler struct {
	hub *realtime.LeaderboardHub
	rg  *gin.RouterGroup
}

func NewLeaderboardWSHandler(hub *realtime.LeaderboardHub, rg *gin.RouterGroup) *LeaderboardWSHandler {
	return &LeaderboardWSHandler{
		hub: hub,
		rg:  rg,
	}
}

func (h *LeaderboardWSHandler) RegisterRoutes() {
	h.rg.GET("/games/:id/leaderboard/ws", h.StreamGameLeaderboard)
}

func (h *LeaderboardWSHandler) StreamGameLeaderboard(c *gin.Context) {
	h.hub.HandleGameWS(c)
}
