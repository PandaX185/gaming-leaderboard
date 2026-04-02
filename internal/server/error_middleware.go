package server

import (
	"gaming-leaderboard/internal/errors"
	"gaming-leaderboard/internal/log"

	"github.com/gin-gonic/gin"
)

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		if err == nil {
			return
		}

		status, message := errors.HTTPError(err)
		log.LogHTTPError(c, err, status)

		if c.Writer.Written() {
			return
		}

		c.AbortWithStatusJSON(status, gin.H{"error": message})
	}
}
