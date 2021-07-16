package rest

import (
    "github.com/GabrielCarpr/cqrs/bus"
    "github.com/GabrielCarpr/cqrs/ports/rest"
    "example/internal/config"
    "example/users/queries"
    "example/users/readmodels"
    "github.com/gin-gonic/gin"
    "net/http"
    cqrsErrs "github.com/GabrielCarpr/cqrs/errors"
)

//go:generate go run github.com/GabrielCarpr/cqrs/gen gen rest routes.yml

func Rest(b *bus.Bus, config config.Config) *rest.Server {
    server := New(b, rest.Config{
        Secret: config.Secret,
        URL: config.AppURL,
        Development: config.Environment == "development",
    })

    server.Map("POST", "/rest/v1/auth/login", func (b *bus.Bus) gin.HandlerFunc {
        return func(c *gin.Context) {
            query := queries.Login{}
            result := readmodels.Authentication{}
            if err := rest.MustBind(c, &query); err != nil {
                return
            }

            err := b.Query(c.Request.Context(), query, &result)
            if err == nil {
                c.SetCookie("refresh", result.RefreshToken, 604800, "/", server.Config.URL, !server.Config.Development, true)
                c.SetSameSite(http.SameSiteStrictMode)
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
    })

    return server
}
