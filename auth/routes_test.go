package auth_test

import (
	"context"
	"embed"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/auth"
	"gosalusa.com/database/builder"
	"gosalusa.com/database/dbtest"
	"gosalusa.com/database/model"
	"gosalusa.com/email/emailtest"
	"gosalusa.com/router"
	"gosalusa.com/router/routertest"
	"gosalusa.com/view"
)

type DevNull struct{}

func (d *DevNull) Write(p []byte) (n int, err error) {
	return len(p), nil
}

//go:embed emails/*
var emails embed.FS

var nullLogger = slog.New(slog.NewTextHandler(&DevNull{}, nil))
var usernameRoutes = auth.NewBasicAuthController[*auth.UsernameUser](auth.CreateUser(auth.NewUsernameUser))
var emailRoutes = auth.NewBasicAuthController[*auth.EmailVerifiedUser](auth.CreateUser(auth.NewEmailVerifiedUser))
var emailTemplates = view.NewViewTemplate(emails)

func TestAuthRoutesUserCreate(t *testing.T) {
	Run(t, "can create user", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()

		resp, err := usernameRoutes.RunUserCreate(&auth.UsernameUser{
			Username:     "user",
			PasswordHash: []byte{},
		}, &auth.UserCreateRequest{
			Update: dbtest.Update(tx),
			Ctx:    ctx,
			Logger: nullLogger,
		})
		assert.NoError(t, err)
		assert.Equal(t, "user", resp.User.Username)

		u, err := builder.From[*auth.UsernameUser]().WithContext(ctx).Find(tx, resp.User.ID)
		assert.NoError(t, err)
		assert.Equal(t, u, resp.User)
		assert.NotNil(t, u.PasswordHash)
		// assert.False(t, u.Validated)
		// assert.NotZero(t, u.ValidationCode)
	})

	Run(t, "force lowercase usernames", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		resp, err := usernameRoutes.RunUserCreate(&auth.UsernameUser{
			Username:     "user",
			PasswordHash: []byte{},
		}, &auth.UserCreateRequest{
			Update: dbtest.Update(tx),
			Ctx:    ctx,
			Logger: nullLogger,
		})
		assert.NoError(t, err)
		assert.Equal(t, "user", resp.User.Username)

		u, err := builder.From[*auth.UsernameUser]().WithContext(ctx).Find(tx, resp.User.ID)
		assert.NoError(t, err)
		assert.Equal(t, u, resp.User)
		assert.NotNil(t, u.PasswordHash)
		// assert.False(t, u.Validated)
		// assert.NotZero(t, u.ValidationCode)
	})

	Run(t, "email verification", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		urlResolver := routertest.NewTestResolver()
		m := emailtest.NewTestMailer()
		resp, err := emailRoutes.RunUserCreate(&auth.EmailVerifiedUser{
			Email:        "user@example.com",
			PasswordHash: []byte{},
		}, &auth.UserCreateRequest{
			Update:   dbtest.Update(tx),
			Ctx:      ctx,
			Mailer:   m,
			Logger:   nullLogger,
			URL:      urlResolver,
			Template: emailTemplates,
		})
		assert.NoError(t, err)
		assert.Equal(t, "user@example.com", resp.User.Email)

		time.Sleep(time.Millisecond * 20)

		sent := m.EmailsSent()
		assert.Len(t, sent, 1)
		assert.Equal(t, []string{"user@example.com"}, sent[0].To)
		assert.Equal(t, "Verify your email", sent[0].Subject)
		assert.Contains(t, string(sent[0].HTMLBody), urlResolver.Resolve("auth.email.verify", "token", resp.User.LookupToken))

		u, err := builder.From[*auth.EmailVerifiedUser]().WithContext(ctx).Find(tx, resp.User.ID)
		assert.NoError(t, err)
		assert.Equal(t, u, resp.User)

		assert.False(t, u.Verified)
		assert.NotZero(t, u.LookupToken)
	})
}

