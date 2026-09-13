# Salusa

[![Coverage](https://github.com/abibby/salusa/actions/workflows/coverage.yml/badge.svg)](https://github.com/abibby/salusa/actions/workflows/coverage.yml)
[![Go Reference](https://pkg.go.dev/badge/gosalusa.com.svg)](https://pkg.go.dev/gosalusa.com)

<!-- TODO: add tagline -->

## What is Salusa?

Salusa is a group of utilities to speed up web development in go. They can be
used individually in existing projects or all together with the project
template.

## Packages

### Dependancy Injection

The core of the Salusa framework is DI. It is used to get data into the http
handlers and services you will create in a clean and consistent way.

Resources can be injected directly into Salusa http handlers.

```go
type Request struct {
    DB *sqlx.DB `inject:""`
}
request.Handler(func(r *Request) (*Model, error) {
    return builder.From[*Model]().Find(r.DB, 1)
})
```

And services

```go

type Service struct {
	Logger *slog.Logger `inject:""`
}

func (s *Service) Name() string {
	return "service-name"
}

func (s *Service) Run(ctx context.Context) error {
    s.Logger.Info("log message")
    return nil
}
```

### Database - Builder

The database builder is used to create database queries. It uses a simple
builder to make it easy to create dynamic queries.

```go
model, err := builder.From[*your.Model]().Where("column", "=", "value").First(db)
```

###

## Creating a Salusa Project

Before creating a salusa project you must have the go toolchain installed,
instructions can be found [here](https://go.dev/doc/install). You may also need
node and npm to create frontends, they can be installed with the instructions
found [here](https://nodejs.org/en).

After you have your environment set up you can install the `spice` utility with
go:

```
go install gosalusa.com/spice@latest
```

Once you have installed `spice` you can create a new project with the `init`
command:

```
spice init github.com/abibby/example-app
```

Once you have the created the app you can run the server with the `dev` command:

```
spice dev
```

After starting the development server you can access the application through you
browser at [http://localhost:2303](http://localhost:2303).

# Project Structure

```
root
├ app
│ ├ events
│ ├ handlers
│ ├ jobs
│ └ models
├ config
├ database
├ migrations
├ resources
└ routes
```

## App

The `app` direcoty contains all of the buisness logic of your application. In a
new project the only file in this directory is `kernel.go` which is the core of
the application. The kernel manages all of the long running services in you
application. You can find more information in the
[kernel](https://pkg.go.dev/gosalusa.com/kernel#Kernel) docs

### Events

The `events` directory contains all of the events that you application can emit.

### Handlers

`handlers` contains all of the http handlers in your application. In a standard
application they all implement the `http.Handler` interface. Salusa provides the
`request.Handler` helper for creating APIs that speeds up handling user input
and returning json. Documentation can be found
[here](https://pkg.go.dev/gosalusa.com/request#Handler).

### Jobs

### Models
