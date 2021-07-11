package db_test

import (
	"example/internal/support"
	"example/internal/tester"
	"example/users/db"
	"example/users/entities"
	"github.com/stretchr/testify/suite"
	"testing"
)

type UserRepositoryTest struct {
	suite.Suite
	tester.Integration

	users db.UserRepository
	roles db.RoleRepository
}

func TestUserRepository(t *testing.T) {
	suite.Run(t, new(UserRepositoryTest))
}

func (s *UserRepositoryTest) SetupTest() {
	s.Integration.SetupTest()

	s.users = s.Get("user-repository").(db.UserRepository)
	s.roles = s.Get("role-repository").(db.RoleRepository)
}

func (s *UserRepositoryTest) TestPersistUser() {
	u, _ := entities.Register("gc@gmail.com", "password123")
	r := entities.CreateRole("User")
	r2 := entities.CreateRole("Admin")
	u = u.Grant(r).Grant(r2)

	err := s.users.Persist(u)
	s.NoError(err)

	users, err := s.users.Find(u.ID)
	s.NoError(err)
	s.Len(users, 1)
	u = users[0]

	s.Equal("gc@gmail.com", u.Email.String())
	s.Len(u.RoleIDs, 2)
	s.Contains(u.RoleIDs, entities.NewRoleID("user"))
}

func (s *UserRepositoryTest) TestFindMissingUser() {
	users, err := s.users.Find(support.NewID())
	s.NoError(err)
	s.Len(users, 0)
}

func (s *UserRepositoryTest) TestFindMultipleUsers() {
	u1, _ := entities.Register("gc1@gmail.com", "password123")
	u2, _ := entities.Register("gc2@gmail.com", "password123")

	err := s.users.Persist(u1, u2)
	s.NoError(err)

	users, err := s.users.Find(u1.ID, u2.ID)
	s.Len(users, 2)
}

func (s *UserRepositoryTest) TestGrantUserRole() {
	u, _ := entities.Register("gc@gmail.com", "password123")
	role := entities.CreateRole("King").ApplyScopes("money:read")
	u = u.Grant(role)

	err := s.roles.Persist(role)
	s.NoError(err)
	err = s.users.Persist(u)
	s.NoError(err)

	us, err := s.users.Find(u.ID)
	s.NoError(err)
	s.Len(us, 1)
	u = us[0]
	s.Equal("gc@gmail.com", u.Email.String())
	s.Len(u.RoleIDs, 1)
	s.Equal("king", u.RoleIDs[0].String())
}
