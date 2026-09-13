package model_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/database"
	"gosalusa.com/database/builder"
	"gosalusa.com/database/hooks"
	"gosalusa.com/database/model"
	"gosalusa.com/internal/test"
)

type FakeModel struct {
	model.BaseModel
	ID   int    `db:"id,autoincrement"`
	Name string `db:"name"`
}

func TestSave_create(t *testing.T) {
	test.Run(t, "create", func(t *testing.T, tx *sqlx.Tx) {
		f := &test.Foo{
			ID:   1,
			Name: "test",
		}
		err := model.Save(tx, f)
		assert.NoError(t, err)

		rows, err := tx.QueryContext(context.Background(), "select id, name from foos")
		if !assert.NoError(t, err) {
			return
		}
		defer rows.Close()

		assert.True(t, rows.Next())
		id := 0
		name := ""
		err = rows.Scan(&id, &name)
		if assert.NoError(t, err) {

			assert.Equal(t, f.ID, id)
			assert.Equal(t, f.Name, name)

			assert.False(t, rows.Next())
		}
	})

	test.Run(t, "autoincrement", func(t *testing.T, tx *sqlx.Tx) {
		f1 := &test.Foo{
			Name: "test",
		}
		err := model.Save(tx, f1)
		assert.NoError(t, err)
		f2 := &test.Foo{
			Name: "test",
		}
		err = model.Save(tx, f2)
		assert.NoError(t, err)

		rows, err := tx.QueryContext(context.Background(), "select id, name from foos")
		assert.NoError(t, err)

		id := 0
		name := ""
		assert.True(t, rows.Next())
		err = rows.Scan(&id, &name)
		assert.NoError(t, err)

		assert.Equal(t, f1.ID, id)
		assert.Equal(t, f1.Name, name)

		assert.True(t, rows.Next())
		err = rows.Scan(&id, &name)
		assert.NoError(t, err)

		assert.Equal(t, f2.ID, id)
		assert.Equal(t, f2.Name, name)

		assert.False(t, rows.Next())
	})
}

func TestSave_update(t *testing.T) {
	test.Run(t, "update", func(t *testing.T, tx *sqlx.Tx) {
		f := &test.Foo{
			ID:   1,
			Name: "test",
		}
		err := model.Save(tx, f)
		assert.NoError(t, err)

		f.Name = "new name"
		err = model.Save(tx, f)
		assert.NoError(t, err)

		rows, err := tx.QueryContext(context.Background(), "select id, name from foos")
		assert.NoError(t, err)

		assert.True(t, rows.Next())
		id := 0
		name := ""
		err = rows.Scan(&id, &name)
		assert.NoError(t, err)

		assert.Equal(t, f.ID, id)
		assert.Equal(t, f.Name, name)

		assert.False(t, rows.Next())
	})
}

func TestSave_model_is_in_database_after_saving(t *testing.T) {
	test.Run(t, "model in database after saving", func(t *testing.T, tx *sqlx.Tx) {
		f := &test.Foo{
			ID: 1,
		}
		err := model.Save(tx, f)
		assert.NoError(t, err)

		assert.True(t, f.InDatabase())
	})
}

func TestSave_autoincrement(t *testing.T) {
	test.Run(t, "autoincrement", func(t *testing.T, tx *sqlx.Tx) {
		f := &test.Foo{}
		err := model.Save(tx, f)
		assert.NoError(t, err)

		assert.NotEqual(t, f.ID, 0)
	})
	test.Run(t, "autoincrement set id", func(t *testing.T, tx *sqlx.Tx) {
		f := &test.Foo{
			ID: 100,
		}
		err := model.Save(tx, f)
		assert.NoError(t, err)

		assert.Equal(t, f.ID, 100)
	})
}

type FooSaveHookTest struct {
	test.Foo
	saved bool
}

type FooSaveHookTestWrapper struct {
	FooSaveHookTest
}

var _ hooks.AfterSaver = &FooSaveHookTest{}

func (f *FooSaveHookTest) AfterSave(context.Context, database.DB) error {
	f.saved = true
	return nil
}
func (f *FooSaveHookTest) Table() string {
	return "foos"
}

