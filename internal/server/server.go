package server

import (
	"gaming-leaderboard/internal/log"
	"gaming-leaderboard/metrics"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
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
	log.Info("NewServer initializing on port %s", port)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.Default())
	r.Use(PrometheusMiddleware())
	r.Use(ErrorHandlerMiddleware())
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return &Server{
		Port: port,
		Srv:  r,
	}
}

func NewServerWithOpts(port string, opts []gin.OptionFunc) *Server {
	log.Info("NewServerWithOpts initializing on port %s", port)
	return &Server{
		Port: port,
		Srv:  gin.New(opts...),
	}
}

func (s *Server) Run() {
	log.Info("Server running on port %s", s.Port)
	s.Srv.Run(s.Port)
}
