package auth_test

import (
	"encoding/base64"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/auth"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		Name   string
		Claims *auth.Claims
	}{
		{
			Name:   "empty",
			Claims: &auth.Claims{},
		},
		{
			Name: "sub",
			Claims: &auth.Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "sub",
				},
			},
		},
		{
			Name: "type",
			Claims: &auth.Claims{
				Scope: []string{"type"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			token, err := auth.GenerateToken(tc.Claims)
			assert.NoError(t, err)
			newClaims, err := auth.Parse(token)
			assert.NoError(t, err)
			assert.Equal(t, tc.Claims, newClaims)
		})
	}
}

func TestParseMalformedToken(t *testing.T) {
	_, err := auth.Parse("not.a.token")
	assert.Error(t, err)
}

func TestParseWrongAlgorithm(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{}`))
	token := header + "." + payload + ".signature"

	_, err := auth.Parse(token)
	assert.ErrorIs(t, err, auth.ErrUnexpectedAlgorithm)
}

func TestParseOfInterface(t *testing.T) {
	_, err := auth.ParseOf[jwt.Claims]("token")
	assert.Error(t, err)
}