func TestSave_hooks(t *testing.T) {
	test.Run(t, "runs hooks", func(t *testing.T, tx *sqlx.Tx) {
		f := &FooSaveHookTest{
			Foo: test.Foo{
				ID: 1,
			},
		}
		err := model.Save(tx, f)
		assert.NoError(t, err)

		assert.True(t, f.saved)
	})

	test.Run(t, "runs hooks on anonymise structs", func(t *testing.T, tx *sqlx.Tx) {
		f := &FooSaveHookTestWrapper{
			FooSaveHookTest{
				Foo: test.Foo{
					ID: 1,
				},
			},
		}
		err := model.Save(tx, f)
		assert.NoError(t, err)

		assert.True(t, f.saved)
	})
}

type SaveFooReadonly struct {
	test.Foo
	Readonly string `db:"computed,readonly"`
}

func (f *SaveFooReadonly) Table() string {
	return "foos"
}

func TestSave_readonly(t *testing.T) {
	test.Run(t, "runs hooks", func(t *testing.T, tx *sqlx.Tx) {
		f := &SaveFooReadonly{
			Foo: test.Foo{
				ID: 1,
			},
			Readonly: "yes",
		}
		err := model.Save(tx, f)
		assert.NoError(t, err)

		newFoo, err := builder.From[*SaveFooReadonly]().First(tx)
		assert.NoError(t, err)
		assert.Equal(t, "", newFoo.Readonly)
	})

}

func TestInsertManyContext(t *testing.T) {
	test.Run(t, "insert", func(t *testing.T, tx *sqlx.Tx) {
		models := []*test.Foo{{Name: "1"}, {Name: "2"}, {Name: "3"}}
		err := model.InsertManyContext(context.TODO(), tx, models)
		if !assert.NoError(t, err) {
			return
		}

		for _, m := range models {
			assert.NotZero(t, m.ID)
		}

		foos, err := builder.From[*SaveFooReadonly]().OrderBy("id").Get(tx)
		if !assert.NoError(t, err) {
			return
		}

		assert.Len(t, foos, 3)
		for i, foo := range foos {
			assert.True(t, foo.InDatabase())
			assert.Equal(t, fmt.Sprint(i+1), foo.Name)
			assert.Equal(t, models[i].ID, foo.ID)
			assert.Equal(t, models[i].Name, foo.Name)
		}
	})

	test.Run(t, "insert many non autoincrement", func(t *testing.T, tx *sqlx.Tx) {
		models := []*test.Foo{{ID: 1, Name: "1"}, {ID: 2, Name: "2"}}
		err := model.InsertManyContext(context.TODO(), tx, models)
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, 1, models[0].ID)
		assert.Equal(t, 2, models[1].ID)

		foos, err := builder.From[*test.Foo]().OrderBy("id").Get(tx)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, foos, 2)
		assert.Equal(t, 1, foos[0].ID)
		assert.Equal(t, 2, foos[1].ID)
	})

	test.Run(t, "no models", func(t *testing.T, tx *sqlx.Tx) {
		err := model.InsertManyContext(context.TODO(), tx, []*test.Foo{})
		assert.NoError(t, err)
	})

	test.Run(t, "before save hook error", func(t *testing.T, tx *sqlx.Tx) {
		models := []*FooBeforeSaveErrorTest{{}, {}}
		err := model.InsertManyContext(context.TODO(), tx, models)

		assert.ErrorContains(t, err, "before save hooks")
		assert.ErrorContains(t, err, "before save error")
	})

	test.Run(t, "autoincrement returning query error", func(t *testing.T, tx *sqlx.Tx) {
		models := []*FakeModel{{Name: "1"}, {Name: "2"}}
		err := model.InsertManyContext(context.TODO(), tx, models)

		assert.ErrorContains(t, err, "insert:")
		assert.ErrorContains(t, err, "failed to insert model")
	})

	test.Run(t, "non autoincrement exec error", func(t *testing.T, tx *sqlx.Tx) {
		models := []*FakeModel{{ID: 1, Name: "1"}, {ID: 2, Name: "2"}}
		err := model.InsertManyContext(context.TODO(), tx, models)

		assert.ErrorContains(t, err, "insert:")
		assert.ErrorContains(t, err, "failed to insert model")
	})
}

func TestInsertMany(t *testing.T) {
	test.Run(t, "insert many", func(t *testing.T, tx *sqlx.Tx) {
		models := []*test.Foo{{Name: "1"}, {Name: "2"}}
		err := model.InsertMany(tx, models)
		if !assert.NoError(t, err) {
			return
		}

		for _, m := range models {
			assert.NotZero(t, m.ID)
		}
	})
}

