package handler

import (
	"gaming-leaderboard/internal/dto"
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
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	if err := dto.ValidateStructRequest(data); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.CreatePlayer(c.Request.Context(), data)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create player " + err.Error()})
		return
	}

	c.JSON(201, resp)
}

func (h *PlayerHandler) GetPlayerByID(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.svc.GetPlayerByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Player not found " + err.Error()})
		return
	}

	c.JSON(200, resp)
}

func (h *PlayerHandler) GetAllPlayers(c *gin.Context) {
	params := &dto.PaginationParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		c.JSON(400, gin.H{"error": "Invalid query parameters"})
		return
	}

	if err := dto.ValidateStructRequest(params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.GetAllPlayers(c.Request.Context(), params)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve players " + err.Error()})
		return
	}

	c.JSON(200, resp)
}
