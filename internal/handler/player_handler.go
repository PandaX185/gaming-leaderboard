package handler

import (
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/errors"
	"gaming-leaderboard/internal/service"

	"github.com/gin-gonic/gin"
)

type PlayerHandler struct {
	svc *service.PlayerService
	rg  *gin.RouterGroup
}

func NewPlayerHandler(svc *service.PlayerService, rg *gin.RouterGroup) *PlayerHandler {
	return &PlayerHandler{
		svc: svc,
		rg:  rg,
	}
}

func (h *PlayerHandler) RegisterRoutes() {
	h.rg.POST("/players", h.CreatePlayer)
	h.rg.GET("/players/:id", h.GetPlayerByID)
	h.rg.GET("/players", h.GetAllPlayers)
}

func (h *PlayerHandler) CreatePlayer(c *gin.Context) {
	data := &dto.CreatePlayerRequest{}
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

	resp, err := h.svc.CreatePlayer(c.Request.Context(), data)
	if err != nil {
		HandleError(c, err, "failed to create player")
		c.Abort()
		return
	}

	c.JSON(201, resp)
}

func (h *PlayerHandler) GetPlayerByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		HandleError(c, errors.NewBadRequest("Invalid ID", nil), "")
		c.Abort()
		return
	}
	resp, err := h.svc.GetPlayerByID(c.Request.Context(), id)
	if err != nil {
		HandleError(c, err, "player not found")
		c.Abort()
		return
	}

	c.JSON(200, resp)
}

func (h *PlayerHandler) GetAllPlayers(c *gin.Context) {
	params := &dto.PaginationParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		HandleError(c, errors.NewBadRequest("Invalid query parameters", err), "")
		c.Abort()
		return
	}
	params.ClampAndDefault()

	resp, err := h.svc.GetAllPlayers(c.Request.Context(), params)
	if err != nil {
		HandleError(c, err, "failed to retrieve players")
		c.Abort()
		return
	}

	c.JSON(200, resp)
}
