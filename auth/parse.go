package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"
	"gosalusa.com/internal/helpers"
)

var ErrInvalidToken = fmt.Errorf("invalid token")

func ParseOf[T jwt.Claims](token string) (T, error) {
	claims, err := helpers.NewOf[T]()
	if err != nil {
		var zero T
		return zero, err
	}
	t, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("expected HMAC received %v: %w", token.Header["alg"], ErrUnexpectedAlgorithm)
		}

		return getAppKey(), nil
	})
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to parse JWT: %w", err)
	}
	if !t.Valid {
		var zero T
		return zero, ErrInvalidToken
	}
	return claims, nil
}
func Parse(token string) (*Claims, error) {
	return ParseOf[*Claims](token)
}
