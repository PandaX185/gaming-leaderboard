package server

import (
	"gaming-leaderboard/metrics"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	Port string `yaml:"port"`
	Srv  *gin.Engine
}

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

func NewServer(port string) *Server {
	r := gin.Default()
	r.Use(PrometheusMiddleware())
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return &Server{
		Port: port,
		Srv:  r,
	}
}

func NewServerWithOpts(port string, opts []gin.OptionFunc) *Server {
	return &Server{
		Port: port,
		Srv:  gin.New(opts...),
	}
}

func (s *Server) Run() {
	s.Srv.Run(s.Port)
}
