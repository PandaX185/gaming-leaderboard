package handler

import (
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc *service.UserService
	rg  *gin.RouterGroup
}

func NewUserHandler(svc *service.UserService, rg *gin.RouterGroup) *UserHandler {
	return &UserHandler{
		svc: svc,
		rg:  rg,
	}
}

func (h *UserHandler) RegisterRoutes() {
	h.rg.POST("/users", h.CreateUser)
	h.rg.PUT("/users/:id/score", h.UpdateUserScore)
	h.rg.GET("/users/:id", h.GetUserByID)
	h.rg.GET("/users", h.GetAllUsers)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	data := &dto.CreateUserRequest{}
	if err := c.ShouldBindJSON(data); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	if err := dto.ValidateStructRequest(data); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.CreateUser(c.Request.Context(), data)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create user "+err.Error()})
		return
	}

	c.JSON(201, resp)
}

func (h *UserHandler) UpdateUserScore(c *gin.Context) {
	id := c.Param("id")
	data := &dto.UpdateScoreRequest{}
	if err := c.ShouldBindJSON(data); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	if err := dto.ValidateStructRequest(data); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.UpdateUserScore(c.Request.Context(), id, data.Score)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update user score "+err.Error()})
		return
	}

	c.JSON(200, resp)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.svc.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "User not found "+err.Error()})
		return
	}

	c.JSON(200, resp)
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	params := &dto.PaginationParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		c.JSON(400, gin.H{"error": "Invalid query parameters"})
		return
	}

	if err := dto.ValidateStructRequest(params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.GetAllUsers(c.Request.Context(), params)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve users "+err.Error()})
		return
	}

	c.JSON(200, resp)
}
