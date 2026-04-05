package handler

import (
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/errors"
	"gaming-leaderboard/internal/service"
	"strconv"

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
		HandleError(c, errors.NewBadRequest("Invalid request", err), "")
		c.Abort()
		return
	}
	if err := dto.ValidateStructRequest(data); err != nil {
		HandleError(c, errors.NewBadRequest(err.Error(), err), "")
		c.Abort()
		return
	}

	resp, err := h.svc.CreateGame(c.Request.Context(), data)
	if err != nil {
		HandleError(c, err, "failed to create game")
		c.Abort()
		return
	}

	c.JSON(201, resp)
}

func (h *GameHandler) GetGameByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		HandleError(c, errors.NewBadRequest("Invalid ID", err), "")
		c.Abort()
		return
	}
	resp, err := h.svc.GetGameByID(c.Request.Context(), id)
	if err != nil {
		HandleError(c, err, "game not found")
		c.Abort()
		return
	}

	c.JSON(200, resp)
}

func (h *GameHandler) GetGameScores(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		HandleError(c, errors.NewBadRequest("Invalid ID", err), "")
		c.Abort()
		return
	}
	params := &dto.PaginationParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		HandleError(c, errors.NewBadRequest("Invalid query parameters", err), "")
		c.Abort()
		return
	}
	params.ClampAndDefault()

	resp, err := h.svc.GetGameScores(c.Request.Context(), id, params)
	if err != nil {
		HandleError(c, err, "failed to retrieve game scores")
		c.Abort()
		return
	}

	c.JSON(200, resp)
}

func (h *GameHandler) GetAllGames(c *gin.Context) {
	params := &dto.PaginationParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		HandleError(c, errors.NewBadRequest("Invalid query parameters", err), "")
		c.Abort()
		return
	}
	params.ClampAndDefault()

	resp, err := h.svc.GetAllGames(c.Request.Context(), params)
	if err != nil {
		HandleError(c, err, "failed to retrieve games")
		c.Abort()
		return
	}

	c.JSON(200, resp)
}

func (h *GameHandler) UpdateGame(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		HandleError(c, errors.NewBadRequest("Invalid ID", err), "")
		c.Abort()
		return
	}
	data := &dto.UpdateGameRequest{}
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

	resp, err := h.svc.UpdateGame(c.Request.Context(), id, data)
	if err != nil {
		HandleError(c, err, "failed to update game")
		c.Abort()
		return
	}

	c.JSON(200, resp)
}

func (h *GameHandler) DeleteGame(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		HandleError(c, errors.NewBadRequest("Invalid ID", err), "")
		c.Abort()
		return
	}
	err = h.svc.DeleteGame(c.Request.Context(), id)
	if err != nil {
		HandleError(c, err, "failed to delete game")
		c.Abort()
		return
	}

	c.Status(204)
}
