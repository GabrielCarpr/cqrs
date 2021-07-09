package commands

import (
	"example/internal/errs"
	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/bus/message"
	"example/pkg/util"
	"example/users/db"
	"example/users/entities"
	"context"
)

type CreateRole struct {
	bus.CommandType

	Name   string
	Scopes []string
}

func (c CreateRole) Valid() error {
	switch true {
	case util.Empty(c.Name):
		return errs.ValidationError("Role name cannot be empty")
	}
	return nil
}

func (CreateRole) Auth(context.Context) [][]string {
	return [][]string{[]string{"roles:write"}}
}

func (c CreateRole) Command() string {
	return "users.create-role"
}

func NewCreateRoleHandler(r db.RoleRepository) *CreateRoleHandler {
	return &CreateRoleHandler{r}
}

var _ bus.CommandHandler = (*CreateRoleHandler)(nil)

type CreateRoleHandler struct {
	roles db.RoleRepository
}

func (h *CreateRoleHandler) Execute(ctx context.Context, c bus.Command) (res bus.CommandResponse, msgs []message.Message) {
	cmd := c.(CreateRole)

	role := entities.CreateRole(cmd.Name)
	role = role.ApplyScopes(cmd.Scopes...)

	err := h.roles.Persist(role)
	switch err {
	case nil:
		res = bus.CommandResponse{ID: role.ID.String()}
		msgs = append(msgs, role.Release()...)
		return
	default:
		res = bus.CommandResponse{Error: err}
		return
	}
}
