package auth

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/go-openapi/spec"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"gosalusa.com/database"
	"gosalusa.com/database/builder"
	"gosalusa.com/database/model"
	"gosalusa.com/email"
	"gosalusa.com/internal/helpers"
	"gosalusa.com/request"
	"gosalusa.com/router"
	"gosalusa.com/view"
)

var (
	ErrInvalidUserPass      = errors.New("invalid username or password")
	ErrTokenNotFound        = errors.New("token not found")
	ErrNonEmailVerifiedUser = errors.New("non email verified user")
)

type AuthController interface {
	UserCreate() http.Handler
	Login() http.Handler
	VerifyEmail() http.Handler
	ResetPassword() http.Handler
	ChangePassword() http.Handler
	Refresh() http.Handler
	ForgotPassword() http.Handler
}

type BasicAuthController[T User] struct {
	basicAuthController
}

type basicAuthController struct {
	createUserHandler   func(controller any) http.Handler
	resetPasswordName   string
	accessTokenOptions  func(u any, claims *Claims) jwt.Claims
	refreshTokenOptions func(u any, claims *Claims) jwt.Claims
}

type AuthOption func(a *basicAuthController) *basicAuthController

type CreateUserFunc[T User] func(user T, req *UserCreateRequest) (*UserCreateResponse[T], error)

func CreateUser[T User, TRequest any](cb func(r *TRequest, c *BasicAuthController[T]) (*UserCreateResponse[T], error)) AuthOption {
	return func(a *basicAuthController) *basicAuthController {
		a.createUserHandler = func(controller any) http.Handler {
			return request.Handler(func(r *TRequest) (*UserCreateResponse[T], error) {
				return cb(r, controller.(*BasicAuthController[T]))
			})
		}
		return a
	}
}

func AccessTokenOptions[T User](cb func(u T, claims *Claims) jwt.Claims) AuthOption {
	return func(a *basicAuthController) *basicAuthController {
		a.accessTokenOptions = func(u any, claims *Claims) jwt.Claims {
			user, _ := u.(T)
			return cb(user, claims)
		}
		return a
	}
}
func RefreshTokenOptions[T User](cb func(u T, claims *Claims) jwt.Claims) AuthOption {
	return func(a *basicAuthController) *basicAuthController {
		a.refreshTokenOptions = func(u any, claims *Claims) jwt.Claims {
			user, _ := u.(T)
			return cb(user, claims)
		}
		return a
	}
}
func ResetPasswordName(name string) AuthOption {
	return func(a *basicAuthController) *basicAuthController {
		a.resetPasswordName = name
		return a
	}
}

//go:embed emails/*
var emails embed.FS
var defaultViewTemplate = view.NewViewTemplate(emails, "**/*.html")

func NewBasicAuthController[T User](options ...AuthOption) *BasicAuthController[T] {
	core := &basicAuthController{
		createUserHandler: func(controller any) http.Handler {
			return nil
		},
		resetPasswordName: "reset-password",
		accessTokenOptions: func(u any, claims *Claims) jwt.Claims {
			return claims
		},
		refreshTokenOptions: func(u any, claims *Claims) jwt.Claims {
			return claims
		},
	}

	for _, opt := range options {
		core = opt(core)
	}
	return &BasicAuthController[T]{
		basicAuthController: *core,
	}
}

func RegisterRoutes(r *router.Router, controller AuthController) {
	r.Post("/login", controller.Login()).Name("auth.login")
	r.Post("/user/password/reset", controller.ResetPassword()).Name("auth.password.reset")
	r.Post("/user/password/forgot", controller.ForgotPassword()).Name("auth.password.forgot")
	r.Post("/user", controller.UserCreate()).Name("auth.user.create")
	r.Get("/user/verify", controller.VerifyEmail()).Name("auth.email.verify")
	r.Post("/login/refresh", controller.Refresh()).Name("auth.refresh")

	r.Group("", func(r *router.Router) {
		r.Use(AttachUser())
		r.Use(LoggedIn())

		r.Post("/user/password/change", controller.ChangePassword()).Name("auth.password.change")
	})
}

type UserCreateRequest struct {
	Password string `json:"password" validate:"required"`

	Mailer   email.Mailer       `inject:""`
	Update   database.Update    `inject:""`
	Ctx      context.Context    `inject:""`
	Logger   *slog.Logger       `inject:""`
	URL      router.URLResolver `inject:""`
	Template *view.ViewTemplate `inject:",optional"`
}
type UserCreateResponse[T User] struct {
	User T `json:"user"`
}

