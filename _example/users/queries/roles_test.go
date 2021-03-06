package queries_test

import (
	"example/internal/tester"
	"github.com/GabrielCarpr/cqrs/auth"
	"example/users/entities"
	"example/users/queries"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type GetRolesTest struct {
	suite.Suite
	tester.Integration
}

func TestRoles(t *testing.T) {
	suite.Run(t, new(GetRolesTest))
}

func (s GetRolesTest) GetRoles() {
	bus := s.Bus()

	q := queries.Roles{IDs: &[]string{"admin", "user"}}
	s.NoError(q.Valid())
	var res []entities.Role
	err := bus.Query(auth.TestCtx(uuid.New(), "roles:read"), q, &res)

	s.NoError(err)
	s.Len(res, 2)
	s.Equal(res[0].ID, entities.NewRoleID("admin"))
	s.Equal(res[0].Label, "Admin")
}
