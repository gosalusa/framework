package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-openapi/spec"
	"gosalusa.com/clog"
	"gosalusa.com/openapidoc"
	"gosalusa.com/request"
	"gosalusa.com/router"
)

type contextKey uint8

const (
	claimKey contextKey = iota
)

var (
	Err401Unauthorized = request.NewHTTPError(errors.New("unauthorized"), http.StatusUnauthorized)
)

var appKey []byte

func SetAppKey(key []byte) {
	appKey = key
}
func getAppKey() []byte {
	if appKey == nil {
		slog.Warn("No app key set, using a random app key. Any auth tokens will not work after the server is restarted")
		appKey = make([]byte, 256)
		_, err := rand.Read(appKey)
		if err != nil {
			panic(fmt.Errorf("could not generate app key: %w", err))
		}
	}
	return appKey
}

func GetClaims(r *http.Request) (*Claims, bool) {
	return GetClaimsCtx(r.Context())
}
func GetClaimsCtx(ctx context.Context) (*Claims, bool) {
	claims, _ := getClaimsCtx(ctx)
	return claims, claims != nil
}
func getClaimsCtx(ctx context.Context) (claims *Claims, inContext bool) {
	iClaims := ctx.Value(claimKey)
	if iClaims == nil {
		return nil, false
	}
	claims, ok := iClaims.(*Claims)
	return claims, ok
}

func setClaims(r *http.Request, claims *Claims) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), claimKey, claims))
}

func AttachUser() router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Context().Value(claimKey) != nil {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := authenticate(r)
			if errors.Is(err, ErrMissingAuthorizationHeader) {
				// noop
			} else if err != nil {
				clog.Use(r.Context()).Warn("authentication failed", "err", err)
			}

			next.ServeHTTP(w, setClaims(r, claims))
		})
	}
}

type LoggedInMiddleware struct {
	securityDefinitionName string
}

var _ router.Middleware = (*LoggedInMiddleware)(nil)
var _ openapidoc.OperationMiddleware = (*LoggedInMiddleware)(nil)

func LoggedIn() *LoggedInMiddleware {
	return &LoggedInMiddleware{
		securityDefinitionName: openapidoc.DefaultSecurityDefinitionName,
	}
}
func (m *LoggedInMiddleware) SecurityDefinition(name string) *LoggedInMiddleware {
	m.securityDefinitionName = name
	return m
}
func (m *LoggedInMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetClaims(r)
		if !ok {
			respond(Err401Unauthorized, w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
func (m *LoggedInMiddleware) OperationMiddleware(s *spec.Operation) *spec.Operation {
	if m.securityDefinitionName == "" {
		return s
	}
	if s.Security == nil {
		s.Security = []map[string][]string{}
	}
	s.Security = append(s.Security, map[string][]string{
		m.securityDefinitionName: {},
	})
	return s
}

type HasClaimMiddleware struct {
	validate func(c *Claims) bool
}

var _ router.Middleware = (*HasClaimMiddleware)(nil)

func (m *HasClaimMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetClaims(r)
		if !ok {
			respond(Err401Unauthorized, w, r)
			return
		}

		if !m.validate(claims) {
			respond(Err401Unauthorized, w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func HasClaim(validate func(c *Claims) bool) *HasClaimMiddleware {
	return &HasClaimMiddleware{
		validate: validate,
	}
}

func respond(resp request.Responder, w http.ResponseWriter, r *http.Request) {
	err := resp.Respond(w, r)
	if err != nil {
		log.Print(err)
	}
}
