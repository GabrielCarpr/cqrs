package commands

import (
	"example/internal/errs"
	"example/internal/support"
	"github.com/GabrielCarpr/cqrs/auth"
	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/bus/message"
	"github.com/GabrielCarpr/cqrs/errors"
	"example/pkg/util"
	"example/users/db"
	"example/users/entities"
	"context"

	"github.com/thoas/go-funk"
)

type UpdateUser struct {
	bus.CommandType

	ID       support.ID `json:"ID"`
	Name     *string    `json:"name"`
	Email    *string    `json:"email"`
	Roles    *[]string  `json:"roles"`
	Password *string    `json:"password"`
}

func (c UpdateUser) Valid() error {
	switch true {
	case c.ID.Nil():
		return errs.ValidationError("ID must be provided")
	case !util.Empty(c.Password) && len(*c.Password) < 8:
		return errs.ValidationError("Password too short")
	}
	return nil
}

func (c UpdateUser) Auth(ctx context.Context) [][]string {
	return [][]string{[]string{"users:write"}, []string{"self:write", auth.UserScope(c.ID.UUID)}}
}

func (c UpdateUser) Command() string {
	return "users.update-user"
}

var _ bus.CommandHandler = (*UpdateUserHandler)(nil)

func NewUpdateUserHandler(u db.UserRepository, r db.RoleRepository) *UpdateUserHandler {
	return &UpdateUserHandler{u, r}
}

type UpdateUserHandler struct {
	users db.UserRepository
	roles db.RoleRepository
}

func (h UpdateUserHandler) Execute(ctx context.Context, c bus.Command) (res bus.CommandResponse, _ []message.Message) {
	cmd := c.(UpdateUser)

	user, err := h.getUser(cmd.ID)
	if err != nil {
		res = bus.CommandResponse{Error: err}
		return
	}

	user, err = h.changeDetails(user, cmd)
	if err != nil {
		res = bus.CommandResponse{Error: err}
		return
	}
	user, err = h.updateRoles(ctx, user, cmd)
	if err != nil {
		res = bus.CommandResponse{Error: err}
		return
	}

	err = h.users.Persist(user)
	switch err {
	case nil:
		res = bus.CommandResponse{ID: user.ID.String()}
		return
	default:
		res = bus.CommandResponse{Error: err}
		return
	}
}

func (h *UpdateUserHandler) getUser(ID support.ID) (entities.User, error) {
	u, err := h.users.Find(ID)
	switch true {
	case len(u) != 1:
		return entities.User{}, errs.EntityNotFound
	default:
		return u[0], err
	}
}

func (h *UpdateUserHandler) changeDetails(user entities.User, cmd UpdateUser) (entities.User, error) {
	var err error
	if !util.Empty(cmd.Name) {
		user = user.ChangeName(*cmd.Name)
	}
	if !util.Empty(cmd.Email) {
		user, err := user.ChangeEmail(*cmd.Email)
		if err != nil {
			return user, err
		}
	}
	if !util.Empty(cmd.Password) {
		user, err = user.ChangePassword(*cmd.Password)
		if err != nil {
			return user, err
		}
	}

	return user, nil
}

func (h *UpdateUserHandler) updateRoles(ctx context.Context, user entities.User, cmd UpdateUser) (entities.User, error) {
	if cmd.Roles == nil {
		return user, nil
	}
	if err := auth.Enforce(ctx, []string{"users:write"}); err != nil {
		return user, err
	}

	roleIDs := funk.Map(*cmd.Roles, func(role string) entities.RoleID {
		return entities.NewRoleID(role)
	}).([]entities.RoleID)

	roles, err := h.roles.Find(roleIDs...)
	switch true {
	case len(roles) != len(roleIDs):
		return user, errors.Error{Code: 400, Message: "Role not found"}
	case err != nil:
		return user, err
	}

	user = user.Revoke()
	for _, role := range roles {
		user = user.Grant(role)
	}

	return user, nil
}