func TestAuthRoutesLogin(t *testing.T) {

	// Hashed password salted with the id
	id := uuid.MustParse("cae3c6b1-7ff1-4f23-9489-a9f6e82478f9")
	password := "pass"
	passwordHash := []byte{
		0x24, 0x32, 0x61, 0x24, 0x30, 0x34, 0x24, 0x78, 0x4d, 0x65,
		0x30, 0x54, 0x66, 0x77, 0x4c, 0x75, 0x48, 0x79, 0x35, 0x78,
		0x64, 0x51, 0x76, 0x58, 0x6b, 0x59, 0x73, 0x4b, 0x2e, 0x36,
		0x34, 0x31, 0x70, 0x6c, 0x63, 0x6c, 0x69, 0x54, 0x43, 0x5a,
		0x51, 0x51, 0x55, 0x49, 0x71, 0x41, 0x72, 0x65, 0x77, 0x51,
		0x45, 0x4c, 0x6b, 0x43, 0x76, 0x6d, 0x6a, 0x62, 0x4d, 0x75,
	}

	Run(t, "can login", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		u := &auth.UsernameUser{
			ID:           id,
			Username:     "user",
			PasswordHash: passwordHash,
		}
		err := model.Save(tx, u)
		assert.NoError(t, err)

		resp, err := usernameRoutes.RunLogin(&auth.LoginRequest{
			Username: "user",
			Password: password,
			Read:     dbtest.Read(tx),
			Ctx:      ctx,
			Log:      nullLogger,
		})
		assert.NoError(t, err)
		assert.NotZero(t, resp.AccessToken)
		assert.NotZero(t, resp.RefreshToken)
		assert.Equal(t, "Bearer", resp.TokenType)
		assert.Equal(t, 3600, resp.ExpiresIn)
	})

	Run(t, "password is salted", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		u := &auth.UsernameUser{
			ID:           uuid.New(),
			Username:     "user",
			PasswordHash: passwordHash,
		}
		err := model.Save(tx, u)
		assert.NoError(t, err)

		_, err = usernameRoutes.RunLogin(&auth.LoginRequest{
			Username: "user",
			Password: password,
			Read:     dbtest.Read(tx),
			Ctx:      ctx,
			Log:      nullLogger,
		})
		assert.ErrorIs(t, err, auth.ErrInvalidUserPass)
	})
	Run(t, "wrong user", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		u := &auth.UsernameUser{
			ID:           id,
			Username:     "user",
			PasswordHash: passwordHash,
		}
		err := model.Save(tx, u)
		assert.NoError(t, err)

		_, err = usernameRoutes.RunLogin(&auth.LoginRequest{
			Username: "not user",
			Password: password,
			Read:     dbtest.Read(tx),
			Ctx:      ctx,
			Log:      nullLogger,
		})
		assert.ErrorIs(t, err, auth.ErrInvalidUserPass)
	})
	Run(t, "wrong pass", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		u := &auth.UsernameUser{
			ID:           id,
			Username:     "user",
			PasswordHash: passwordHash,
		}
		err := model.Save(tx, u)
		assert.NoError(t, err)

		_, err = usernameRoutes.RunLogin(&auth.LoginRequest{
			Username: "user",
			Password: "not pass",
			Read:     dbtest.Read(tx),
			Ctx:      ctx,
			Log:      nullLogger,
		})
		assert.ErrorIs(t, err, auth.ErrInvalidUserPass)
	})
}

func TestAuthRoutesVerifyEmail(t *testing.T) {
	Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		urlResolver := routertest.NewTestResolver()
		token := "test"
		id := uuid.New()
		err := model.Save(tx, &auth.EmailVerifiedUser{
			ID:           id,
			Email:        "",
			PasswordHash: []byte{},
			LookupToken:  token,
			Verified:     false,
		})
		assert.NoError(t, err)

		resp, err := emailRoutes.RunVerifyEmail(&auth.VerifyEmailRequest{
			Token:  token,
			Ctx:    ctx,
			Update: dbtest.Update(tx),
			URL:    urlResolver,
		})
		assert.NoError(t, err)

		u, err := builder.From[*auth.EmailVerifiedUser]().Find(tx, id)
		assert.NoError(t, err)
		assert.True(t, u.Verified)

		w := httptest.NewRecorder()
		resp.ServeHTTP(w, httptest.NewRequest("GET", "/", http.NoBody))
		assert.Equal(t, http.StatusFound, w.Result().StatusCode)
		assert.Equal(t, urlResolver.Resolve("login"), w.Result().Header.Get("Location"))
	})
}

