package rest

import (
	"context"

	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/gin-gonic/gin"
)

func NewServer(b *bus.Bus) *Server {
	s := &Server{b, gin.Default()}
	return s
}

type Server struct {
	bus    *bus.Bus
	router *gin.Engine
}

func (s *Server) Map(method string, route string, handler func(*bus.Bus) gin.HandlerFunc) {
	s.router.Handle(method, route, handler(s.bus))
}

func (s *Server) Run(ctx context.Context) error {
	return s.router.Run()
}
