package rest

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/GabrielCarpr/cqrs/auth"
	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/log"
	"github.com/gin-gonic/gin"
)

type Config struct {
	Secret      string
	URL         string
	Development bool
}

func NewServer(b *bus.Bus, conf Config) *Server {
	s := &Server{b, gin.Default(), conf}
	return s
}

type Server struct {
	bus    *bus.Bus
	Router *gin.Engine
	Config Config
}

func (s *Server) Map(method string, route string, handler func(*bus.Bus) gin.HandlerFunc, middlewares ...gin.HandlerFunc) {
	handlers := make([]gin.HandlerFunc, len(middlewares)+1)
	for i, mw := range middlewares {
		handlers[i] = mw
	}
	handlers[len(handlers)-1] = handler(s.bus)
	s.Router.Handle(method, route, handlers...)
}

func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:    ":80",
		Handler: s.Router,
	}

	errord := make(chan struct{})
	var err error
	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			close(errord)
		}
	}()

	select {
	case <-errord:
		return err
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		return srv.Shutdown(ctx)
	}
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

		credentials, err := auth.ReadToken(parts[1], s.Config.Secret)
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
