package readmodels

import (
	"example/internal/config"
	"github.com/GabrielCarpr/cqrs/auth"
	"example/users/entities"
	"strings"
	"fmt"
	"time"
)

func NewAuthentication(conf *config.Config, user entities.User, roles ...entities.Role) (Authentication, error) {
	scopes := entities.ScopeNames(entities.AllScopes(user, roles...)...)
	credentials := auth.Credentials{Scopes: scopes, ID: user.ID.UUID}
	accessToken, err := auth.CreateAccessToken(credentials, conf.Secret)
	if err != nil {
		return Authentication{}, fmt.Errorf("%w", err)
	}

	refreshToken, err := auth.CreateRefreshToken(credentials, conf.Secret)
	if err != nil {
		return Authentication{}, fmt.Errorf("%w", err)
	}

	return Authentication{
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		TokenType:      "JWT",
		JWTTokenExpiry: int32(time.Now().Add(time.Minute * 15).Unix()),
		Scopes:         strings.Join(scopes, " "),
	}, nil
}

type Authentication struct {
	AccessToken    string `json:"access_token"`
	TokenType      string `json:"token_type"`
	JWTTokenExpiry int32  `json:"jwt_token_expiry"`
	Scopes         string `json:"scopes"`
	RefreshToken   string `json:"-"`
}
