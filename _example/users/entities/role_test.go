package entities_test

import (
	"example/users/entities"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRole_Scopes(t *testing.T) {
	role := entities.CreateRole("user").ApplyScopes("users:read", "payments:write")

	assert.Len(t, role.Scopes(), 3)
	assert.Contains(t, role.Scopes(), entities.Scope{"access:user"})
	assert.Contains(t, role.Scopes(), entities.Scope{"payments:write"})
	assert.NotContains(t, role.Scopes(), entities.Scope{"users:write"})
	events := role.Commit()
	assert.Len(t, events, 3)
	created := events[0].(*entities.RoleCreated)
	assert.Equal(t, "user", created.Owner)
}
