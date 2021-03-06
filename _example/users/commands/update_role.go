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

type UpdateRole struct {
	bus.CommandType

	ID     entities.RoleID `json:"ID"`
	Label  *string         `json:"label"`
	Scopes *[]string       `json:"scopes"`
}

func (c UpdateRole) Valid() error {
	switch true {
	case len(c.ID.String()) == 0:
		return errs.ValidationError("Role ID is missing")
	case util.Empty(c.Label) && util.Empty(c.Scopes):
		return errs.ValidationError("Label and scopes cannot be nil")
	}
	return nil
}

func (UpdateRole) Auth(context.Context) [][]string {
	return [][]string{[]string{"roles:write"}}
}

func (c UpdateRole) Command() string {
	return "users.update-role"
}

func NewUpdateRoleHandler(r db.RoleRepository) *UpdateRoleHandler {
	return &UpdateRoleHandler{r}
}

var _ bus.CommandHandler = (*UpdateRoleHandler)(nil)

type UpdateRoleHandler struct {
	roles db.RoleRepository
}

func (h UpdateRoleHandler) Execute(ctx context.Context, c bus.Command) (res bus.CommandResponse, msgs []message.Message) {
	cmd := c.(UpdateRole)

	var role entities.Role
	r, err := h.roles.Find(cmd.ID)
	switch true {
	case len(r) != 1:
		res = bus.CommandResponse{Error: errs.EntityNotFound}
		return
	case err != nil:
		res = bus.CommandResponse{Error: err}
		return
	default:
		role = r[0]
	}

	if cmd.Label != nil {
		role = role.ChangeLabel(*cmd.Label)
	}
	if cmd.Scopes != nil {
		role = role.DropScopes().ApplyScopes(*cmd.Scopes...)
	}

	events := role.Messages(ctx)
	role.Commit()
	err = h.roles.Persist(role)
	switch err {
	case nil:
		res.ID = role.ID.String()
		msgs = events
		return
	default:
		res.Error = err
		return
	}
}
