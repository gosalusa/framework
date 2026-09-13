package migrate_test

import (
	"testing"
	"time"

	"github.com/abibby/nulls"
	"github.com/bradleyjkemp/cupaloy"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/builder"
	"gosalusa.com/database/dialects"
	"gosalusa.com/database/migrate"
	"gosalusa.com/database/model"
	"gosalusa.com/database/schema"
)

type Date int

func (d Date) DataType() dialects.DataType {
	return dialects.DataTypeDate
}

func TestGenerateMigration(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		type TestModel struct {
			model.BaseModel
			ID       int           `db:"id,primary"`
			Nullable *nulls.String `db:"nullable"`
			Indexed  bool          `db:"indexed,index"`
		}
		src, err := migrate.New().GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, src)
	})

	t.Run("add column", func(t *testing.T) {
		m := migrate.New()
		m.Add(&migrate.Migration{
			Name: "2023-01-01T00:00:00Z create test model",
			Up: schema.Create("test_models", func(table *schema.Blueprint) {
				table.Int("id").Primary()
			}),
			Down: nil,
		})

		type TestModel struct {
			model.BaseModel
			ID    int    `db:"id,primary"`
			ToAdd string `db:"to_add"`
		}
		src, err := m.GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, src)
	})

	t.Run("drop column", func(t *testing.T) {
		m := migrate.New()
		m.Add(&migrate.Migration{
			Name: "2023-01-01T00:00:00Z create test model",
			Up: schema.Create("test_models", func(table *schema.Blueprint) {
				table.Int("id").Primary()
				table.String("to_drop")
			}),
			Down: nil,
		})

		type TestModel struct {
			model.BaseModel
			ID int `db:"id,primary"`
		}
		src, err := m.GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, src)
	})

	t.Run("change", func(t *testing.T) {
		m := migrate.New()
		m.Add(&migrate.Migration{
			Name: "2023-01-01T00:00:00Z create test model",
			Up: schema.Create("test_models", func(table *schema.Blueprint) {
				table.Int("id").Primary()
			}),
			Down: nil,
		})

		type TestModel struct {
			model.BaseModel
			ID string `db:"id,primary"`
		}
		src, err := m.GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, src)
	})

	t.Run("no changes", func(t *testing.T) {
		m := migrate.New()
		m.Add(&migrate.Migration{
			Name: "2023-01-01T00:00:00Z create test model",
			Up: schema.Create("test_models", func(table *schema.Blueprint) {
				table.Int("id").Primary()
			}),
			Down: nil,
		})

		type TestModel struct {
			model.BaseModel
			ID int `db:"id,primary"`
		}
		_, err := m.GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.ErrorIs(t, err, migrate.ErrNoChanges)
	})

	t.Run("multiple migrations", func(t *testing.T) {
		m := migrate.New()
		m.Add(&migrate.Migration{
			Name: "2023-01-01T00:00:00Z create test model",
			Up: schema.Create("test_models", func(table *schema.Blueprint) {
				table.Int("id").Primary()
			}),
			Down: nil,
		})
		m.Add(&migrate.Migration{
			Name: "2023-01-01T00:00:01Z change",
			Up: schema.Table("test_models", func(table *schema.Blueprint) {
				table.String("id").Primary().Change()
			}),
			Down: nil,
		})

		type TestModel struct {
			model.BaseModel
			ID int `db:"id,primary"`
		}
		src, err := m.GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, src)
	})

	t.Run("ignore - fields", func(t *testing.T) {
		type TestModel struct {
			model.BaseModel
			ID     int `db:"id,primary"`
			Ignore int `db:"-"`
		}
		src, err := migrate.New().GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, src)
	})

	t.Run("dates", func(t *testing.T) {
		type TestModel struct {
			model.BaseModel
			ID   int       `db:"id,primary"`
			Time time.Time `db:"time"`
		}
		src, err := migrate.New().GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, src)
	})

	t.Run("uuid", func(t *testing.T) {
		type TestModel struct {
			model.BaseModel
			ID   uuid.UUID `db:"id,primary"`
			Time time.Time `db:"time"`
		}
		src, err := migrate.New().GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, src)
	})

	t.Run("custom type", func(t *testing.T) {
		type TestModel struct {
			model.BaseModel
			ID int `db:"id,primary,type:date"`
		}
		src, err := migrate.New().GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, src)
	})

	t.Run("custom type DataTyper", func(t *testing.T) {
		type TestModel struct {
			model.BaseModel
			ID *Date `db:"id,primary"`
		}
		src, err := migrate.New().GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, src)
	})

	t.Run("custom type DataTyper not pointer", func(t *testing.T) {
		type TestModel struct {
			model.BaseModel
			ID Date `db:"id,primary"`
		}
		src, err := migrate.New().GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, src)
	})

	t.Run("relationships", func(t *testing.T) {
		type RelatedModel struct {
			model.BaseModel
			ID int
		}
		type TestModel struct {
			model.BaseModel
			Related *builder.BelongsTo[*RelatedModel]
		}
		src, err := migrate.New().GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, src)
	})

	t.Run("add relationship", func(t *testing.T) {
		type RelatedModel struct {
			model.BaseModel
			ID int
		}
		type TestModel struct {
			model.BaseModel
			Related *builder.BelongsTo[*RelatedModel]
		}
		m := migrate.New()
		m.Add(&migrate.Migration{
			Name: "2023-01-01T00:00:00Z create test model",
			Up: schema.Create("test_models", func(table *schema.Blueprint) {
			}),
			Down: schema.DropIfExists("test_models"),
		})
		src, err := m.GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, src)
	})

	t.Run("composite primary key", func(t *testing.T) {
		type TestModel struct {
			model.BaseModel
			ID1 int `db:"id1,primary"`
			ID2 int `db:"id2,primary"`
		}
		src, err := migrate.New().GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, src)
	})

	t.Run("readonly", func(t *testing.T) {
		type TestModel struct {
			model.BaseModel
			ID       int    `db:"id,primary"`
			Readonly string `db:"readonly,readonly"`
		}
		src, err := migrate.New().GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, src)
	})

	t.Run("nullable", func(t *testing.T) {
		type TestModel struct {
			model.BaseModel
			ID       int    `db:"id,primary"`
			Nullable string `db:"nullable,nullable"`
		}
		src, err := migrate.New().GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, src)
	})
}