func (o *BasicAuthController[T]) UserCreate() http.Handler {
	h := o.createUserHandler(o)
	if docser, ok := h.(interface{ Docs(*spec.OperationProps) }); ok {
		docser.Docs(&spec.OperationProps{
			Tags: []string{"auth"},
		})
	}
	return h
}
func (o *BasicAuthController[T]) RunUserCreate(u T, r *UserCreateRequest) (*UserCreateResponse[T], error) {
	for _, col := range u.UsernameColumns() {
		v, err := helpers.RGetValue(reflect.ValueOf(u), col)
		if err != nil {
			return nil, err
		}
		if v.Kind() == reflect.String {
			v.Set(reflect.ValueOf(strings.ToLower(v.String())))
		}
	}

	err := r.Update(func(tx *sqlx.Tx) error {
		err := model.SaveContext(r.Ctx, tx, u)
		if err != nil {
			return err
		}

		err = updatePassword(u, r.Password)
		if err != nil {
			return err
		}

		if v, ok := cast[EmailVerified](u); ok {
			o.sendEmails(&sendEmailOptions{
				URL:          r.URL,
				ViewTemplate: r.Template,
				User:         v,
				Mailer:       r.Mailer,
				Logger:       r.Logger,
				TemplateName: "verify_email.html",
				Subject:      "Verify your email",
			})
		}

		return model.SaveContext(r.Ctx, tx, u)
	})
	if err != nil {
		return nil, err
	}

	return &UserCreateResponse[T]{
		User: u,
	}, nil
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`

	Ctx  context.Context `inject:""`
	Read database.Read   `inject:""`
	Log  *slog.Logger    `inject:""`
}
type LoginResponse struct {
	AccessToken  string `json:"token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh"`
	ExpiresIn    int    `json:"expires_in"`
}

func (o *BasicAuthController[T]) Login() http.Handler {
	return request.Handler(o.RunLogin).Docs(&spec.OperationProps{
		Tags: []string{"auth"},
	})
}
func (o *BasicAuthController[T]) RunLogin(r *LoginRequest) (*LoginResponse, error) {
	u, err := helpers.NewOf[T]()
	if err != nil {
		return nil, err
	}
	userColumns := u.UsernameColumns()
	if len(userColumns) == 0 {
		panic("need columns")
	}

	q := builder.From[T]().WithContext(r.Ctx)
	for _, column := range userColumns {
		q = q.OrWhere(column, "=", strings.ToLower(r.Username))
	}

	u, err = database.Value(r.Read, func(tx *sqlx.Tx) (T, error) {
		return q.First(tx)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to log in: %w", err)
	}
	if reflect.ValueOf(u).IsNil() {
		r.Log.Info("login attempt with unknown username", "username", r.Username)
		return nil, request.NewHTTPError(ErrInvalidUserPass, http.StatusUnauthorized)
	}

	if v, ok := cast[EmailVerified](u); ok {
		if !v.IsVerified() {
			return nil, Err401Unauthorized
		}
	}

	err = bcrypt.CompareHashAndPassword(u.GetPasswordHash(), u.SaltedPassword(r.Password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		r.Log.Info("login attempt with incorrect password", "username", r.Username, "user_id", u.GetID())
		return nil, request.NewHTTPError(ErrInvalidUserPass, http.StatusUnauthorized)
	} else if err != nil {
		return nil, fmt.Errorf("could not check password hash: %w", err)
	}

	expires := time.Hour

	access, err := GenerateToken(
		o.accessTokenOptions(u, NewClaims().
			WithSubject(u.GetID()).
			WithLifetime(expires).
			WithScopes(ScopeAccess),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("could not generate token: %w", err)
	}
	refresh, err := GenerateToken(
		o.refreshTokenOptions(u, NewClaims().
			WithSubject(u.GetID()).
			WithLifetime(time.Hour*24*30).
			WithScopes(ScopeRefresh),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("could not generate refresh: %w", err)
	}

	return &LoginResponse{
		AccessToken:  access,
		TokenType:    "Bearer",
		RefreshToken: refresh,
		ExpiresIn:    int(expires.Seconds()),
	}, nil
}

type VerifyEmailRequest struct {
	Token  string             `query:"token"`
	Update database.Update    `inject:""`
	Ctx    context.Context    `inject:""`
	URL    router.URLResolver `inject:""`
}

func (o *BasicAuthController[T]) VerifyEmail() http.Handler {
	return request.Handler(o.RunVerifyEmail).Docs(&spec.OperationProps{
		Tags: []string{"auth"},
	})
}
func (o *BasicAuthController[T]) RunVerifyEmail(r *VerifyEmailRequest) (http.Handler, error) {
	v, err := helpers.NewOf[T]()
	if err != nil {
		return nil, err
	}
	emptyValidated, ok := cast[EmailVerified](v)
	if !ok {
		return nil, request.NewHTTPError(ErrNonEmailVerifiedUser, http.StatusUnauthorized)
	}

	err = r.Update(func(tx *sqlx.Tx) error {
		u, err := builder.From[T]().
			WithContext(r.Ctx).
			Where(emptyValidated.LookupTokenColumn(), "=", r.Token).
			First(tx)
		if err != nil {
			return fmt.Errorf("failed to verify: %w", err)
		}
		if reflect.ValueOf(u).IsNil() {
			return request.NewHTTPError(ErrTokenNotFound, http.StatusUnauthorized)
		}

		v := mustCast[EmailVerified](u)
		if v.IsVerified() {
			return request.NewHTTPError(fmt.Errorf("already verified"), http.StatusUnauthorized)
		}

		v.SetVerified(true)
		v.SetLookupToken("")

		return model.SaveContext(r.Ctx, tx, u)
	})
	if err != nil {
		return nil, err
	}

	return http.RedirectHandler(r.URL.Resolve("login"), http.StatusFound), nil
}

type ResetPasswordRequest struct {
	Token    string             `json:"token" validate:"required|min:1"`
	Password string             `json:"password" validate:"required"`
	Update   database.Update    `inject:""`
	Ctx      context.Context    `inject:""`
	URL      router.URLResolver `inject:""`
}
type ResetPasswordResponse[T User] struct {
	User T `json:"user"`
}

func (o *BasicAuthController[T]) ResetPassword() http.Handler {
	return request.Handler(o.RunResetPassword).Docs(&spec.OperationProps{
		Tags: []string{"auth"},
	})
}
func (o *BasicAuthController[T]) RunResetPassword(r *ResetPasswordRequest) (*ResetPasswordResponse[T], error) {
	u, err := helpers.NewOf[T]()
	if err != nil {
		return nil, err
	}
	zeroValidated, ok := cast[EmailVerified](u)
	if !ok {
		return nil, ErrNonEmailVerifiedUser
	}
	err = r.Update(func(tx *sqlx.Tx) error {
		u, err = builder.From[T]().
			WithContext(r.Ctx).
			Where(zeroValidated.LookupTokenColumn(), "=", r.Token).
			First(tx)
		if err != nil {
			return fmt.Errorf("failed to verify: %w", err)
		}
		if reflect.ValueOf(u).IsNil() {
			return request.NewHTTPError(ErrTokenNotFound, http.StatusUnauthorized)
		}

		v := mustCast[EmailVerified](u)
		if !v.IsVerified() {
			return Err401Unauthorized
		}

		v.SetLookupToken("")

		err = updatePassword(u, r.Password)
		if err != nil {
			return err
		}

		return model.SaveContext(r.Ctx, tx, u)

	})
	if err != nil {
		return nil, err
	}
	return &ResetPasswordResponse[T]{
		User: u,
	}, nil
}

type ForgotPasswordRequest struct {
	Email    string             `json:"email" validate:"required|email"`
	Update   database.Update    `inject:""`
	Ctx      context.Context    `inject:""`
	Mailer   email.Mailer       `inject:""`
	Logger   *slog.Logger       `inject:""`
	URL      router.URLResolver `inject:""`
	Template *view.ViewTemplate `inject:",optional"`
}
type ForgotPasswordResponse struct {
}

func (o *BasicAuthController[T]) ForgotPassword() http.Handler {
	return request.Handler(o.RunForgotPassword).Docs(&spec.OperationProps{
		Tags: []string{"auth"},
	})
}
func (o *BasicAuthController[T]) RunForgotPassword(r *ForgotPasswordRequest) (*ForgotPasswordResponse, error) {
	u, err := helpers.NewOf[T]()
	if err != nil {
		return nil, err
	}
	_, ok := cast[EmailVerified](u)
	if !ok {
		panic("not email verified")
	}
	userColumns := u.UsernameColumns()
	if len(userColumns) == 0 {
		panic("need columns")
	}
	err = r.Update(func(tx *sqlx.Tx) error {
		q := builder.From[T]().WithContext(r.Ctx)
		for _, column := range userColumns {
			q = q.OrWhere(column, "=", strings.ToLower(r.Email))
		}

		u, err := q.First(tx)
		if err != nil {
			return err
		}
		if reflect.ValueOf(u).IsZero() {
			r.Logger.Info("password reset attempt for unused email", slog.String("email", r.Email))
			return nil
		}

		o.sendEmails(&sendEmailOptions{
			URL:          r.URL,
			ViewTemplate: r.Template,
			User:         mustCast[EmailVerified](u),
			Mailer:       r.Mailer,
			Logger:       r.Logger,
			TemplateName: "reset_password.html",
			Subject:      "Password reset",
		})
		return model.SaveContext(r.Ctx, tx, u)
	})
	if err != nil {
		return nil, err
	}
	return &ForgotPasswordResponse{}, err
}

type ChangePasswordRequest[T User] struct {
	OldPassword string          `json:"old_password"`
	NewPassword string          `json:"new_password"`
	User        T               `inject:""`
	Update      database.Update `inject:""`
	Ctx         context.Context `inject:""`
}
type ChangePasswordResponse[T User] struct {
	User T `json:"user"`
}

func (o *BasicAuthController[T]) ChangePassword() http.Handler {
	return request.Handler(o.RunChangePassword).Docs(&spec.OperationProps{
		Tags: []string{"auth"},
	})
}
func (o *BasicAuthController[T]) RunChangePassword(r *ChangePasswordRequest[T]) (*ChangePasswordResponse[T], error) {
	err := bcrypt.CompareHashAndPassword(r.User.GetPasswordHash(), r.User.SaltedPassword(r.OldPassword))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return nil, Err401Unauthorized
	} else if err != nil {
		return nil, fmt.Errorf("could not check password hash: %w", err)
	}

	err = updatePassword(r.User, r.NewPassword)
	if err != nil {
		return nil, err
	}

	err = r.Update(func(tx *sqlx.Tx) error {
		return model.SaveContext(r.Ctx, tx, r.User)
	})
	if err != nil {
		return nil, err
	}

	return &ChangePasswordResponse[T]{
		User: r.User,
	}, nil
}

type RefreshRequest[T User] struct {
	RefreshToken string          `json:"refresh"`
	Read         database.Read   `inject:""`
	Ctx          context.Context `inject:""`
}

func (o *BasicAuthController[T]) Refresh() http.Handler {
	return request.Handler(o.RunRefresh).Docs(&spec.OperationProps{
		Tags: []string{"auth"},
	})
}
func (o *BasicAuthController[T]) RunRefresh(r *RefreshRequest[T]) (*LoginResponse, error) {
	claims, err := Parse(r.RefreshToken)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(claims.Scope, ScopeRefresh) {
		return nil, request.NewHTTPError(fmt.Errorf("invalid token"), http.StatusUnauthorized)
	}

	var u T
	err = r.Read(func(tx *sqlx.Tx) error {
		u, err = builder.From[T]().
			WithContext(r.Ctx).
			Find(tx, claims.Subject)
		return err
	})
	if err != nil {
		return nil, request.NewHTTPError(fmt.Errorf("failed to verify: %w", err), http.StatusUnauthorized)
	}

	if reflect.ValueOf(u).IsNil() {
		return nil, request.NewHTTPError(fmt.Errorf("no user found"), http.StatusUnauthorized)
	}

	expires := time.Hour

	access, err := GenerateToken(
		o.accessTokenOptions(u, NewClaims().
			WithSubject(u.GetID()).
			WithLifetime(expires).
			WithIssuedAtTime(time.Now()).
			WithScopes(ScopeAccess),
		),
	)
	if err != nil {
		return nil, request.NewHTTPError(fmt.Errorf("could not generate token: %w", err), http.StatusUnauthorized)
	}

	return &LoginResponse{
		AccessToken:  access,
		TokenType:    "Bearer",
		RefreshToken: r.RefreshToken,
		ExpiresIn:    int(expires.Seconds()),
	}, nil
}

func updatePassword(u User, password string) error {
	hash, err := bcrypt.GenerateFromPassword(u.SaltedPassword(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.SetPasswordHash(hash)

	return nil
}

type sendEmailOptions struct {
	URL          router.URLResolver
	ViewTemplate *view.ViewTemplate
	User         EmailVerified
	Mailer       email.Mailer
	Logger       *slog.Logger
	TemplateName string
	Subject      string
}

func (o *BasicAuthController[T]) sendEmails(opt *sendEmailOptions) {
	token := uuid.New().String()

	opt.User.SetLookupToken(token)

	go func() {
		if opt.ViewTemplate == nil {
			opt.ViewTemplate = defaultViewTemplate
		}

		b, err := view.View(opt.TemplateName, map[string]any{
			"ResetPasswordName": o.resetPasswordName,
			"Token":             token,
		}).BytesData(&view.ViewData{
			URL:      opt.URL,
			Template: opt.ViewTemplate,
		})
		if err != nil {
			opt.Logger.Warn("failed to generate auth email body",
				"template", opt.TemplateName,
				"error", err,
			)
			return
		}
		err = opt.Mailer.Mail(&email.Message{
			To:       []string{opt.User.GetEmail()},
			Subject:  opt.Subject,
			HTMLBody: string(b),
		})
		if err != nil {
			opt.Logger.Warn("failed to send auth email",
				"email", opt.User.GetEmail(),
				"subject", opt.Subject,
				"error", err,
			)
			return
		}
		opt.Logger.Info("auth email sent successfully",
			"email", opt.User.GetEmail(),
			"subject", opt.Subject,
		)
	}()
}

func cast[T any](v any) (T, bool) {
	t, ok := v.(T)
	return t, ok
}
func mustCast[T any](v any) T {
	return v.(T)
}
