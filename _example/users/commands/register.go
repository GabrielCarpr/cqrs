package commands

import (
	"example/internal/errs"
	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/bus/message"
	"github.com/GabrielCarpr/cqrs/errors"
	"example/pkg/util"
	"example/users/db"
	"example/users/entities"
	"context"
)

var (
	ErrUserExists = errors.Error{Code: 400, Message: "User exists"}
)

// Register registers a user on the system
type Register struct {
	bus.CommandType

	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Admin    bool   `json:"-"`
}

func (c Register) Valid() error {
	switch true {
	case util.Empty(c.Email):
		return errs.ValidationError("Email must be present")
	case util.Empty(c.Password):
		return errs.ValidationError("Password must be present")
	case len(c.Password) < 7:
		return errs.ValidationError("Password too short")
	}
	return nil
}

func (c Register) Command() string {
	return "users.register"
}

func (c Register) WithEmail(e string) Register {
	c.Email = e
	return c
}

func (c Register) WithPassword(p string) Register {
	c.Password = p
	return c
}

func (c Register) WithName(n string) Register {
	c.Name = n
	return c
}

func (c Register) ByAdmin(a bool) Register {
	c.Admin = a
	return c
}

func NewRegisterHandler(u db.UserRepository, r db.RoleRepository) *RegisterHandler {
	return &RegisterHandler{u, r}
}

var _ bus.CommandHandler = (*RegisterHandler)(nil)

type RegisterHandler struct {
	users db.UserRepository
	roles db.RoleRepository
}

func (h RegisterHandler) Execute(ctx context.Context, c bus.Command) (res bus.CommandResponse, msgs []message.Message) {
	cmd := c.(Register)
	userRoles, err := h.roles.Find(entities.NewRoleID("user"))
	if err != nil {
		res = bus.CommandResponse{Error: err}
		return
	}
	if len(userRoles) == 0 {
		res = bus.CommandResponse{
			Error: errs.EntityNotFound,
		}
	}
	userRole := userRoles[0]

	user, err := entities.Register(cmd.Email, cmd.Password)
	if err != nil {
		res = bus.CommandResponse{Error: err}
		return
	}
	user = user.ChangeName(cmd.Name)
	user = user.Grant(userRole)

	events := user.Messages(ctx)
	user.Commit()
	err = h.users.Persist(user)
	switch err {
	case nil:
		res = bus.CommandResponse{ID: user.ID.String()}
		msgs = append(msgs, events...)
		return
	case errs.UniqueEntityExists:
		res = bus.CommandResponse{Error: ErrUserExists}
		return
	default:
		res = bus.CommandResponse{Error: err}
		return
	}
}
