package queries

import (
	"example/internal/config"
	"example/internal/errs"
	"example/internal/support"
	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/errors"
	"example/pkg/util"
	"example/users/db"
	"example/users/entities"
	"example/users/readmodels"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

var (
	ErrLoginFail = errors.Error{Code: 401, Message: "Email or password incorrect"}
)

type Login struct {
	bus.QueryType

	Email    string `json:"email"`
	Password string `json:"password"`
}

func (q Login) Valid() error {
	switch true {
	case util.Empty(q.Password):
		return errs.ValidationError("Password is missing")
	case util.Empty(q.Email):
		return errs.ValidationError("Email is empty")
	}
	return nil
}

func (Login) Query() string {
	return "login"
}

func NewLoginHandler(c *config.Config, u db.UserRepository, r db.RoleRepository, db *sqlx.DB) *LoginHandler {
	return &LoginHandler{c, u, r, db}
}

type LoginHandler struct {
	conf  *config.Config
	users db.UserRepository
	roles db.RoleRepository
	db    *sqlx.DB
}

var _ bus.QueryHandler = (*LoginHandler)(nil)

func (h LoginHandler) Execute(ctx context.Context, q bus.Query, r interface{}) error {
	query := q.(Login)
	res := r.(*readmodels.Authentication)
	var ID support.ID
	err := h.db.Get(&ID, "SELECT ID FROM users WHERE email = $1", query.Email)
	switch err {
	case sql.ErrNoRows:
		return ErrLoginFail
	case nil:
		break
	default:
		return err
	}

	var user entities.User
	u, err := h.users.Find(ID)
	switch true {
	case err != nil:
		return err
	case len(u) == 0:
		return ErrLoginFail
	default:
		user = u[0]
	}

	authed := user.CheckPassword(query.Password)
	if !authed {
		return ErrLoginFail
	}

	roles, err := h.roles.Find(user.RoleIDs...)
	if err != nil {
		return err
	}

	auth, err := readmodels.NewAuthentication(h.conf, user, roles...)
	*res = auth
	return err
}
