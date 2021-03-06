package entities_test

import (
	"example/users/entities"
	"github.com/GabrielCarpr/cqrs/errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	user, err := entities.Register("gabriel.carpreau@gmail.com", "password123")

	assert.NoError(t, err)
	assert.Equal(t, user.Email.String(), "gabriel.carpreau@gmail.com")
	events := user.Commit()
	assert.Len(t, events, 1)
	assert.IsType(t, &entities.UserCreated{}, events[0])
	assert.Equal(t, user.ID.String(), events[0].(*entities.UserCreated).Owner)
}

func TestRegisterAndPassword(t *testing.T) {
	user, err := entities.Register("gabriel.carpreau@gmail.com", "password123")

	assert.NoError(t, err)
	assert.True(t, user.CheckPassword("password123"))
}

func TestRegisterBadPassword(t *testing.T) {
	_, err := entities.Register("gc@gmail.com", "123")
	herr := err.(errors.Error)
	assert.Equal(t, 400, herr.Code)
}

func TestCheckPassword(t *testing.T) {
	user, err := entities.Register("gc@gmail.com", "password234")
	assert.NoError(t, err)
	assert.False(t, user.CheckPassword("password123"))
	assert.True(t, user.CheckPassword("password234"))
}

func TestChangePassword(t *testing.T) {
	user, _ := entities.Register("gc@gmail.com", "password123")

	user, err := user.ChangePassword("helloworld")
	assert.NoError(t, err)
	assert.False(t, user.CheckPassword("password123"))
	assert.True(t, user.CheckPassword("helloworld"))
}

func TestRegisterInvalidEmail(t *testing.T) {
	_, err := entities.Register("gc.com", "password123")
	httpErr := err.(errors.Error)
	assert.Equal(t, 400, httpErr.Code)
	assert.Contains(t, httpErr.Message, "Email")
}

func TestGetUserScopes(t *testing.T) {
	u, _ := entities.Register("gabriel@gmail.com", "password123")

	scopes := u.Scopes()
	assert.Len(t, scopes, 1)
}

func TestGrantRole(t *testing.T) {
	u, _ := entities.Register("gabriel@gmail.com", "password123")
	u.Commit()
	r := entities.CreateRole("Ambassador")

	u = u.Grant(r)

	assert.Len(t, u.RoleIDs, 1)
	assert.Contains(t, u.RoleIDs, entities.NewRoleID("ambassador"))
	events := u.Commit()
	assert.Len(t, events, 1)
	assert.IsType(t, &entities.UserGrantedRole{}, events[0])
}

func TestAllScopes(t *testing.T) {
	u, _ := entities.Register("gabriel@gmail.com", "password123")
	r := entities.CreateRole("User").ApplyScopes("self:read", "self:write")
	u = u.Grant(r)
	admin := entities.CreateRole("Admin").ApplyScopes("users:write")

	scopes := entities.AllScopes(u, r, admin)

	assert.Len(t, scopes, 4)
	assert.NotContains(t, scopes, entities.Scope{"users:write"})
	assert.Contains(t, scopes, entities.Scope{"self:read"})
}

func TestRevokeAll(t *testing.T) {
	u, _ := entities.Register("gc@gmail.com", "password123")
	r := entities.CreateRole("User")
	admin := entities.CreateRole("Admin")
	u = u.Grant(r).Grant(admin)
	assert.Len(t, u.RoleIDs, 2)
	u.Commit()

	u = u.Revoke()

	assert.Len(t, u.RoleIDs, 0)
	events := u.Commit()
	assert.Len(t, events, 1)
	assert.IsType(t, &entities.UserRevokedAllRoles{}, events[0])
	event := events[0].(*entities.UserRevokedAllRoles)
	assert.Len(t, event.Payload, 2)
}

func TestRevokeOne(t *testing.T) {
	u, _ := entities.Register("gc@gmail.com", "password123")
	r := entities.CreateRole("User")
	admin := entities.CreateRole("Admin")
	u = u.Grant(admin).Grant(r)
	assert.Len(t, u.RoleIDs, 2)
	u.Commit()

	u = u.Revoke(admin)

	assert.Len(t, u.RoleIDs, 1)
	assert.True(t, u.RoleIDs[0].Equals(r.ID))
	events := u.Commit()
	assert.Len(t, events, 1)
	assert.IsType(t, &entities.UserRevokedRole{}, events[0])
}
