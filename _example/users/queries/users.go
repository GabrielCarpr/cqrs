package queries

import (
	"example/internal/errs"
	"example/internal/support"
	"github.com/GabrielCarpr/cqrs/bus"
	"example/pkg/util"
	"example/users/db"
	"example/users/entities"
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Users struct {
	bus.QueryType

	support.Paging

	IDs     *[]support.ID
	RoleIDs *[]entities.RoleID
}

func (Users) Query() string {
	return "users"
}

func (q Users) Valid() error {
	switch true {
	case !util.Empty(q.IDs) && len(*q.IDs) == 0:
		return errs.ValidationError("ID query length must be greater than 0")
	case !util.Empty(q.IDs) && len(*q.IDs) > 100:
		return errs.ValidationError("Can only query for 100 users at a time")
	case q.Paging.Valid() != nil:
		return q.Paging.Valid()
	}
	return nil
}

func (Users) Auth(context.Context) [][]string {
	return [][]string{[]string{"users:read"}}
}

var _ bus.QueryHandler = (*UsersHandler)(nil)

func NewUsersHandler(u db.UserRepository, db *sqlx.DB) *UsersHandler {
	return &UsersHandler{u, db}
}

type UsersHandler struct {
	users db.UserRepository
	db    *sqlx.DB
}

func (h UsersHandler) Execute(ctx context.Context, q bus.Query, r interface{}) error {
	query := q.(Users)
	res := r.(*support.PaginatedQuery)

	switch true {
	case !util.Empty(query.IDs):
		return h.getIds(*query.IDs, res)
	case !util.Empty(query.RoleIDs):
		return h.getByRoleIds(query, res)
	default:
		return h.getAll(query, res)
	}
}

func (h UsersHandler) getAll(q Users, res *support.PaginatedQuery) error {
	var userIDs []support.ID
	query := fmt.Sprintf(`
		SELECT ID FROM users
		ORDER BY %s %s OFFSET %v LIMIT %v
	`, q.Paging.OrderField(), q.Paging.OrderBy(), q.Paging.Offset(), q.Paging.Limit())

	err := h.db.Select(&userIDs, query)
	if err != nil {
		return support.TransduceError(err)
	}

	var count int
	err = h.db.Get(&count, "SELECT COUNT(*) FROM USERS")
	if err != nil {
		return support.TransduceError(err)
	}

	users, err := h.users.Find(userIDs...)
	switch true {
	case err == nil:
		*res = support.NewPaginatedQuery(users, count)
		return nil
	default:
		return support.TransduceError(err)
	}
}

func (h UsersHandler) getIds(ids []support.ID, res *support.PaginatedQuery) error {
	users, err := h.users.Find(ids...)
	switch true {
	case err == nil:
		*res = support.NewPaginatedQuery(users, len(ids))
		return nil
	default:
		return err
	}
}

func (h UsersHandler) getByRoleIds(q Users, res *support.PaginatedQuery) error {
	var userIDs []support.ID
	query, args, err := sqlx.In(fmt.Sprintf(`
		SELECT ID FROM users
		LEFT JOIN user_roles ON user_roles.user_id = users.ID
		WHERE user_roles.role_id IN(?)
		ORDER BY %s %s OFFSET %v LIMIT %v
	`, q.Paging.OrderField(), q.Paging.OrderBy(), q.Paging.Offset(), q.Paging.Limit()), *q.RoleIDs)
	if err != nil {
		return err
	}

	query = h.db.Rebind(query)
	err = h.db.Select(&userIDs, query, args...)
	if err != nil {
		return support.TransduceError(err)
	}

	var count int
	query, args, err = sqlx.In(`SELECT COUNT(*) FROM users
		LEFT JOIN user_roles ON user_roles.user_id = users.ID
		WHERE user_roles.role_id IN(?)`, *q.RoleIDs)
	if err != nil {
		return err
	}
	query = h.db.Rebind(query)
	err = h.db.Get(&count, query, args...)
	if err != nil {
		return support.TransduceError(err)
	}

	users, err := h.users.Find(userIDs...)
	switch true {
	case err == nil:
		*res = support.NewPaginatedQuery(users, count)
		return nil
	default:
		return support.TransduceError(err)
	}
}
