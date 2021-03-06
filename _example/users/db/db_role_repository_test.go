package db_test

import (
	"example/internal/tester"
	"example/users/db"
	"example/users/entities"
	"github.com/stretchr/testify/suite"
	"testing"
)

type RoleRepositoryTest struct {
	suite.Suite
	tester.Integration

	roles db.RoleRepository
}

func TestRoleRepository(t *testing.T) {
	suite.Run(t, new(RoleRepositoryTest))
}

func (s *RoleRepositoryTest) SetupTest() {
	s.Integration.SetupTest()

	s.roles = s.Get("role-repository").(db.RoleRepository)
}

func (s *RoleRepositoryTest) TestPersistRoles() {
	r := entities.CreateRole("User")

	err := s.roles.Persist(r)
	s.NoError(err)

	roles, err := s.roles.Find(r.ID)
	s.NoError(err)
	s.Len(roles, 1)
	role := roles[0]

	s.Equal(role.ID, r.ID)
	s.Len(role.Scopes(), 3)
}

func (s *RoleRepositoryTest) TestFindMissingRoles() {
	roles, err := s.roles.Find(entities.NewRoleID("king"))
	s.NoError(err)
	s.Len(roles, 0)
}
