package server

import (
	"gaming-leaderboard/internal/log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	Port string `yaml:"port"`
	Srv  *gin.Engine
}

func NewServer(port string) *Server {
	log.Info("NewServer initializing on port %s", port)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.Default())
	r.Use(TimeoutMiddleware())
	r.Use(PrometheusMiddleware())
	r.Use(ErrorHandlerMiddleware())
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return &Server{
		Port: port,
		Srv:  r,
	}
}

func NewServerWithOpts(port string, opts []gin.OptionFunc) *Server {
	log.Info("NewServerWithOpts initializing on port %s", port)
	r := gin.New(opts...)
	r.Use(TimeoutMiddleware())
	return &Server{
		Port: port,
		Srv:  r,
	}
}

func (s *Server) Run() {
	log.Info("Server running on port %s", s.Port)
	s.Srv.Run(s.Port)
}