func TestAuthRoutesResetPassword(t *testing.T) {
	Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		urlResolver := routertest.NewTestResolver()
		token := "lookup token"
		id := uuid.New()
		oldPasswordHash := []byte("old hash")
		err := model.Save(tx, &auth.EmailVerifiedUser{
			ID:           id,
			Email:        "",
			PasswordHash: oldPasswordHash,
			LookupToken:  token,
			Verified:     true,
		})
		assert.NoError(t, err)

		resp, err := emailRoutes.RunResetPassword(&auth.ResetPasswordRequest{
			Token:    token,
			Password: "new password",
			Ctx:      ctx,
			Update:   dbtest.Update(tx),
			URL:      urlResolver,
		})
		assert.NoError(t, err)

		u, err := builder.From[*auth.EmailVerifiedUser]().Find(tx, id)
		assert.NoError(t, err)
		assert.NotEqual(t, oldPasswordHash, u.PasswordHash)
		assert.Equal(t, u, resp.User)
	})
}

func TestAuthRoutesChangePassword(t *testing.T) {
	// Hashed password salted with the id
	id := uuid.MustParse("cae3c6b1-7ff1-4f23-9489-a9f6e82478f9")
	oldPassword := "pass"
	oldPasswordHash := []byte{
		0x24, 0x32, 0x61, 0x24, 0x30, 0x34, 0x24, 0x78, 0x4d, 0x65,
		0x30, 0x54, 0x66, 0x77, 0x4c, 0x75, 0x48, 0x79, 0x35, 0x78,
		0x64, 0x51, 0x76, 0x58, 0x6b, 0x59, 0x73, 0x4b, 0x2e, 0x36,
		0x34, 0x31, 0x70, 0x6c, 0x63, 0x6c, 0x69, 0x54, 0x43, 0x5a,
		0x51, 0x51, 0x55, 0x49, 0x71, 0x41, 0x72, 0x65, 0x77, 0x51,
		0x45, 0x4c, 0x6b, 0x43, 0x76, 0x6d, 0x6a, 0x62, 0x4d, 0x75,
	}

	Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		createdUser := &auth.UsernameUser{
			ID:           id,
			PasswordHash: oldPasswordHash,
		}
		err := model.Save(tx, createdUser)
		assert.NoError(t, err)

		resp, err := usernameRoutes.RunChangePassword(&auth.ChangePasswordRequest[*auth.UsernameUser]{
			OldPassword: oldPassword,
			NewPassword: "new password",
			User:        createdUser,
			Ctx:         ctx,
			Update:      dbtest.Update(tx),
		})
		assert.NoError(t, err)

		u, err := builder.From[*auth.UsernameUser]().Find(tx, id)
		assert.NoError(t, err)
		assert.NotEqual(t, oldPasswordHash, u.PasswordHash)
		assert.Equal(t, u, resp.User)
	})
}

func TestAuthRoutesRefresh(t *testing.T) {
	Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		createdUser := &auth.UsernameUser{
			ID:           uuid.New(),
			PasswordHash: []byte(""),
		}
		err := model.Save(tx, createdUser)
		assert.NoError(t, err)

		token, err := auth.GenerateToken(&auth.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: createdUser.GetID(),
			},
			Scope: []string{auth.ScopeRefresh},
		})
		assert.NoError(t, err)

		resp, err := usernameRoutes.RunRefresh(&auth.RefreshRequest[*auth.UsernameUser]{
			RefreshToken: token,
			Ctx:          ctx,
			Read:         dbtest.Read(tx),
		})
		assert.NoError(t, err)
		assert.Equal(t, token, resp.RefreshToken)
		assert.Equal(t, "Bearer", resp.TokenType)
		assert.Equal(t, token, resp.RefreshToken)
		assert.Equal(t, 3600, resp.ExpiresIn)

		claims, err := auth.Parse(resp.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, createdUser.GetID(), claims.Subject)
		assert.Equal(t, auth.ScopeStrings{auth.ScopeAccess}, claims.Scope)
	})
}

