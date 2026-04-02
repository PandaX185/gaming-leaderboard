package handler

import (
	"fmt"

	"gaming-leaderboard/internal/log"

	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, err error, contextMsg string) {
	if err == nil {
		return
	}

	if contextMsg != "" {
		err = fmt.Errorf("%s: %w", contextMsg, err)
	}

	log.Error("Request error: %v", err)
	c.Error(err)
}
