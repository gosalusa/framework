package database_test

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/database"
	"gosalusa.com/database/dialects/sqlite"
)

type fakeDB struct {
	err error
}

func (f *fakeDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, f.err
}
func (f *fakeDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, f.err
}
func (f *fakeDB) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	return nil, f.err
}
func (f *fakeDB) QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row {
	return nil
}
func (f *fakeDB) DriverName() string {
	return "fake_db"
}

func newTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	sqlite.UseSQLite()
	db, err := sqlx.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})
	_, err = db.ExecContext(context.Background(), "CREATE TABLE foos (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)
	return db
}

func TestExec(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	result, err := database.Exec(ctx, db, "CREATE TABLE bars (id INTEGER)", []any{})
	require.NoError(t, err)
	require.NotNil(t, result)

	_, err = database.Exec(ctx, db, "INVALID SQL", []any{})
	assert.Error(t, err)
}

func TestUpdateRun(t *testing.T) {
	u := database.Update(func(cb func(tx *sqlx.Tx) error) error {
		return cb(nil)
	})
	err := u.Run(func(tx *sqlx.Tx) error {
		return errors.New("boom")
	})
	assert.EqualError(t, err, "boom")
}

func TestReadRun(t *testing.T) {
	r := database.Read(func(cb func(tx *sqlx.Tx) error) error {
		return cb(nil)
	})
	ran := false
	err := r.Run(func(tx *sqlx.Tx) error {
		ran = true
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, ran)
}

func TestNewUpdate(t *testing.T) {
	ctx := context.Background()
	t.Run("db", func(t *testing.T) {
		db := newTestDB(t)
		u := database.NewUpdate(ctx, nil, db)
		ran := false
		err := u(func(tx *sqlx.Tx) error {
			ran = true
			assert.NotNil(t, tx)
			return nil
		})
		assert.NoError(t, err)
		assert.True(t, ran)
	})
	t.Run("tx", func(t *testing.T) {
		db := newTestDB(t)
		tx, err := db.Beginx()
		require.NoError(t, err)
		defer tx.Rollback()
		u := database.NewUpdate(ctx, nil, tx)
		err = u(func(tx *sqlx.Tx) error {
			return nil
		})
		assert.NoError(t, err)
	})
	t.Run("unsupported type", func(t *testing.T) {
		u := database.NewUpdate(ctx, nil, &fakeDB{})
		err := u(func(tx *sqlx.Tx) error {
			return nil
		})
		assert.Error(t, err)
	})
	t.Run("begin fails", func(t *testing.T) {
		db := newTestDB(t)
		err := db.Close()
		require.NoError(t, err)
		u := database.NewUpdate(ctx, nil, db)
		err = u(func(tx *sqlx.Tx) error {
			return nil
		})
		assert.Error(t, err)
	})
	t.Run("commit error", func(t *testing.T) {
		db := newTestDB(t)
		tx, err := db.Beginx()
		require.NoError(t, err)
		defer tx.Rollback()
		u := database.NewUpdate(ctx, nil, tx)
		err = u(func(tx *sqlx.Tx) error {
			return tx.Commit()
		})
		assert.Error(t, err)
	})
	t.Run("rollback error joins", func(t *testing.T) {
		db := newTestDB(t)
		tx, err := db.Beginx()
		require.NoError(t, err)
		defer tx.Rollback()
		u := database.NewUpdate(ctx, nil, tx)
		err = u(func(tx *sqlx.Tx) error {
			_ = tx.Commit()
			return errors.New("boom")
		})
		assert.Error(t, err)
	})
	t.Run("panic rolls back", func(t *testing.T) {
		db := newTestDB(t)
		u := database.NewUpdate(ctx, nil, db)
		assert.Panics(t, func() {
			_ = u(func(tx *sqlx.Tx) error {
				panic("kaboom")
			})
		})
	})
	t.Run("with mutex", func(t *testing.T) {
		db := newTestDB(t)
		mtx := &sync.Mutex{}
		u := database.NewUpdate(ctx, mtx, db)
		err := u(func(tx *sqlx.Tx) error {
			return nil
		})
		assert.NoError(t, err)
	})
	t.Run("returns callback error", func(t *testing.T) {
		db := newTestDB(t)
		u := database.NewUpdate(ctx, nil, db)
		err := u(func(tx *sqlx.Tx) error {
			return errors.New("boom")
		})
		assert.EqualError(t, err, "boom")
	})
}

func TestNewRead(t *testing.T) {
	ctx := context.Background()
	t.Run("db", func(t *testing.T) {
		db := newTestDB(t)
		r := database.NewRead(ctx, nil, db)
		ran := false
		err := r(func(tx *sqlx.Tx) error {
			ran = true
			assert.NotNil(t, tx)
			return nil
		})
		assert.NoError(t, err)
		assert.True(t, ran)
	})
	t.Run("rw mutex", func(t *testing.T) {
		db := newTestDB(t)
		r := database.NewRead(ctx, &sync.RWMutex{}, db)
		err := r(func(tx *sqlx.Tx) error {
			return nil
		})
		assert.NoError(t, err)
	})
	t.Run("locker", func(t *testing.T) {
		db := newTestDB(t)
		r := database.NewRead(ctx, &sync.Mutex{}, db)
		err := r(func(tx *sqlx.Tx) error {
			return nil
		})
		assert.NoError(t, err)
	})
	t.Run("unsupported type", func(t *testing.T) {
		r := database.NewRead(ctx, nil, &fakeDB{})
		err := r(func(tx *sqlx.Tx) error {
			return nil
		})
		assert.Error(t, err)
	})
}

func TestValue(t *testing.T) {
	u := database.Update(func(cb func(tx *sqlx.Tx) error) error {
		return cb(nil)
	})
	v, err := database.Value(u, func(tx *sqlx.Tx) (int, error) {
		return 42, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 42, v)

	_, err = database.Value(u, func(tx *sqlx.Tx) (int, error) {
		return 0, errors.New("boom")
	})
	assert.EqualError(t, err, "boom")
}

type tabler struct{}

func (tabler) Table() string { return "my_custom_table" }

func TestGetTableTabler(t *testing.T) {
	assert.Equal(t, "my_custom_table", database.GetTable(tabler{}))
	assert.Equal(t, "my_custom_table", database.GetTable(&tabler{}))
	assert.Equal(t, "my_custom_table", database.GetTable((*tabler)(nil)))
}