func TestAuthRoutesForgotPassword(t *testing.T) {
	Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		urlResolver := routertest.NewTestResolver()
		m := emailtest.NewTestMailer()

		id := uuid.New()
		err := model.Save(tx, &auth.EmailVerifiedUser{
			ID:           id,
			Email:        "user@example.com",
			PasswordHash: []byte{},
			Verified:     true,
		})
		assert.NoError(t, err)

		_, err = emailRoutes.RunForgotPassword(&auth.ForgotPasswordRequest{
			Email:    "user@example.com",
			Update:   dbtest.Update(tx),
			Ctx:      ctx,
			Mailer:   m,
			Logger:   nullLogger,
			URL:      urlResolver,
			Template: emailTemplates,
		})
		assert.NoError(t, err)

		time.Sleep(time.Millisecond * 20)

		u, err := builder.From[*auth.EmailVerifiedUser]().WithContext(ctx).Find(tx, id)
		assert.NoError(t, err)

		assert.NotZero(t, u.LookupToken)

		sent := m.EmailsSent()
		assert.Len(t, sent, 1)
		assert.Equal(t, []string{"user@example.com"}, sent[0].To)
		assert.Equal(t, "Password reset", sent[0].Subject)
		assert.Contains(t, string(sent[0].HTMLBody), urlResolver.Resolve("reset-password", "token", u.LookupToken))

	})
}

func TestAuthRoutesHandlers(t *testing.T) {
	assert.NotNil(t, usernameRoutes.UserCreate())
	assert.NotNil(t, usernameRoutes.Login())
	assert.NotNil(t, emailRoutes.VerifyEmail())
	assert.NotNil(t, emailRoutes.ResetPassword())
	assert.NotNil(t, usernameRoutes.ChangePassword())
	assert.NotNil(t, usernameRoutes.Refresh())
	assert.NotNil(t, emailRoutes.ForgotPassword())
}

func TestRegisterRoutes(t *testing.T) {
	r := router.New()
	auth.RegisterRoutes(r, usernameRoutes)

	type methodPath struct {
		Method string
		Path   string
	}
	routes := make([]methodPath, 0)
	for _, route := range r.Routes() {
		routes = append(routes, methodPath{Method: route.Method, Path: route.Path})
	}

	assert.Contains(t, routes, methodPath{"POST", "/login"})
	assert.Contains(t, routes, methodPath{"POST", "/user/password/reset"})
	assert.Contains(t, routes, methodPath{"POST", "/user/password/forgot"})
	assert.Contains(t, routes, methodPath{"POST", "/user"})
	assert.Contains(t, routes, methodPath{"GET", "/user/verify"})
	assert.Contains(t, routes, methodPath{"POST", "/login/refresh"})
	assert.Contains(t, routes, methodPath{"POST", "/user/password/change"})
}

func TestAuthTokenOptions(t *testing.T) {
	type customClaims struct {
		*auth.Claims
		Extra string `json:"extra"`
	}

	routes := auth.NewBasicAuthController[*auth.UsernameUser](
		auth.CreateUser(auth.NewUsernameUser),
		auth.AccessTokenOptions(func(u *auth.UsernameUser, claims *auth.Claims) jwt.Claims {
			return &customClaims{Claims: claims, Extra: "access-" + u.Username}
		}),
		auth.RefreshTokenOptions(func(u *auth.UsernameUser, claims *auth.Claims) jwt.Claims {
			return &customClaims{Claims: claims, Extra: "refresh-" + u.Username}
		}),
	)

	Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		_, err := routes.RunUserCreate(&auth.UsernameUser{
			Username:     "user",
			PasswordHash: []byte{},
		}, &auth.UserCreateRequest{
			Password: "pass",
			Update:   dbtest.Update(tx),
			Ctx:      ctx,
			Logger:   nullLogger,
		})
		assert.NoError(t, err)

		resp, err := routes.RunLogin(&auth.LoginRequest{
			Username: "user",
			Password: "pass",
			Read:     dbtest.Read(tx),
			Ctx:      ctx,
			Log:      nullLogger,
		})
		assert.NoError(t, err)

		accessClaims, err := auth.ParseOf[*customClaims](resp.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, "access-user", accessClaims.Extra)

		refreshClaims, err := auth.ParseOf[*customClaims](resp.RefreshToken)
		assert.NoError(t, err)
		assert.Equal(t, "refresh-user", refreshClaims.Extra)
	})
}

