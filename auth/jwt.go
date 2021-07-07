package auth

import (
	"errors"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type claims struct {
	jwt.StandardClaims
	Scopes string `json:"scopes"`
}

func CreateAccessTokenClaims(c Credentials) claims {
	return claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    "users",
			Subject:   c.ID.String(),
			ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
		},
		Scopes: strings.Join(c.Scopes, " "),
	}
}

func CreateRefreshTokenClaims(c Credentials) jwt.StandardClaims {
	return jwt.StandardClaims{
		Issuer:    "users",
		Subject:   c.ID.String(),
		ExpiresAt: time.Now().Add(720 * time.Hour).Unix(),
	}
}

func SignTokenClaims(claims jwt.Claims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(secret))
}

func ReadAccessToken(tokenString string, secret string) (Credentials, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return Credentials{}, errors.New("Invalid token")
	}

	claims, ok := token.Claims.(*claims)
	if !ok {
		return Credentials{}, errors.New("Invalid token")
	}

	if !token.Valid {
		return Credentials{}, errors.New("Invalid token")
	}

	ID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return Credentials{}, errors.New("Invalid token")
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
