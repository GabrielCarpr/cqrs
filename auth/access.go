package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/GabrielCarpr/cqrs/errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
)

var (
	// AuthCtxKey is the key used for storing credentials in the context
	AuthCtxKey = authCtxKeyType("authCtx")

	// Forbidden is an error returned when access is denied
	Forbidden = errors.Error{Code: 403, Message: "Forbidden"}

	// BlankCredentials are carried by unauthenticated users of the system
	BlankCredentials = Credentials{}
)

type authCtxKeyType string

func (k authCtxKeyType) String() string {
	return string(k)
}

// Credentials is an access control record
type Credentials struct {
	Scopes []string  `json:"scopes"`
	ID     uuid.UUID `json:"id"`
}

// Valid determines if the Credentials are valid
func (c Credentials) Valid() bool {
	return c.ID != uuid.Nil
}

// WithCredentials returns a ctx with access control credentials
func WithCredentials(ctx context.Context, c Credentials) context.Context {
	return context.WithValue(ctx, AuthCtxKey, c)
}

// GetCredentials returns the context's credentials
func GetCredentials(ctx context.Context) Credentials {
	cred := ctx.Value(AuthCtxKey)
	if cred == nil {
		return Credentials{}
	}
	return cred.(Credentials)
}

// IsUser returns whether the provided user ID
// is the authenticated user
func IsUser(ctx context.Context, testID uuid.UUID) bool {
	creds := GetCredentials(ctx)
	if !creds.Valid() {
		return false
	}

	return testID == creds.ID
}

// UserScope returns a single scope that identifies a user
func UserScope(userID uuid.UUID) string {
	return fmt.Sprintf("user:%s", userID.String())
}

// TestCtx is a testing utility for generating any
// access control ctx
func TestCtx(userID uuid.UUID, scopes ...string) context.Context {
	userScope := UserScope(userID)
	auth := Credentials{append(scopes, userScope), userID}
	ctx := context.Background()
	return WithCredentials(ctx, auth)
}

// Enforce ensures that the context stores scopes required to satisfy
// the required scopes. The user must possess all of the required
// scopes to proceed. Accepts an or condition, eg:
// Enforce(ctx, []string{"users:write"}, []string{"self:write}"})
// So the user must have either users:write, or self:write, or both
func Enforce(ctx context.Context, requiredScopes ...[]string) error {
	if len(requiredScopes) == 0 {
		return nil
	}

	creds := GetCredentials(ctx)
	if !creds.Valid() {
		return Forbidden
	}
	userScopes := creds.Scopes
	if len(userScopes) == 0 {
		return Forbidden
	}

	accessGranted := false

	// OR Conditional - iterates over groups
	for _, group := range requiredScopes {

		// AND Conditional - iterates over 1 group
		hasGroup := false
		for _, scope := range group {

			// Iterates over user scopes, checks if one of them matches
			for _, uScope := range userScopes {
				if scopeSatisfiesScope(scope, uScope) {
					hasGroup = true
					break
				} else {
					hasGroup = false
				}
			}
			if !hasGroup {
				break
			}
		}
		if hasGroup {
			accessGranted = true
			break
		}

	}

	if accessGranted == false {
		return Forbidden
	}

	return nil
}

func scopeSatisfiesScope(requiredScope string, accessScope string) bool {
	if requiredScope == accessScope {
		return true
	}

	rParts := strings.Split(requiredScope, ":")
	rResource := rParts[0]

	aParts := strings.Split(accessScope, ":")
	aResource := aParts[0]
	aScope := aParts[1]

	if rResource == aResource && aScope == "*" {
		return true
	}

	return false
}

// HashPassword generates a hash for a provided password
func HashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 13)
	if err != nil {
		panic(err)
	}
	return string(hash)
}

// CheckPassword takes a hash and a password and returns
// an error if the password doesn't match the hash
func CheckPassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
