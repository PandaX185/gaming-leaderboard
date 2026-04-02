package handler

import (
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/errors"
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
		HandleError(c, errors.NewBadRequest("Invalid request", err), "")
		c.Abort()
		return
	}
	if err := dto.ValidateStructRequest(data); err != nil {
		HandleError(c, errors.NewBadRequest(err.Error(), err), "")
		c.Abort()
		return
	}

	resp, err := h.svc.UpdateScore(c.Request.Context(), data.PlayerID, data.GameID, data.Score)
	if err != nil {
		HandleError(c, err, "failed to update score")
		c.Abort()
		return
	}

	c.JSON(200, resp)
}
