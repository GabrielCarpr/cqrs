package queries

import (
	"example/internal/errs"
	"example/internal/support"
	"github.com/GabrielCarpr/cqrs/auth"
	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/errors"
	"example/pkg/util"
	"example/users/db"
	"example/users/entities"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

var (
	ErrUserNotFound = errors.Error{Code: 404, Message: "User not found"}
)

type User struct {
	bus.QueryType

	Email *string     `json:"email"`
	ID    *support.ID `json:"id" cqrs:"id"`
}

func (User) Query() string {
	return "user"
}

func (q User) Valid() error {
	if util.Empty(q.Email) && util.Empty(q.ID) {
		return errs.ValidationError("Must specify Email or ID")
	}
	return nil
}

func (q User) Auth(ctx context.Context) [][]string {
	if !util.Empty(q.ID) {
		return [][]string{[]string{"users:read"}, []string{"self:read", auth.UserScope(q.ID.UUID)}}
	}
	return [][]string{[]string{"users:read"}}
}

var _ bus.QueryHandler = (*UserHandler)(nil)

func NewUserHandler(u db.UserRepository, db *sqlx.DB) *UserHandler {
	return &UserHandler{u, db}
}

type UserHandler struct {
	users db.UserRepository
	db    *sqlx.DB
}

func (h UserHandler) Execute(ctx context.Context, q bus.Query, r interface{}) error {
	query := q.(User)
	res := r.(*entities.User)

	if !util.Empty(query.ID) {
		return h.getById(*query.ID, res)
	}
	return h.getByEmail(*query.Email, res)
}

func (h UserHandler) getById(id support.ID, res *entities.User) error {
	users, err := h.users.Find(id)
	switch true {
	case len(users) == 0:
		return ErrUserNotFound
	case err != nil:
		return err
	}

	*res = users[0]
	return err
}

func (h UserHandler) getByEmail(email string, res *entities.User) error {
	var id support.ID
	err := h.db.Get(&id, "SELECT ID FROM users WHERE email = $1", email)
	switch err {
	case nil:
		return h.getById(id, res)
	case sql.ErrNoRows:
		return ErrUserNotFound
	default:
		return support.TransduceError(err)
	}
}
