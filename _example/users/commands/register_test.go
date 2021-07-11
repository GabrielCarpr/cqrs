package commands_test

import (
	"example/internal/support"
	"example/users/commands"
	"example/users/db"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupRegister() (*db.MemoryUserRepository, *db.MemoryRoleRepository, *commands.RegisterHandler) {
	users := db.NewMemoryUserRepository()
	roles := db.NewMemoryRoleRepository()
	handler := commands.NewRegisterHandler(users, roles)
	return users, roles, handler
}

func TestRegister(t *testing.T) {
	users, _, handler := setupRegister()

	cmd := commands.Register{}.WithEmail("gabriel@gmail.com")
	cmd = cmd.WithName("Gabriel Carpreau")
	cmd = cmd.WithPassword("password123")
	res, _ := handler.Execute(context.Background(), cmd)

	assert.NoError(t, res.Error)
	ID := support.MustParseID(res.ID)

	u, _ := users.Find(ID)
	assert.Len(t, u, 1)
	user := u[0]
	assert.Equal(t, "Gabriel Carpreau", user.Name)
	assert.True(t, user.Active)
	assert.Equal(t, "gabriel@gmail.com", user.Email.String())
	assert.True(t, user.CheckPassword("password123"))
	assert.Len(t, user.RoleIDs, 1)
}
