package rest

import (
	"context"
	"strings"

	"github.com/GabrielCarpr/cqrs/auth"
	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/log"
	"github.com/gin-gonic/gin"
)

func NewServer(b *bus.Bus, secret string) *Server {
	s := &Server{b, gin.Default(), secret}
	return s
}

type Server struct {
	bus    *bus.Bus
	router *gin.Engine
	secret string
}

func (s *Server) Map(method string, route string, handler func(*bus.Bus) gin.HandlerFunc, middlewares ...gin.HandlerFunc) {
	handlers := make([]gin.HandlerFunc, len(middlewares)+1)
	for i, mw := range middlewares {
		handlers[i] = mw
	}
	handlers[len(handlers)-1] = handler(s.bus)
	s.router.Handle(method, route, handlers...)
}

func (s *Server) Run(ctx context.Context) error {
	return s.router.Run()
}

func (s *Server) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		if authorization == "" {
			ctx := auth.WithCredentials(c.Request.Context(), auth.BlankCredentials)
			c.Request = c.Request.WithContext(ctx)
			c.Next()
			return
		}

		parts := strings.Split(authorization, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(400, gin.H{"message": "Malformed Authorization header", "code": 400})
			return
		}

		credentials, err := auth.ReadToken(parts[1], s.secret)
		if err != nil {
			log.Error(c.Request.Context(), "JWT token invalid", log.F{"error": err.Error()})
			c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized", "code": 401})
			return
		}

		ctx := auth.WithCredentials(c.Request.Context(), credentials)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
