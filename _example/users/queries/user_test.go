package queries_test

import (
	"example/internal/support"
	"example/internal/tester"
	"github.com/GabrielCarpr/cqrs/auth"
	"example/pkg/util"
	"example/users/db"
	"example/users/entities"
	"example/users/queries"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UserTest struct {
	suite.Suite
	tester.Integration

	users db.UserRepository
	user  entities.User
}

func (s *UserTest) SetupTest() {
	s.Integration.SetupTest()

	s.users = s.Get("user-repository").(db.UserRepository)
	s.user, _ = entities.Register("gc@gmail.com", "password123")
	err := s.users.Persist(s.user)
	s.NoError(err)
}

func TestUser(t *testing.T) {
	suite.Run(t, new(UserTest))
}

func (s *UserTest) TestGetByID() {
	q := queries.User{ID: &s.user.ID}
	s.NoError(q.Valid())
	var res entities.User
	err := s.Bus().Query(auth.TestCtx(support.NewID().UUID, "users:read"), q, &res)
	s.NoError(err)
	s.Equal("gc@gmail.com", res.Email.String())
}

func (s *UserTest) TestGetByEmail() {
	q := queries.User{Email: util.StrPtr(s.user.Email.String())}
	s.NoError(q.Valid())
	var res entities.User
	err := s.Bus().Query(auth.TestCtx(support.NewID().UUID, "users:read"), q, &res)

	s.NoError(err)
	s.Equal("gc@gmail.com", res.Email.String())
	s.Equal(s.user.Email, res.Email)
}

func (s *UserTest) TestFindMissingEmail() {
	q := queries.User{Email: util.StrPtr("lolman@gmail.com")}
	var res entities.User
	err := s.Bus().Query(auth.TestCtx(support.NewID().UUID, "users:read"), q, &res)

	s.Error(err)
	s.ErrorIs(err, queries.ErrUserNotFound)
}

func (s *UserTest) TestFindMissingID() {
	id := support.NewID()
	q := queries.User{ID: &id}
	var res entities.User
	err := s.Bus().Query(auth.TestCtx(support.NewID().UUID, "users:read"), q, &res)

	s.Error(err)
	s.ErrorIs(err, queries.ErrUserNotFound)
}
