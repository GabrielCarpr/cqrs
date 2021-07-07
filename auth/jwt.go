package auth

import (
	"github.com/GabrielCarpr/cqrs/errors"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

var (
	// InvalidToken is returned when a token isn't valid
	InvalidToken = errors.Error{Code: 401, Message: "Invalid token"}
)

type claims struct {
	jwt.StandardClaims
	Scopes string `json:"scopes"`
}

func createClaims(c Credentials, expires time.Duration) claims {
	return claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    "users",
			Subject:   c.ID.String(),
			ExpiresAt: time.Now().Add(expires).Unix(),
		},
		Scopes: strings.Join(c.Scopes, " "),
	}
}

func signTokenClaims(claims jwt.Claims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(secret))
}

// CreateAccessToken converts credentials into a JWT access token, expiring in 15 minutes
func CreateAccessToken(c Credentials, secret string) (string, error) {
	accessClaims := createClaims(c, time.Minute*15)
	return signTokenClaims(accessClaims, secret)
}

// CreateRefreshToken converts credentials into a JWT refresh token, expiring in 24 hours
func CreateRefreshToken(c Credentials, secret string) (string, error) {
	refreshClaims := createClaims(c, time.Hour*24)
	return signTokenClaims(refreshClaims, secret)
}

// ReadToken reads and checks an access token
func ReadToken(tokenString string, secret string) (Credentials, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return Credentials{}, InvalidToken
	}

	claims, ok := token.Claims.(*claims)
	if !ok {
		return Credentials{}, InvalidToken
	}

	if !token.Valid {
		return Credentials{}, InvalidToken
	}

	ID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return Credentials{}, InvalidToken
	}

	var scopes []string
	if claims.Scopes == "" {
		scopes = []string{}
	} else {
		scopes = strings.Split(claims.Scopes, " ")
	}

	return Credentials{
		ID:     ID,
		Scopes: scopes,
	}, nil
}