func TestAuthResetPasswordName(t *testing.T) {
	routes := auth.NewBasicAuthController[*auth.EmailVerifiedUser](
		auth.CreateUser(auth.NewEmailVerifiedUser),
		auth.ResetPasswordName("custom-reset-password"),
	)

	Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		urlResolver := routertest.NewTestResolver()
		m := emailtest.NewTestMailer()

		id := uuid.New()
		err := model.Save(tx, &auth.EmailVerifiedUser{
			ID:           id,
			Email:        "user2@example.com",
			PasswordHash: []byte{},
			Verified:     true,
		})
		assert.NoError(t, err)

		_, err = routes.RunForgotPassword(&auth.ForgotPasswordRequest{
			Email:    "user2@example.com",
			Update:   dbtest.Update(tx),
			Ctx:      ctx,
			Mailer:   m,
			Logger:   nullLogger,
			URL:      urlResolver,
			Template: emailTemplates,
		})
		assert.NoError(t, err)

		time.Sleep(time.Millisecond * 20)

		sent := m.EmailsSent()
		assert.Len(t, sent, 1)
		assert.Contains(t, string(sent[0].HTMLBody), urlResolver.Resolve("custom-reset-password", "token", ""))
	})
}

func TestAuthRoutesLoginUnverified(t *testing.T) {
	Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		_, err := emailRoutes.RunUserCreate(&auth.EmailVerifiedUser{
			Email:        "unverified@example.com",
			PasswordHash: []byte{},
		}, &auth.UserCreateRequest{
			Password: "pass",
			Update:   dbtest.Update(tx),
			Ctx:      ctx,
			Mailer:   emailtest.NewTestMailer(),
			Logger:   nullLogger,
			URL:      routertest.NewTestResolver(),
			Template: emailTemplates,
		})
		assert.NoError(t, err)

		time.Sleep(time.Millisecond * 20)

		_, err = emailRoutes.RunLogin(&auth.LoginRequest{
			Username: "unverified@example.com",
			Password: "pass",
			Read:     dbtest.Read(tx),
			Ctx:      ctx,
			Log:      nullLogger,
		})
		assert.ErrorIs(t, err, auth.Err401Unauthorized)
	})
}

func TestAuthRoutesVerifyEmailErrors(t *testing.T) {
	Run(t, "non email verified user", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		_, err := usernameRoutes.RunVerifyEmail(&auth.VerifyEmailRequest{
			Token:  "token",
			Ctx:    ctx,
			Update: dbtest.Update(tx),
			URL:    routertest.NewTestResolver(),
		})
		assert.ErrorIs(t, err, auth.ErrNonEmailVerifiedUser)
	})

	Run(t, "token not found", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		_, err := emailRoutes.RunVerifyEmail(&auth.VerifyEmailRequest{
			Token:  "unknown-token",
			Ctx:    ctx,
			Update: dbtest.Update(tx),
			URL:    routertest.NewTestResolver(),
		})
		assert.ErrorIs(t, err, auth.ErrTokenNotFound)
	})

	Run(t, "already verified", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		err := model.Save(tx, &auth.EmailVerifiedUser{
			ID:           uuid.New(),
			Email:        "already@example.com",
			PasswordHash: []byte{},
			LookupToken:  "already-verified-token",
			Verified:     true,
		})
		assert.NoError(t, err)

		_, err = emailRoutes.RunVerifyEmail(&auth.VerifyEmailRequest{
			Token:  "already-verified-token",
			Ctx:    ctx,
			Update: dbtest.Update(tx),
			URL:    routertest.NewTestResolver(),
		})
		assert.Error(t, err)
	})
}

func TestAuthRoutesResetPasswordErrors(t *testing.T) {
	Run(t, "non email verified user", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		_, err := usernameRoutes.RunResetPassword(&auth.ResetPasswordRequest{
			Token:    "token",
			Password: "new password",
			Ctx:      ctx,
			Update:   dbtest.Update(tx),
			URL:      routertest.NewTestResolver(),
		})
		assert.ErrorIs(t, err, auth.ErrNonEmailVerifiedUser)
	})

	Run(t, "token not found", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		_, err := emailRoutes.RunResetPassword(&auth.ResetPasswordRequest{
			Token:    "unknown-token",
			Password: "new password",
			Ctx:      ctx,
			Update:   dbtest.Update(tx),
			URL:      routertest.NewTestResolver(),
		})
		assert.ErrorIs(t, err, auth.ErrTokenNotFound)
	})

	Run(t, "not verified", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		err := model.Save(tx, &auth.EmailVerifiedUser{
			ID:           uuid.New(),
			Email:        "unverified2@example.com",
			PasswordHash: []byte{},
			LookupToken:  "unverified-token",
			Verified:     false,
		})
		assert.NoError(t, err)

		_, err = emailRoutes.RunResetPassword(&auth.ResetPasswordRequest{
			Token:    "unverified-token",
			Password: "new password",
			Ctx:      ctx,
			Update:   dbtest.Update(tx),
			URL:      routertest.NewTestResolver(),
		})
		assert.ErrorIs(t, err, auth.Err401Unauthorized)
	})
}

