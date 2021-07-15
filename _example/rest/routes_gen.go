package rest

import (
	dcdfbaac "example/users/commands"
	efebecad "example/users/entities"
	ddfedaff "example/users/queries"
	cabadefc "example/users/readmodels"
	"github.com/GabrielCarpr/cqrs/bus"
	cqrsErrs "github.com/GabrielCarpr/cqrs/errors"
	adapter "github.com/GabrielCarpr/cqrs/ports/rest"
	"net/http"

	"github.com/gin-gonic/gin"
)

func New(b *bus.Bus) *adapter.Server {
	s := adapter.NewServer(b)
	s.Map("POST", "/rest/v1/auth/login", func(b *bus.Bus) gin.HandlerFunc {
		return func(c *gin.Context) {
			query := ddfedaff.Login{}
			result := cabadefc.Authentication{}
			if err := adapter.MustBind(c, &query); err != nil {
				return
			}

			err := b.Query(c.Request.Context(), query, &result)
			if err == nil {
				c.JSON(http.StatusOK, result)
			}
			panic(err)
			switch err := err.(type) {
			case cqrsErrs.Error:
				c.JSON(err.Code, err)
				return
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
				return
			}
		}
	})
	s.Map("POST", "/rest/v1/auth/register", func(b *bus.Bus) gin.HandlerFunc {
		return func(c *gin.Context) {
			cmd := dcdfbaac.Register{}
			if err := adapter.MustBind(c, &cmd); err != nil {
				return
			}

			res, err := b.Dispatch(c.Request.Context(), cmd, true)
			if err != nil {
				if err, ok := err.(cqrsErrs.Error); ok {
					c.JSON(err.Code, err)
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
				return
			}

			c.JSON(http.StatusOK, res)
		}
	})
	s.Map("GET", "/rest/v1/users/:email", func(b *bus.Bus) gin.HandlerFunc {
		return func(c *gin.Context) {
			query := ddfedaff.User{}
			result := efebecad.User{}
			if err := adapter.MustBind(c, &query); err != nil {
				return
			}

			err := b.Query(c.Request.Context(), query, &result)
			if err == nil {
				c.JSON(http.StatusOK, result)
			}
			switch err := err.(type) {
			case cqrsErrs.Error:
				c.JSON(err.Code, err)
				return
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
				return
			}
		}
	})

	return s
}
