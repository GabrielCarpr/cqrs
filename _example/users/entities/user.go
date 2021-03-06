package entities

import (
	"example/internal/errs"
	"example/internal/support"
	"github.com/GabrielCarpr/cqrs/auth"
	"github.com/GabrielCarpr/cqrs/bus"
	"fmt"
	"time"
)

/*
 * EVENTS
**/
type UserCreated struct {
	bus.EventType

	Payload User
}

func (UserCreated) Event() string {
	return "user.created"
}

type UserDetailsChanged struct {
	bus.EventType

	Payload User
}

func (UserDetailsChanged) Event() string {
	return "user.details.changed"
}

type UserGrantedRole struct {
	bus.EventType

	Payload Role
}

func (UserGrantedRole) Event() string {
	return "user.roles.granted"
}

type UserRevokedRole struct {
	bus.EventType

	Payload Role
}

func (UserRevokedRole) Event() string {
	return "user.roles.revoked"
}

type UserRevokedAllRoles struct {
	bus.EventType

	Payload []RoleID
}

func (UserRevokedAllRoles) Event() string {
	return "user.roles.all-revoked"
}

/**
 * MISC
 */

func validatePassword(password string) bool {
	return len(password) > 7 && len(password) < 128
}

/**
 * USER
**/

// Register creates a new user
func Register(email string, password string) (User, error) {
	e, err := support.NewEmail(email)
	if err != nil {
		return User{}, err
	}
	id := support.NewID()
	u := User{
		ID:         id,
		Email:      e,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Active:     true,
		EventBuffer: bus.NewEventBuffer(id, "User"),
	}
	u.Buffer(true, &UserCreated{Payload: u})
	return u.ChangePassword(password)
}

// User is a user of the application
type User struct {
	ID           support.ID `json:"ID"`
	Name         string `json:"name"`
	Email        support.Email `json:"email"`
	RoleIDs      []RoleID `json:"role_ids"`
	Hash         string   `json:"-"`
	Active       bool `json:"active"`
	LastSignedIn *time.Time `json:"last_signed_in"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	bus.EventBuffer `json:"version"`
}

// ChangePassword sets the users password, while hashing it
func (u User) ChangePassword(password string) (User, error) {
	if !validatePassword(password) {
		return u, errs.ValidationError("Password not valid")
	}
	u.Hash = auth.HashPassword(password)
	return u, nil
}

// CheckPassword returns if the provided password matches the users
func (u User) CheckPassword(password string) bool {
	err := auth.CheckPassword(u.Hash, password)
	return err == nil
}

// Scopes returns the user provided scopes
func (u User) Scopes() []Scope {
	userScope := fmt.Sprintf("user:%s", u.ID.String())
	return []Scope{Scope{userScope}}
}

// Grant gives a user a role. If the user has the role already, noop
func (u User) Grant(role Role) User {
	for _, roleID := range u.RoleIDs {
		if role.ID.Equals(roleID) {
			return u
		}
	}
	u.RoleIDs = append(u.RoleIDs, role.ID)
	u.Buffer(true, &UserGrantedRole{Payload: role})
	return u
}

func (u User) Revoke(roles ...Role) User {
	if len(roles) == 0 {
		u.Buffer(true, &UserRevokedAllRoles{Payload: u.RoleIDs})
		u.RoleIDs = []RoleID{}
		return u
	}

	new := make([]RoleID, 0)
	for _, roleID := range u.RoleIDs {
		for _, role := range roles {
			if role.ID.Equals(roleID) {
				u.Buffer(true, &UserRevokedRole{Payload: role})
				break
			}
			new = append(new, roleID)
		}
	}
	u.RoleIDs = new
	return u
}

// Is returns whether the user is a provided role
func (u User) Is(role Role) bool {
	for _, roleID := range u.RoleIDs {
		if role.ID.Equals(roleID) {
			return true
		}
	}
	return false
}

// AllScopes returns all the scopes of a user and roles
func AllScopes(u User, roles ...Role) []Scope {
	scopes := u.Scopes()
	for _, r := range roles {
		if !u.Is(r) {
			continue
		}
		scopes = append(scopes, r.Scopes()...)
	}
	return scopes
}

// ChangeName changes the user's stored name
func (u User) ChangeName(name string) User {
	u.Name = name
	u.Buffer(true, &UserDetailsChanged{Payload: u})
	return u
}

// ChangeEmail changes the user's email
func (u User) ChangeEmail(email string) (User, error) {
	e, err := support.NewEmail(email)
	if err != nil {
		return u, err
	}
	u.Email = e
	u.Buffer(true, &UserDetailsChanged{Payload: u})
	return u, nil
}
