package queries_test

import (
	"example/internal/tester"
	"github.com/GabrielCarpr/cqrs/auth"
	"github.com/GabrielCarpr/cqrs/errors"
	"example/users/db"
	"example/users/entities"
	"example/users/queries"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestRole(t *testing.T) {
	suite.Run(t, new(TestRoleSuite))
}

type TestRoleSuite struct {
	suite.Suite
	tester.Integration

	roles db.RoleRepository
}

func (s *TestRoleSuite) SetupTest() {
	s.Integration.SetupTest()
	s.roles = s.Get("role-repository").(db.RoleRepository)

}

func (s TestRoleSuite) TestGetAdminRole() {
	query := queries.Role{ID: entities.NewRoleID("admin")}
	var r entities.Role
	err := s.Bus().Query(auth.TestCtx(uuid.New(), "roles:read"), query, &r)

	s.NoError(err)
	s.Equal(entities.NewRoleID("admin"), r.ID)
	s.Equal("Admin", r.Label)
}

func (s TestRoleSuite) TestGetMissingRole() {
	query := queries.Role{ID: entities.NewRoleID("king")}
	var r entities.Role
	err := s.Bus().Query(auth.TestCtx(uuid.New(), "roles:read"), query, &r)

	s.Error(err)
	role := r
	s.Empty(role.ID.String())
	s.IsType(errors.Error{}, err)
}
