package routes

import (
	"github.com/google/uuid"
	"gosalusa.com/auth"
	"gosalusa.com/openapidoc"
	"gosalusa.com/request"
	"gosalusa.com/router"
	"gosalusa.com/static/template/app/handlers"
	"gosalusa.com/static/template/app/models"
	"gosalusa.com/view"
)

func InitRoutes(r *router.Router) {
	r.Use(request.HandleErrors())
	r.Use(auth.AttachUser())

	r.Get("/", view.View("index.html", nil)).Name("home")
	r.Get("/login", view.View("login.html", nil)).Name("login")
	r.Get("/user/create", view.View("create_user.html", nil)).Name("user.create")

	r.Handle("/docs", openapidoc.SwaggerUI())

	r.Group("/api", func(r *router.Router) {
		auth.RegisterRoutes(r, auth.NewBasicAuthController[*models.User](
			auth.CreateUser(func(r *auth.EmailVerifiedUserCreateRequest, c *auth.BasicAuthController[*models.User]) (*auth.UserCreateResponse[*models.User], error) {
				return c.RunUserCreate(&models.User{
					EmailVerifiedUser: auth.EmailVerifiedUser{
						ID:           uuid.New(),
						Email:        r.Email,
						PasswordHash: []byte{},
					},
				}, &r.UserCreateRequest)
			}),
			auth.ResetPasswordName("reset-password"),
		))

		r.Get("/user", handlers.UserList)
		r.Get("/user/{id}", handlers.UserGet)
	})
}
