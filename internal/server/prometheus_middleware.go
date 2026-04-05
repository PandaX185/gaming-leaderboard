package server

import (
	"gaming-leaderboard/metrics"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}
		metrics.RequestsInFlight.WithLabelValues(c.Request.Method, endpoint).Inc()
		defer metrics.RequestsInFlight.WithLabelValues(c.Request.Method, endpoint).Dec()

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		metrics.RequestsTotal.WithLabelValues(c.Request.Method, endpoint, status).Inc()
		metrics.RequestDuration.WithLabelValues(c.Request.Method, endpoint).Observe(duration)
	}
}
