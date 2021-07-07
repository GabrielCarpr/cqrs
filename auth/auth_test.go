package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAccessTokenClaimsScopes(t *testing.T) {
	creds := Credentials{ID: uuid.New(), Scopes: []string{"hello", "there"}}
	res := CreateAccessTokenClaims(creds)
	if res.Scopes != "hello there" {
		t.Errorf("Did not join scopes %s", res.Scopes)
	}

	creds = Credentials{ID: uuid.New(), Scopes: []string{"hello"}}
	res = CreateAccessTokenClaims(creds)
	if res.Scopes != "hello" {
		t.Errorf("Scopes were malformed %s", res.Scopes)
	}
}

func TestSignTokenClaims(t *testing.T) {
	creds := Credentials{ID: uuid.New(), Scopes: []string{"hello", "world"}}
	claims := CreateAccessTokenClaims(creds)

	res, err := SignTokenClaims(claims, "secret")
	if err != nil {
		t.Errorf("Produced error %s", err)
	}
	if len(strings.Split(res, ".")) != 3 {
		t.Errorf("Signed claims were not a valid JWT %s", res)
	}
}

func TestHashPassword(t *testing.T) {
	hash := HashPassword("top-secret")

	if hash == "top-secret" {
		t.Errorf("Hash was equal to password %s", hash)
	}
}

func TestCheckPassword(t *testing.T) {
	hash := HashPassword("top-secret")

	err := CheckPassword(hash, "top-secret")
	if err != nil {
		t.Errorf("Valid password produced error: %s", err)
	}

	err = CheckPassword(hash, "wrong-pass")
	if err == nil {
		t.Errorf("Invalid password passed")
	}
}

func TestCheckAccessToken(t *testing.T) {
	ID := uuid.New()
	creds := Credentials{[]string{"hello", "world"}, ID}
	claims := CreateAccessTokenClaims(creds)
	res, _ := SignTokenClaims(claims, "secret")

	authCtx, err := ReadAccessToken(res, "secret")
	if err != nil {
		t.Errorf("Produced error: %w", err)
	}

	if authCtx.ID != ID {
		t.Errorf("Token ID (%s) not equal to input (%s)", authCtx.ID.String(), ID.String())
	}

	if len(authCtx.Scopes) != 2 {
		t.Errorf("Wrong number of scopes: %v", authCtx.Scopes)
	}
	if authCtx.Scopes[1] != "world" {
		t.Errorf("Scopes missing world")
	}
}

func TestCheckExpiredToken(t *testing.T) {
	claims := jwt.StandardClaims{
		Issuer:    "users",
		Subject:   "hello",
		ExpiresAt: time.Now().Add(-2 * time.Minute).Unix(),
	}
	res, _ := SignTokenClaims(claims, "secret")

	authCtx, err := ReadAccessToken(res, "secret")

	if err == nil {
		t.Errorf("Error was not produced")
	}
	if authCtx.ID != uuid.Nil {
		t.Errorf("authCtx ID was not nil UUID: %s", authCtx.ID.String())
	}
}

func TestCheckTokenIncorrectKey(t *testing.T) {
	claims := jwt.StandardClaims{
		Issuer:    "users",
		Subject:   "hello",
		ExpiresAt: time.Now().Add(-2 * time.Minute).Unix(),
	}
	res, _ := SignTokenClaims(claims, "secret")

	authCtx, err := ReadAccessToken(res, "secrett")

	if err == nil {
		t.Errorf("Error was not produced")
	}
	if authCtx.ID != uuid.Nil {
		t.Errorf("authCtx ID was not nil UUID: %s", authCtx.ID.String())
	}
}

func TestEnforce(t *testing.T) {
	tests := []struct {
		name           string
		userScopes     []string
		requiresScopes [][]string
		valid          bool
	}{
		{
			"Valid basic scopes",
			[]string{"self:write", "users:read"},
			[][]string{{"users:read", "self:write"}},
			true,
		},
		{
			"Valid wildcard scopes",
			[]string{"self:*", "users:read"},
			[][]string{{"self:write", "users:read"}},
			true,
		},
		{
			"2 valid wildcard scopes",
			[]string{"self:*", "users:*"},
			[][]string{{"self:read", "users:read"}},
			true,
		},
		{
			"Missing single scope",
			[]string{"self:*"},
			[][]string{{"self:read", "self:write", "users:read"}},
			false,
		},
		{
			"Missing multiple scopes",
			[]string{"self:*"},
			[][]string{{"self:read", "users:write", "books:read"}},
			false,
		},
		{
			"Missing all scopes",
			[]string{"self:read", "self:write"},
			[][]string{{"books:read", "books:write", "movies:read"}},
			false,
		},
		{
			"No scopes",
			[]string{},
			[][]string{{"users:write"}},
			false,
		},
		{
			"OR scopes",
			[]string{"self:write"},
			[][]string{{"users:write"}, {"self:write"}},
			true,
		},
		{
			"Multi OR scopes",
			[]string{"self:write"},
			[][]string{{"plans:read", "self:read"}, {"self:read"}, {"self:write"}},
			true,
		},
		{
			"Mutli No Scopes",
			[]string{"self:read"},
			[][]string{{"plans:read", "users:read"}, {"self:write"}},
			false,
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.Background()
			ac := Credentials{c.userScopes, uuid.New()}
			ctx = WithCredentials(ctx, ac)

			err := Enforce(ctx, c.requiresScopes...)
			if c.valid && err != nil {
				t.Error("Produced an error when should be valid")
			}
			if !c.valid && err == nil {
				t.Errorf("Did not produce error when should")
			}
		})
	}
}

func TestEnforceMissingScopes(t *testing.T) {
	ctx := context.Background()

	err := Enforce(ctx, []string{"users:read", "users:write"})
	if err == nil {
		t.Errorf("Error not produced with missing context")
	}
}

func TestIsUser(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()
	ac := Credentials{[]string{}, userID}
	ctx = WithCredentials(ctx, ac)

	if !IsUser(ctx, userID) {
		t.Errorf("Reported not used")
	}
}

func TestNotUser(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()
	ac := Credentials{[]string{}, userID}
	ctx = WithCredentials(ctx, ac)

	if IsUser(ctx, uuid.New()) {
		t.Errorf("Reported is user, was not")
	}
}

func TestNotUserMissing(t *testing.T) {
	ctx := context.Background()

	if IsUser(ctx, uuid.New()) {
		t.Errorf("Reported is user, was missing")
	}
}

func TestUserScope(t *testing.T) {
	userID := uuid.MustParse("e67347d6-9a19-4bf0-83ed-fd62d2a53906")

	scope := UserScope(userID)

	assert.Equal(t, "user:e67347d6-9a19-4bf0-83ed-fd62d2a53906", scope)
}
