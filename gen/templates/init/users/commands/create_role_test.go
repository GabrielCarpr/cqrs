package commands_test

import (
	"api/pkg/auth"
	"api/users/commands"
	"api/users/db"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupCreateRole() (db.RoleRepository, *commands.CreateRoleHandler) {
	roles := db.NewMemoryRoleRepository()
	h := commands.NewCreateRoleHandler(roles)
	return roles, h
}

func TestCreateRole(t *testing.T) {
	roles, h := setupCreateRole()

	cmd := commands.CreateRole{Name: "King", Scopes: []string{"royalty:write"}}
	assert.NoError(t, cmd.Valid())
	res, _ := h.Execute(auth.TestCtx(uuid.New(), "roles:write"), cmd)

	assert.NoError(t, res.Error)
	assert.Equal(t, "king", res.ID)
	rs, _ := roles.All()
	assert.Len(t, rs, 3)
}
