package app

import (
	"context"

	"github.com/go-openapi/spec"
	"gosalusa.com/auth"
	"gosalusa.com/clog"
	"gosalusa.com/database"
	"gosalusa.com/email"
	"gosalusa.com/event"
	"gosalusa.com/event/cron"
	"gosalusa.com/filesystem"
	"gosalusa.com/kernel"
	"gosalusa.com/openapidoc"
	"gosalusa.com/openapidoc/openapidocdi"
	"gosalusa.com/pubsub/channelpubsub"
	"gosalusa.com/request"
	"gosalusa.com/static/template/app/events"
	"gosalusa.com/static/template/app/jobs"
	"gosalusa.com/static/template/app/models"
	"gosalusa.com/static/template/app/providers"
	"gosalusa.com/static/template/config"
	"gosalusa.com/static/template/migrations"
	"gosalusa.com/static/template/resources"
	"gosalusa.com/static/template/routes"
	"gosalusa.com/view"
)

var Kernel = kernel.New(
	kernel.Config(config.Load),
	kernel.Bootstrap(
		view.Register(resources.Content, "**/*.html"),
		providers.Register,
		kernel.Register(func(ctx context.Context, c *config.Config) {
			database.Register(ctx, c.Database, migrations.Use())
			email.Register(ctx, c.Mail)
			channelpubsub.Register(ctx)

			clog.RegisterDefault(ctx)
			request.Register(ctx)
			auth.Register[*models.User](ctx)
			event.Register(ctx)
			filesystem.Register(ctx, c.FileSystem)
			openapidocdi.Register(ctx)
		}),
	),
	kernel.Services(
		cron.Service().
			Schedule("* * * * *", &events.LogEvent{Message: "cron event"}),
		event.Service(
			event.NewListener[*jobs.LogJob](),
		),
	),
	kernel.InitRoutes(routes.InitRoutes),
	kernel.APIDocumentation(
		openapidoc.Info(spec.InfoProps{
			Title:       "Salusa Example API",
			Description: `This is the API documentaion for the example Salusa application`,
		}),
		openapidoc.BasePath("/api"),
	),
)
