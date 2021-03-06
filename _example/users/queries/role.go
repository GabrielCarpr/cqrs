package queries

import (
	"example/internal/errs"
	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/errors"
	"example/users/db"
	"example/users/entities"
	"context"
)

type Role struct {
	bus.QueryType

	ID entities.RoleID
}

func (q Role) Valid() error {
	switch true {
	case len(q.ID.String()) == 0:
		return errs.ValidationError("ID must not be empty")
	}
	return nil
}

func (Role) Auth(context.Context) [][]string {
	return [][]string{[]string{"roles:read"}}
}

func (Role) Query() string {
	return "role"
}

func NewRoleHandler(r db.RoleRepository) *RoleHandler {
	return &RoleHandler{r}
}

var _ bus.QueryHandler = (*RoleHandler)(nil)

type RoleHandler struct {
	roles db.RoleRepository
}

func (h RoleHandler) Execute(ctx context.Context, q bus.Query, r interface{}) error {
	query := q.(Role)
	res := r.(*entities.Role)

	roles, err := h.roles.Find(query.ID)
	switch true {
	case len(roles) != 1:
		return errors.Error{Code: 404, Message: "Role not found"}
	case err != nil:
		return err
	default:
		*res = roles[0]
		return nil
	}
}
