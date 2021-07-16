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
    grp := server.Router.Group("")

    
func(grp gin.IRouter) {
grp = grp.Group("/rest/v1")

    
func(grp gin.IRouter) {
grp = grp.Group("/auth")

grp.Handle("POST", "/register", func(c *gin.Context) {
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
})}(grp)

    
func(grp gin.IRouter) {
grp = grp.Group("")
grp.Use(server.Auth(),)
    
func(grp gin.IRouter) {
grp = grp.Group("/users")

grp.Handle("GET", "/:ID", func (c *gin.Context) {
    query := ddfedaff.User{}
    result := efebecad.User{}
    if err := adapter.MustBind(c, &query); err != nil {
        return
    }
    err := b.Query(c.Request.Context(), query, &result)
    if err == nil {
        c.JSON(http.StatusOK,result)
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
})
grp.Handle("GET", "/", func (c *gin.Context) {
    query := ddfedaff.Users{}
    result := cbedaaff.PaginatedQuery{}
    if err := adapter.MustBind(c, &query); err != nil {
        return
    }
    err := b.Query(c.Request.Context(), query, &result)
    if err == nil {
        c.JSON(http.StatusOK,result)
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
})
grp.Handle("PUT", "/:ID", func(c *gin.Context) {
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
})}(grp)

    
func(grp gin.IRouter) {
grp = grp.Group("/roles")

grp.Handle("GET", "/:ID", func (c *gin.Context) {
    query := ddfedaff.Role{}
    result := efebecad.Role{}
    if err := adapter.MustBind(c, &query); err != nil {
        return
    }
    err := b.Query(c.Request.Context(), query, &result)
    if err == nil {
        c.JSON(http.StatusOK,roleAdapter{result})
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
})
grp.Handle("GET", "/", func (c *gin.Context) {
    query := ddfedaff.Roles{}
    result := cbedaaff.PaginatedQuery{}
    if err := adapter.MustBind(c, &query); err != nil {
        return
    }
    err := b.Query(c.Request.Context(), query, &result)
    if err == nil {
        c.JSON(http.StatusOK,rolesAdapter{result})
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
})
grp.Handle("PUT", "/:ID", func(c *gin.Context) {
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
})
grp.Handle("POST", "/", func(c *gin.Context) {
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
})}(grp)
}(grp)
}(grp)


    return server
}
