package auth_test

import (
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/auth"
	"gosalusa.com/database"
	"gosalusa.com/database/migrate"
	"gosalusa.com/database/model"
	"gosalusa.com/di"
)

func TestRegister(t *testing.T) {
	t.Run("register autoincrement", func(t *testing.T) {
		db := sqlx.MustOpen("sqlite3", ":memory:")
		defer db.Close()
		migrate.MustMigrateModel(db, &AutoIncrementUser{})
		createdUser := &AutoIncrementUser{
			Username:     "user",
			PasswordHash: []byte{},
		}
		err := model.Save(db, createdUser)
		assert.NoError(t, err)

		ctx := di.TestDependencyProviderContext()
		database.RegisterDB(ctx, db)
		auth.Register[*AutoIncrementUser](ctx)

		ctx = auth.SetClaims(ctx, &auth.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: createdUser.GetID(),
			},
		})

		u, err := di.Resolve[*AutoIncrementUser](ctx)
		assert.NoError(t, err)
		assert.NotNil(t, u)
	})

	t.Run("register uuid", func(t *testing.T) {
		db := sqlx.MustOpen("sqlite3", ":memory:")
		defer db.Close()
		migrate.MustMigrateModel(db, &auth.UsernameUser{})
		createdUser := &auth.UsernameUser{
			ID:           uuid.New(),
			Username:     "user",
			PasswordHash: []byte{},
		}
		err := model.Save(db, createdUser)
		assert.NoError(t, err)

		ctx := di.TestDependencyProviderContext()
		database.RegisterDB(ctx, db)
		auth.Register[*auth.UsernameUser](ctx)

		ctx = auth.SetClaims(ctx, &auth.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: createdUser.GetID(),
			},
		})

		u, err := di.Resolve[*auth.UsernameUser](ctx)
		if assert.NoError(t, err) {
			assert.NotNil(t, u)
		}
	})

	t.Run("no claims", func(t *testing.T) {
		db := sqlx.MustOpen("sqlite3", ":memory:")
		defer db.Close()
		migrate.MustMigrateModel(db, &auth.UsernameUser{})
		createdUser := &auth.UsernameUser{
			ID:           uuid.New(),
			Username:     "user",
			PasswordHash: []byte{},
		}
		err := model.Save(db, createdUser)
		assert.NoError(t, err)

		ctx := di.TestDependencyProviderContext()
		database.RegisterDB(ctx, db)
		auth.Register[*auth.UsernameUser](ctx)

		u, err := di.Resolve[*auth.UsernameUser](ctx)
		assert.ErrorIs(t, err, auth.Err401Unauthorized)
		assert.Nil(t, u)
	})

	t.Run("invalid numeric subject", func(t *testing.T) {
		db := sqlx.MustOpen("sqlite3", ":memory:")
		defer db.Close()
		migrate.MustMigrateModel(db, &AutoIncrementUser{})
		createdUser := &AutoIncrementUser{
			Username:     "user",
			PasswordHash: []byte{},
		}
		err := model.Save(db, createdUser)
		assert.NoError(t, err)

		ctx := di.TestDependencyProviderContext()
		database.RegisterDB(ctx, db)
		auth.Register[*AutoIncrementUser](ctx)

		ctx = auth.SetClaims(ctx, &auth.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "not-a-number",
			},
		})

		u, err := di.Resolve[*AutoIncrementUser](ctx)
		assert.Error(t, err)
		assert.Nil(t, u)
	})

	t.Run("user not found", func(t *testing.T) {
		db := sqlx.MustOpen("sqlite3", ":memory:")
		defer db.Close()
		migrate.MustMigrateModel(db, &auth.UsernameUser{})

		ctx := di.TestDependencyProviderContext()
		database.RegisterDB(ctx, db)
		auth.Register[*auth.UsernameUser](ctx)

		ctx = auth.SetClaims(ctx, &auth.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: uuid.New().String(),
			},
		})

		u, err := di.Resolve[*auth.UsernameUser](ctx)
		assert.ErrorIs(t, err, auth.Err401Unauthorized)
		assert.Nil(t, u)
	})
}
