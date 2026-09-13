package migrate_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/migrate"
	"gosalusa.com/database/model"
	"gosalusa.com/internal/test"
)

func TestRunModelCreate(t *testing.T) {
	test.Run(t, "success", func(t *testing.T, db *sqlx.Tx) {
		type RunModelCreateModel struct {
			model.BaseModel
			ID int `db:"id,primary"`
		}
		err := migrate.RunModelCreate(context.Background(), db, &RunModelCreateModel{})
		assert.NoError(t, err)
	})

	test.Run(t, "error", func(t *testing.T, db *sqlx.Tx) {
		type RunModelCreateErrorModel struct {
			model.BaseModel
			ID    int `db:"id,primary"`
			Value complex128
		}
		err := migrate.RunModelCreate(context.Background(), db, &RunModelCreateErrorModel{})
		assert.Error(t, err)
	})
}
