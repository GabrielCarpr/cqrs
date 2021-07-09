package commands_test

import (
	"api/internal/tester"
	"api/pkg/auth"
	"api/pkg/util"
	"api/users/commands"
	"api/users/db"
	"api/users/entities"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

func TestRunUpdateRole(t *testing.T) {
	suite.Run(t, new(UpdateRoleTest))
}

type UpdateRoleTest struct {
	suite.Suite
	tester.Integration

	roles db.RoleRepository
}

func (s *UpdateRoleTest) SetupTest() {
	s.Integration.SetupTest()
	s.roles = s.Get("role-repository").(db.RoleRepository)
}

func (s UpdateRoleTest) TestHappyPath() {
	cmd := commands.UpdateRole{
		ID:    entities.NewRoleID("user"),
		Label: util.StrPtr("King"),
	}
	res, err := s.Bus().Dispatch(auth.TestCtx(uuid.New(), "roles:write"), cmd, true)

	s.NoError(err)
	s.NoError(res.Error)
	s.Equal("user", res.ID)
}
