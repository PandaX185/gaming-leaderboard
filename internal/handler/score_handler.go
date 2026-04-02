package handler

import (
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/service"

	"github.com/gin-gonic/gin"
)

type ScoreHandler struct {
	svc *service.ScoreService
	rg  *gin.RouterGroup
}

func NewScoreHandler(svc *service.ScoreService, rg *gin.RouterGroup) *ScoreHandler {
	return &ScoreHandler{
		svc: svc,
		rg:  rg,
	}
}

func (h *ScoreHandler) RegisterRoutes() {
	h.rg.PUT("/scores", h.UpdateScore)
}

func (h *ScoreHandler) UpdateScore(c *gin.Context) {
	data := &dto.UpdateScoreRequest{}
	if err := c.ShouldBindJSON(data); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	if err := dto.ValidateStructRequest(data); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.UpdateScore(c.Request.Context(), data.PlayerID, data.GameID, data.Score)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update score " + err.Error()})
		return
	}

	c.JSON(200, resp)
}
