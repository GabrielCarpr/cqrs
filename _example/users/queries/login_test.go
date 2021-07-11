package queries_test

import (
	"example/internal/tester"
	"example/users/queries"
	"example/users/readmodels"
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LoginTest struct {
	suite.Suite
	tester.Integration
}

func TestLogin(t *testing.T) {
	suite.Run(t, new(LoginTest))
}

func (s LoginTest) TestLogin() {
	query := queries.Login{Email: "gabriel.carpreau@gmail.com", Password: "password123"}

	var res readmodels.Authentication
	err := s.Bus().Query(context.Background(), query, &res)
	s.NoError(err)
	s.NotNil(res)
	s.Contains(res.Scopes, "user:")
	s.Contains(res.Scopes, "access:admin")
	s.Contains(res.Scopes, "access:user")
	s.Contains(res.Scopes, "self:read")
}

func (s LoginTest) TestLoginNotFound() {
	query := queries.Login{Email: "harrypotter@gmail.com", Password: "password"}

	var res readmodels.Authentication
	err := s.Bus().Query(context.Background(), query, &res)
	s.Error(err)
	s.ErrorIs(err, queries.LoginFail)
	s.Empty(res.AccessToken)
}
