package server

import "github.com/gin-gonic/gin"

type Server struct {
	Port string `yaml:"port"`
	Srv  *gin.Engine
}

func NewServer(port string) *Server {
	return &Server{
		Port: port,
		Srv:  gin.Default(),
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
