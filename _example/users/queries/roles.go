package queries

import (
	"example/internal/errs"
	"github.com/GabrielCarpr/cqrs/bus"
	"example/pkg/util"
	"example/users/db"
	"example/users/entities"
	"context"
)

type Roles struct {
	bus.QueryType

	IDs *[]string `json:"IDs"`
}

func (q Roles) Valid() error {
	switch true {
	case !util.Empty(q.IDs) && len(*q.IDs) == 0:
		return errs.ValidationError("Role IDs must be longer than 0 IDs")
	}
	return nil
}

func (Roles) Auth(context.Context) [][]string {
	return [][]string{[]string{"roles:read"}}
}

func (Roles) Query() string {
	return "query"
}

var _ bus.QueryHandler = (*RolesHandler)(nil)

func NewRolesHandler(r db.RoleRepository) *RolesHandler {
	return &RolesHandler{r}
}

type RolesHandler struct {
	roles db.RoleRepository
}

func (h RolesHandler) Execute(ctx context.Context, q bus.Query, r interface{}) error {
	query := q.(Roles)
	res := r.(*[]entities.Role)

	if !util.Empty(query.IDs) {
		return h.getByIds(*query.IDs, res)
	}

	return h.getAll(res)
}

func (h RolesHandler) getAll(res *[]entities.Role) error {
	roles, err := h.roles.All()
	switch true {
	case err == nil:
		*res = roles
		return nil
	default:
		return err
	}
}

func (h RolesHandler) getByIds(ids []string, res *[]entities.Role) error {
	roleIDs := make([]entities.RoleID, len(ids))
	for i, id := range ids {
		roleIDs[i] = entities.NewRoleID(id)
	}
	roles, err := h.roles.Find(roleIDs...)
	switch true {
	case err == nil:
		*res = roles
		return nil
	default:
		return err
	}
}
