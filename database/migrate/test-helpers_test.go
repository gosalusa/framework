package migrate_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/migrate"
	"gosalusa.com/database/model"
	"gosalusa.com/internal/test"
)

func TestMustMigrateModel(t *testing.T) {
	test.Run(t, "success", func(t *testing.T, db *sqlx.Tx) {
		type MustMigrateModelModel struct {
			model.BaseModel
			ID int `db:"id,primary"`
		}
		assert.NotPanics(t, func() {
			migrate.MustMigrateModel(db, &MustMigrateModelModel{})
		})
	})

	test.Run(t, "panics on invalid field", func(t *testing.T, db *sqlx.Tx) {
		type MustMigrateModelErrorModel struct {
			model.BaseModel
			ID    int `db:"id,primary"`
			Value complex128
		}
		assert.Panics(t, func() {
			migrate.MustMigrateModel(db, &MustMigrateModelErrorModel{})
		})
	})
}
