package handler

import (
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/service"

	"github.com/gin-gonic/gin"
)

type GameHandler struct {
	svc *service.GameService
	rg  *gin.RouterGroup
}

func NewGameHandler(svc *service.GameService, rg *gin.RouterGroup) *GameHandler {
	return &GameHandler{
		svc: svc,
		rg:  rg,
	}
}

func (h *GameHandler) RegisterRoutes() {
	h.rg.POST("/games", h.CreateGame)
	h.rg.GET("/games/:id", h.GetGameByID)
	h.rg.GET("/games/:id/scores", h.GetGameScores)
	h.rg.GET("/games", h.GetAllGames)
	h.rg.PUT("/games/:id", h.UpdateGame)
	h.rg.DELETE("/games/:id", h.DeleteGame)
}

func (h *GameHandler) CreateGame(c *gin.Context) {
	data := &dto.CreateGameRequest{}
	if err := c.ShouldBindJSON(data); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	if err := dto.ValidateStructRequest(data); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.CreateGame(c.Request.Context(), data)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create game " + err.Error()})
		return
	}

	c.JSON(201, resp)
}

func (h *GameHandler) GetGameByID(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.svc.GetGameByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Game not found " + err.Error()})
		return
	}

	c.JSON(200, resp)
}

func (h *GameHandler) GetGameScores(c *gin.Context) {
	id := c.Param("id")
	params := &dto.PaginationParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		c.JSON(400, gin.H{"error": "Invalid query parameters"})
		return
	}

	if err := dto.ValidateStructRequest(params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.GetGameScores(c.Request.Context(), id, params)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve game scores " + err.Error()})
		return
	}

	c.JSON(200, resp)
}

func (h *GameHandler) GetAllGames(c *gin.Context) {
	params := &dto.PaginationParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		c.JSON(400, gin.H{"error": "Invalid query parameters"})
		return
	}

	if err := dto.ValidateStructRequest(params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.GetAllGames(c.Request.Context(), params)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve games " + err.Error()})
		return
	}

	c.JSON(200, resp)
}

func (h *GameHandler) UpdateGame(c *gin.Context) {
	id := c.Param("id")
	data := &dto.UpdateGameRequest{}
	if err := c.ShouldBindJSON(data); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	if err := dto.ValidateStructRequest(data); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.UpdateGame(c.Request.Context(), id, data)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update game " + err.Error()})
		return
	}

	c.JSON(200, resp)
}

func (h *GameHandler) DeleteGame(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteGame(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete game " + err.Error()})
		return
	}

	c.JSON(204, nil)
}
