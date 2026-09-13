package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/openapidoc"
)

func TestSetAppKey(t *testing.T) {
	old := appKey
	defer func() { appKey = old }()

	SetAppKey([]byte("test-key"))
	assert.Equal(t, []byte("test-key"), appKey)

	appKey = nil
	key := getAppKey()
	assert.Len(t, key, 256)
}

func TestGetClaims(t *testing.T) {
	ctx := SetClaims(context.Background(), &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "sub",
		},
	})

	claims, ok := GetClaimsCtx(ctx)
	assert.True(t, ok)
	assert.Equal(t, "sub", claims.Subject)

	r := httptest.NewRequest("GET", "https://example.com", http.NoBody)
	r = r.WithContext(ctx)
	claims, ok = GetClaims(r)
	assert.True(t, ok)
	assert.Equal(t, "sub", claims.Subject)

	_, ok = GetClaimsCtx(context.Background())
	assert.False(t, ok)
}

func TestAttachUser(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetClaims(r)
		assert.False(t, ok)
		assert.Nil(t, claims)
	})

	t.Run("no header", func(t *testing.T) {
		r := httptest.NewRequest("GET", "https://example.com", http.NoBody)
		AttachUser()(next).ServeHTTP(httptest.NewRecorder(), r)
	})

	t.Run("already has claims", func(t *testing.T) {
		r := httptest.NewRequest("GET", "https://example.com", http.NoBody)
		r = setClaims(r, &Claims{
			RegisteredClaims: jwt.RegisteredClaims{Subject: "sub"},
		})
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaims(r)
			assert.True(t, ok)
			assert.Equal(t, "sub", claims.Subject)
		})
		AttachUser()(h).ServeHTTP(httptest.NewRecorder(), r)
	})

	t.Run("valid token", func(t *testing.T) {
		token, err := GenerateToken(&Claims{
			RegisteredClaims: jwt.RegisteredClaims{Subject: "sub"},
			Scope:            ScopeStrings{ScopeAccess},
		})
		assert.NoError(t, err)

		r := httptest.NewRequest("GET", "https://example.com", http.NoBody)
		r.Header.Add("Authorization", "Bearer "+token)
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaims(r)
			assert.True(t, ok)
			assert.Equal(t, "sub", claims.Subject)
		})
		AttachUser()(h).ServeHTTP(httptest.NewRecorder(), r)
	})

	t.Run("no access scope", func(t *testing.T) {
		token, err := GenerateToken(&Claims{
			RegisteredClaims: jwt.RegisteredClaims{Subject: "sub"},
			Scope:            ScopeStrings{ScopeRefresh},
		})
		assert.NoError(t, err)

		r := httptest.NewRequest("GET", "https://example.com", http.NoBody)
		r.Header.Add("Authorization", "Bearer "+token)
		AttachUser()(next).ServeHTTP(httptest.NewRecorder(), r)
	})

	t.Run("invalid token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "https://example.com", http.NoBody)
		r.Header.Add("Authorization", "Bearer not-a-token")
		AttachUser()(next).ServeHTTP(httptest.NewRecorder(), r)
	})
}

func TestLoggedIn(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("no claims", func(t *testing.T) {
		r := httptest.NewRequest("GET", "https://example.com", http.NoBody)
		w := httptest.NewRecorder()
		LoggedIn().Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("has claims", func(t *testing.T) {
		r := httptest.NewRequest("GET", "https://example.com", http.NoBody)
		r = setClaims(r, &Claims{})
		w := httptest.NewRecorder()
		LoggedIn().Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestLoggedInSecurityDefinition(t *testing.T) {
	m := LoggedIn()
	assert.Equal(t, openapidoc.DefaultSecurityDefinitionName, m.securityDefinitionName)

	op := m.OperationMiddleware(&spec.Operation{})
	assert.Len(t, op.Security, 1)

	m2 := LoggedIn().SecurityDefinition("")
	op = m2.OperationMiddleware(&spec.Operation{})
	assert.Nil(t, op.Security)
}

func TestHasClaimMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "https://example.com", http.NoBody)

	t.Run("no claims", func(t *testing.T) {
		w := httptest.NewRecorder()
		HasClaim(func(c *Claims) bool { return true }).
			Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	r = setClaims(r, &Claims{})
	t.Run("validate false", func(t *testing.T) {
		w := httptest.NewRecorder()
		HasClaim(func(c *Claims) bool { return false }).
			Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("validate true", func(t *testing.T) {
		w := httptest.NewRecorder()
		HasClaim(func(c *Claims) bool { return true }).
			Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestScopeStringsJSON(t *testing.T) {
	s := ScopeStrings{"access", "refresh"}
	b, err := s.MarshalJSON()
	assert.NoError(t, err)
	assert.JSONEq(t, `"access refresh"`, string(b))

	var out ScopeStrings
	err = out.UnmarshalJSON(b)
	assert.NoError(t, err)
	assert.Equal(t, s, out)

	var empty ScopeStrings
	err = empty.UnmarshalJSON([]byte{})
	assert.NoError(t, err)
	assert.Nil(t, empty)

	err = empty.UnmarshalJSON([]byte("not-json"))
	assert.Error(t, err)
}

func TestParseErrors(t *testing.T) {
	_, err := Parse("not-a-token")
	assert.Error(t, err)

	token, err := GenerateToken(&Claims{})
	assert.NoError(t, err)
	_, err = Parse(token + "extra")
	assert.Error(t, err)

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	badToken, err := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{"sub": "test"}).SignedString(key)
	assert.NoError(t, err)
	_, err = Parse(badToken)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnexpectedAlgorithm)

	_, err = Parse("")
	assert.Error(t, err)
}
