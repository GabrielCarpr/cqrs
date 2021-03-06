package queries_test

import (
	"example/internal/support"
	"example/internal/tester"
	"github.com/GabrielCarpr/cqrs/auth"
	"example/users/db"
	"example/users/entities"
	"example/users/queries"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type UsersTest struct {
	suite.Suite
	tester.Integration

	users   db.UserRepository
	user    entities.User
	handler *queries.UsersHandler
}

func (s *UsersTest) SetupTest() {
	s.Integration.SetupTest()

	s.users = s.Get("user-repository").(db.UserRepository)
	s.user, _ = entities.Register("gc@gmail.com", "password123")
	role := entities.CreateRole("Admin")
	urole := entities.CreateRole("User")
	s.user = s.user.Grant(role).Grant(urole)
	basicUser, _ := entities.Register("basic@gmail.com", "password123")
	basicUser = basicUser.Grant(urole)

	s.users.Persist(s.user)
	s.users.Persist(basicUser)
	s.handler = s.Get(queries.UsersHandler{}).(*queries.UsersHandler)
}

func TestUsers(t *testing.T) {
	suite.Run(t, new(UsersTest))
}

func (s *UsersTest) TestGetUsersByID() {
	q := queries.Users{IDs: &[]support.ID{s.user.ID}}
	s.NoError(q.Valid())
	var res support.PaginatedQuery
	err := s.handler.Execute(auth.TestCtx(uuid.New(), "users:read"), q, &res)

	s.NoError(err)
	users := res.Data.([]entities.User)
	s.Len(users, 1)
}

func (s *UsersTest) TestGetUsersByRoleID() {
	q := queries.Users{RoleIDs: &[]entities.RoleID{entities.NewRoleID("admin")}}
	var res support.PaginatedQuery
	err := s.Bus().Query(auth.TestCtx(uuid.New(), "users:read"), q, &res)

	s.NoError(err)
	users := res.Data.([]entities.User)
	s.Len(users, 2)
	u := users[1]
	s.Equal("gc@gmail.com", u.Email.String())
}

func (s *UsersTest) TestGetAllUsers() {
	page := 1
	perPage := 10
	q := queries.Users{Paging: support.Paging{Page: &page, PerPage: &perPage}}
	var res support.PaginatedQuery
	err := s.Bus().Query(auth.TestCtx(uuid.New(), "users:read"), q, &res)

	s.NoError(err)
	s.Equal(int32(3), res.Metadata.Count)
	users := res.Data.([]entities.User)
	s.Len(users, 3)
}
