package entities

import (
	"fmt"
	"github.com/GabrielCarpr/cqrs/bus"
	"strings"
)

/*
 * EVENTS
**/

type RoleCreated struct {
	bus.EventType

	Payload Role
}

func (RoleCreated) Event() string {
	return "role.created"
}

type RoleScopeApplied struct {
	bus.EventType

	Payload Scope
}

func (RoleScopeApplied) Event() string {
	return "role.scope.applied"
}

/*
 * ENTITIES
**/

func BuildRole(ID string, Label string) Role {
	i := NewRoleID(ID)
	r := Role{
		ID:         i,
		Label:      Label,
		scopes:     make(map[string]Scope),
		EventQueue: bus.NewEventQueue(i),
	}
	return r
}

// CreateRole creates a new role
func CreateRole(label string) Role {
	name := strings.ReplaceAll(label, " ", "-")
	name = strings.ToLower(name)
	id := NewRoleID(name)

	r := Role{
		ID:         id,
		Label:      label,
		scopes:     make(map[string]Scope),
		EventQueue: bus.NewEventQueue(id),
	}
	r.Publish(&RoleCreated{Payload: r})
	return r
}

// Role is a user provided role in the access control system
type Role struct {
	ID    RoleID `json:"id"`
	Label string

	scopes map[string]Scope

	bus.EventQueue
}

// Scopes returns the role's scopes
func (r Role) Scopes() []Scope {
	roleScope := fmt.Sprintf("access:%s", r.ID.String())
	r.scopes[roleScope] = Scope{roleScope}
	scopes := make([]Scope, len(r.scopes))
	i := 0
	for _, scope := range r.scopes {
		scopes[i] = scope
		i++
	}
	return scopes
}

// DropScopes removes all of a role's scopes
func (r Role) DropScopes() Role {
	r.scopes = make(map[string]Scope)
	return r
}

// ApplyScopes gives a role a list of scopes
func (r Role) ApplyScopes(scopes ...string) Role {
	for _, scope := range scopes {
		s := Scope{scope}
		r.scopes[scope] = s
		r.Publish(&RoleScopeApplied{Payload: s})
	}
	return r
}

// ChangeLabel changes the role's human readable label
func (r Role) ChangeLabel(label string) Role {
	r.Label = label
	return r
}

func (r Role) Equals(role Role) bool {
	return r.ID.Equals(role.ID)
}

func ScopeNames(scopes ...Scope) []string {
	output := make([]string, len(scopes))
	for i, scope := range scopes {
		output[i] = scope.Name
	}
	return output
}
