package rest

import (
    "github.com/GabrielCarpr/cqrs/bus"
    adapter "github.com/GabrielCarpr/cqrs/ports/rest"
    cqrsErrs "github.com/GabrielCarpr/cqrs/errors"
    "net/http"
    cbedaaff "example/internal/support"
    dcdfbaac "example/users/commands"
    efebecad "example/users/entities"
    ddfedaff "example/users/queries"

    "github.com/gin-gonic/gin"
)

func New(b *bus.Bus, config adapter.Config) *adapter.Server {
    server := adapter.NewServer(b, config)
    server.Map("POST", "/rest/v1/auth/register", func (b *bus.Bus) gin.HandlerFunc {
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
    server.Map("GET", "/rest/v1/users/:ID", func (b *bus.Bus) gin.HandlerFunc {
        return func(c *gin.Context) {
            query := ddfedaff.User{}
            result := efebecad.User{}
            if err := adapter.MustBind(c, &query); err != nil {
                return
            }

            err := b.Query(c.Request.Context(), query, &result)
            if err == nil {
                c.JSON(http.StatusOK, result)
                return
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
    },server.Auth())
    server.Map("GET", "/rest/v1/roles/:ID", func (b *bus.Bus) gin.HandlerFunc {
        return func(c *gin.Context) {
            query := ddfedaff.Role{}
            result := efebecad.Role{}
            if err := adapter.MustBind(c, &query); err != nil {
                return
            }

            err := b.Query(c.Request.Context(), query, &result)
            if err == nil {
                c.JSON(http.StatusOK, result)
                return
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
    },server.Auth())
    server.Map("GET", "/rest/v1/users", func (b *bus.Bus) gin.HandlerFunc {
        return func(c *gin.Context) {
            query := ddfedaff.Users{}
            result := cbedaaff.PaginatedQuery{}
            if err := adapter.MustBind(c, &query); err != nil {
                return
            }

            err := b.Query(c.Request.Context(), query, &result)
            if err == nil {
                c.JSON(http.StatusOK, result)
                return
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
    },server.Auth())
    server.Map("GET", "/rest/v1/roles", func (b *bus.Bus) gin.HandlerFunc {
        return func(c *gin.Context) {
            query := ddfedaff.Roles{}
            result := cbedaaff.PaginatedQuery{}
            if err := adapter.MustBind(c, &query); err != nil {
                return
            }

            err := b.Query(c.Request.Context(), query, &result)
            if err == nil {
                c.JSON(http.StatusOK, result)
                return
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
    },server.Auth())
    server.Map("POST", "/rest/v1/roles", func (b *bus.Bus) gin.HandlerFunc {
        return func(c *gin.Context) {
            cmd := dcdfbaac.CreateRole{}
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
    },server.Auth())
    server.Map("PUT", "/rest/v1/users", func (b *bus.Bus) gin.HandlerFunc {
        return func(c *gin.Context) {
            cmd := dcdfbaac.UpdateUser{}
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
    },server.Auth())
    server.Map("PUT", "/rest/v1/roles", func (b *bus.Bus) gin.HandlerFunc {
        return func(c *gin.Context) {
            cmd := dcdfbaac.UpdateRole{}
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

    return server
}