func TestAuthRoutesForgotPasswordUnknownEmail(t *testing.T) {
	Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		m := emailtest.NewTestMailer()

		_, err := emailRoutes.RunForgotPassword(&auth.ForgotPasswordRequest{
			Email:    "does-not-exist@example.com",
			Update:   dbtest.Update(tx),
			Ctx:      ctx,
			Mailer:   m,
			Logger:   nullLogger,
			URL:      routertest.NewTestResolver(),
			Template: emailTemplates,
		})
		assert.NoError(t, err)
		assert.Empty(t, m.EmailsSent())
	})
}

func TestAuthRoutesChangePasswordWrongPassword(t *testing.T) {
	// Hashed password salted with the id
	id := uuid.MustParse("cae3c6b1-7ff1-4f23-9489-a9f6e82478f9")
	passwordHash := []byte{
		0x24, 0x32, 0x61, 0x24, 0x30, 0x34, 0x24, 0x78, 0x4d, 0x65,
		0x30, 0x54, 0x66, 0x77, 0x4c, 0x75, 0x48, 0x79, 0x35, 0x78,
		0x64, 0x51, 0x76, 0x58, 0x6b, 0x59, 0x73, 0x4b, 0x2e, 0x36,
		0x34, 0x31, 0x70, 0x6c, 0x63, 0x6c, 0x69, 0x54, 0x43, 0x5a,
		0x51, 0x51, 0x55, 0x49, 0x71, 0x41, 0x72, 0x65, 0x77, 0x51,
		0x45, 0x4c, 0x6b, 0x43, 0x76, 0x6d, 0x6a, 0x62, 0x4d, 0x75,
	}
	Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		createdUser := &auth.UsernameUser{
			ID:           id,
			PasswordHash: passwordHash,
		}
		err := model.Save(tx, createdUser)
		assert.NoError(t, err)

		_, err = usernameRoutes.RunChangePassword(&auth.ChangePasswordRequest[*auth.UsernameUser]{
			OldPassword: "wrong password",
			NewPassword: "new password",
			User:        createdUser,
			Ctx:         ctx,
			Update:      dbtest.Update(tx),
		})
		assert.ErrorIs(t, err, auth.Err401Unauthorized)
	})
}

func TestAuthRoutesRefreshErrors(t *testing.T) {
	Run(t, "wrong scope", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()
		createdUser := &auth.UsernameUser{
			ID:           uuid.New(),
			PasswordHash: []byte(""),
		}
		err := model.Save(tx, createdUser)
		assert.NoError(t, err)

		token, err := auth.GenerateToken(&auth.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: createdUser.GetID(),
			},
			Scope: []string{auth.ScopeAccess},
		})
		assert.NoError(t, err)

		_, err = usernameRoutes.RunRefresh(&auth.RefreshRequest[*auth.UsernameUser]{
			RefreshToken: token,
			Ctx:          ctx,
			Read:         dbtest.Read(tx),
		})
		assert.Error(t, err)
	})

	Run(t, "no user found", func(t *testing.T, tx *sqlx.Tx) {
		ctx := context.Background()

		token, err := auth.GenerateToken(&auth.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: uuid.New().String(),
			},
			Scope: []string{auth.ScopeRefresh},
		})
		assert.NoError(t, err)

		_, err = usernameRoutes.RunRefresh(&auth.RefreshRequest[*auth.UsernameUser]{
			RefreshToken: token,
			Ctx:          ctx,
			Read:         dbtest.Read(tx),
		})
		assert.Error(t, err)
	})
}