func TestMustSave(t *testing.T) {
	test.Run(t, "success", func(t *testing.T, tx *sqlx.Tx) {
		f := &test.Foo{
			Name: "test",
		}
		model.MustSave(tx, f)

		assert.True(t, f.InDatabase())
	})
	test.Run(t, "panics on error", func(t *testing.T, tx *sqlx.Tx) {
		assert.Panics(t, func() {
			model.MustSave(tx, &FakeModel{Name: "test"})
		})
	})
}

func TestMustSaveContext(t *testing.T) {
	test.Run(t, "success", func(t *testing.T, tx *sqlx.Tx) {
		f := &test.Foo{
			Name: "test",
		}
		model.MustSaveContext(context.Background(), tx, f)

		assert.True(t, f.InDatabase())
	})
	test.Run(t, "panics on error", func(t *testing.T, tx *sqlx.Tx) {
		assert.Panics(t, func() {
			model.MustSaveContext(context.Background(), tx, &FakeModel{Name: "test"})
		})
	})
}

func TestSave_before_save_hook_error(t *testing.T) {
	test.Run(t, "before save hook error", func(t *testing.T, tx *sqlx.Tx) {
		f := &FooBeforeSaveErrorTest{
			Foo: test.Foo{
				Name: "test",
			},
		}
		err := model.Save(tx, f)

		assert.ErrorContains(t, err, "before save hooks")
		assert.ErrorContains(t, err, "before save error")
	})
}

func TestSave_insert_error(t *testing.T) {
	test.Run(t, "autoincrement returning query error", func(t *testing.T, tx *sqlx.Tx) {
		err := model.Save(tx, &FakeModel{Name: "test"})

		assert.ErrorContains(t, err, "insert:")
		assert.ErrorContains(t, err, "failed to insert model")
	})
	test.Run(t, "non autoincrement exec error", func(t *testing.T, tx *sqlx.Tx) {
		err := model.Save(tx, &FakeModel{ID: 100, Name: "test"})

		assert.ErrorContains(t, err, "insert:")
		assert.ErrorContains(t, err, "failed to insert model")
	})
}

func TestSave_update_error(t *testing.T) {
	test.Run(t, "exec error", func(t *testing.T, tx *sqlx.Tx) {
		type Foo struct {
			test.Foo
			NotName string `db:"not_name"`
		}

		f := &test.Foo{
			ID:   1,
			Name: "test",
		}
		err := model.Save(tx, f)
		assert.NoError(t, err)

		err = model.Save(tx, &Foo{
			Foo:     *f,
			NotName: "anything",
		})

		assert.ErrorContains(t, err, "update:")
	})
	test.Run(t, "no primary key found", func(t *testing.T, tx *sqlx.Tx) {
		f := &SaveUpdateNoPrimaryKeyTest{
			Foo: test.Foo{
				Name: "test",
			},
		}
		err := model.Save(tx, f)

		assert.ErrorContains(t, err, "update:")
		assert.ErrorContains(t, err, "no primary key found")
	})
}

func TestSave_after_save_hook_error(t *testing.T) {
	test.Run(t, "after save hook error", func(t *testing.T, tx *sqlx.Tx) {
		f := &FooAfterSaveErrorTest{
			Foo: test.Foo{
				Name: "test",
			},
		}
		err := model.Save(tx, f)

		assert.ErrorContains(t, err, "after save hooks")
		assert.ErrorContains(t, err, "after save error")
	})
}

type FooBeforeSaveErrorTest struct {
	test.Foo
}

var _ hooks.BeforeSaver = &FooBeforeSaveErrorTest{}

func (f *FooBeforeSaveErrorTest) BeforeSave(context.Context, database.DB) error {
	return fmt.Errorf("before save error")
}
func (f *FooBeforeSaveErrorTest) Table() string {
	return "foos"
}

type FooAfterSaveErrorTest struct {
	test.Foo
}

var _ hooks.AfterSaver = &FooAfterSaveErrorTest{}

func (f *FooAfterSaveErrorTest) AfterSave(context.Context, database.DB) error {
	return fmt.Errorf("after save error")
}
func (f *FooAfterSaveErrorTest) Table() string {
	return "foos"
}

type SaveUpdateNoPrimaryKeyTest struct {
	test.Foo
}

func (f *SaveUpdateNoPrimaryKeyTest) InDatabase() bool {
	return true
}
func (f *SaveUpdateNoPrimaryKeyTest) PrimaryKey() []string {
	return []string{"missing_key"}
}
func (f *SaveUpdateNoPrimaryKeyTest) Table() string {
	return "foos"
}
