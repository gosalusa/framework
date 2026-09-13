package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/kernel"
	"gosalusa.com/router"
	"gosalusa.com/static/template/app"
)

type mainTestConfig struct{}

func (mainTestConfig) GetHTTPPort() int   { return 0 }
func (mainTestConfig) GetBaseURL() string { return "" }

func withArgs(args []string, cb func()) {
	oldArgs := os.Args
	oldCommandLine := pflag.CommandLine
	defer func() {
		os.Args = oldArgs
		pflag.CommandLine = oldCommandLine
	}()
	pflag.CommandLine = pflag.NewFlagSet(oldArgs[0], pflag.ContinueOnError)
	os.Args = append([]string{oldArgs[0]}, args...)
	cb()
}

func TestMainSuccess(t *testing.T) {
	oldKernel := app.Kernel
	defer func() { app.Kernel = oldKernel }()

	app.Kernel = kernel.New(
		kernel.Config(func() mainTestConfig { return mainTestConfig{} }),
		kernel.InitRoutes(func(r *router.Router) {
			r.Get("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
		}),
	)

	withArgs([]string{"--fetch", "/health"}, func() {
		assert.NotPanics(t, func() {
			main()
		})
	})
}

func TestMainBootstrapError(t *testing.T) {
	if os.Getenv("MAIN_TEST_BOOTSTRAP_ERROR") == "1" {
		app.Kernel = kernel.New(
			kernel.InitRoutes(func(r *router.Router) {}),
			kernel.Bootstrap(func(ctx context.Context) error {
				return fmt.Errorf("boom")
			}),
		)
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMainBootstrapError")
	cmd.Env = append(os.Environ(), "MAIN_TEST_BOOTSTRAP_ERROR=1")
	err := cmd.Run()

	var exitErr *exec.ExitError
	if assert.ErrorAs(t, err, &exitErr) {
		assert.Equal(t, 1, exitErr.ExitCode())
	}
}

func TestMainRunError(t *testing.T) {
	if os.Getenv("MAIN_TEST_RUN_ERROR") == "1" {
		app.Kernel = kernel.New(
			kernel.Config(func() mainTestConfig { return mainTestConfig{} }),
			kernel.InitRoutes(func(r *router.Router) {}),
		)
		os.Args = []string{os.Args[0], "--fetch", "http://example.com/%zz"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMainRunError")
	cmd.Env = append(os.Environ(), "MAIN_TEST_RUN_ERROR=1")
	err := cmd.Run()

	var exitErr *exec.ExitError
	if assert.ErrorAs(t, err, &exitErr) {
		assert.Equal(t, 1, exitErr.ExitCode())
	}
}
