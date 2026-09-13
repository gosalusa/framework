package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	"gosalusa.com/database"
	"gosalusa.com/database/dialects/sqlite"
	"gosalusa.com/email"
	"gosalusa.com/env"
	"gosalusa.com/filesystem"
)

type Config struct {
	Port     int
	BasePath string

	Database   database.Config
	Mail       email.Config
	FileSystem filesystem.Config
}

func Load() *Config {
	err := godotenv.Load("./.env")
	if errors.Is(err, os.ErrNotExist) {
		// fall through
	} else if err != nil {
		panic(err)
	}

	return &Config{
		Port:     env.Int("PORT", 2303),
		BasePath: env.String("BASE_PATH", ""),
		Database: sqlite.NewConfig(env.String("DATABASE_PATH", "./db.sqlite")),
		Mail: &email.SMTPConfig{
			From:     env.String("MAIL_FROM", "salusa@example.com"),
			Host:     env.String("MAIL_HOST", "sandbox.smtp.mailtrap.io"),
			Port:     env.Int("MAIL_PORT", 2525),
			Username: env.String("MAIL_USERNAME", "user"),
			Password: env.String("MAIL_PASSWORD", "pass"),
		},
		FileSystem: filesystem.NewLocalFS("./files"),
	}
}

func (c *Config) GetHTTPPort() int {
	return c.Port
}
func (c *Config) GetBaseURL() string {
	return c.BasePath
}
