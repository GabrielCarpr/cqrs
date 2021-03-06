package commands_test

import (
	"github.com/GabrielCarpr/cqrs/auth"
	"github.com/GabrielCarpr/cqrs/errors"
	"example/pkg/util"
	"example/users/commands"
	"example/users/db"
	"example/users/entities"
	"github.com/google/uuid"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupUpdateUser() (db.UserRepository, db.RoleRepository, *commands.UpdateUserHandler, entities.User) {
	users := db.NewMemoryUserRepository()
	roles := db.NewMemoryRoleRepository()
	r, _ := roles.Find(entities.NewRoleID("user"))
	handler := commands.NewUpdateUserHandler(users, roles)
	u, _ := entities.Register("gc@gmail.com", "password123")
	u = u.Grant(r[0])
	users.Persist(u)

	return users, roles, handler, u
}

func TestUpdatesSelf(t *testing.T) {
	users, _, handler, u := setupUpdateUser()

	cmd := commands.UpdateUser{ID: u.ID, Name: util.StrPtr("Gabriel")}
	res, _ := handler.Execute(auth.TestCtx(u.ID.UUID, "self:write"), cmd)

	assert.NoError(t, res.Error)
	assert.Equal(t, cmd.ID, u.ID)
	us, _ := users.Find(u.ID)
	assert.Len(t, us, 1)
	u = us[0]
	assert.Equal(t, u.Name, "Gabriel")
	assert.Equal(t, u.Email.String(), "gc@gmail.com")
	assert.Len(t, u.RoleIDs, 1)
}

func TestUpdateShortPassword(t *testing.T) {
	_, _, handler, u := setupUpdateUser()

	cmd := commands.UpdateUser{ID: u.ID, Password: util.StrPtr("123")}
	res, _ := handler.Execute(auth.TestCtx(u.ID.UUID, "self:write"), cmd)

	assert.Error(t, res.Error)
	assert.IsType(t, res.Error, errors.Error{})
	assert.Contains(t, res.Error.Error(), "Password")
}

func TestUpdatePassword(t *testing.T) {
	users, _, handler, u := setupUpdateUser()

	cmd := commands.UpdateUser{ID: u.ID, Password: util.StrPtr("helloworld")}
	assert.NoError(t, cmd.Valid())
	res, _ := handler.Execute(auth.TestCtx(u.ID.UUID, "self:write"), cmd)

	assert.NoError(t, res.Error)
	assert.Equal(t, u.ID.String(), res.ID)
	us, _ := users.Find(u.ID)
	uv := us[0]
	assert.True(t, uv.CheckPassword("helloworld"))
}

func TestCannotUpdateRoles(t *testing.T) {
	users, _, handler, u := setupUpdateUser()

	cmd := commands.UpdateUser{ID: u.ID, Roles: &[]string{"admin"}}
	assert.NoError(t, cmd.Valid())
	res, _ := handler.Execute(auth.TestCtx(u.ID.UUID, "self:write"), cmd)

	assert.Error(t, res.Error)
	assert.ErrorIs(t, res.Error, auth.Forbidden)
	us, _ := users.Find(u.ID)
	u = us[0]
	assert.Len(t, u.RoleIDs, 1)
}

func TestAdminUpdateRoles(t *testing.T) {
	users, _, handler, u := setupUpdateUser()

	cmd := commands.UpdateUser{ID: u.ID, Roles: &[]string{"admin"}}
	res, _ := handler.Execute(auth.TestCtx(uuid.New(), "users:write"), cmd)

	assert.NoError(t, res.Error)
	assert.Equal(t, u.ID.String(), res.ID)
	us, _ := users.Find(u.ID)
	u = us[0]
	assert.Len(t, u.RoleIDs, 1)
	assert.Equal(t, u.RoleIDs[0], entities.NewRoleID("admin"))
}

func TestAdminRemoveRoles(t *testing.T) {
	users, _, handler, u := setupUpdateUser()

	cmd := commands.UpdateUser{ID: u.ID, Roles: &[]string{}}
	assert.NoError(t, cmd.Valid())
	res, _ := handler.Execute(auth.TestCtx(uuid.New(), "users:write"), cmd)

	assert.NoError(t, res.Error)
	assert.Equal(t, u.ID.String(), res.ID)
	us, _ := users.Find(u.ID)
	u = us[0]
	assert.Len(t, u.RoleIDs, 0)
}

func TestAdminEmptyPatch(t *testing.T) {
	users, _, handler, u := setupUpdateUser()

	cmd := commands.UpdateUser{ID: u.ID}
	assert.NoError(t, cmd.Valid())
	res, _ := handler.Execute(auth.TestCtx(uuid.New(), "users:write"), cmd)

	assert.NoError(t, res.Error)
	us, _ := users.Find(u.ID)
	u = us[0]
	assert.Len(t, u.RoleIDs, 1)
	assert.Equal(t, u.RoleIDs[0], entities.NewRoleID("user"))
	assert.Equal(t, "", u.Name)
	assert.Equal(t, "gc@gmail.com", u.Email.String())
}
