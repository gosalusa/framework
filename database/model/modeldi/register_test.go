package modeldi

import (
	"context"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/database/dialects/sqlite"
	"gosalusa.com/database/model"
	"gosalusa.com/di"
	"gosalusa.com/request"
)

type testModel struct {
	model.BaseModel
	ID   int    `db:"id" primaryKey:"true"`
	Name string `db:"name"`
}

func (m *testModel) Table() string { return "test_models" }

type modelRequest struct {
	Model *testModel `inject:"id"`
}

func TestGetValue(t *testing.T) {
	t.Run("query param", func(t *testing.T) {
		r, err := http.NewRequest("GET", "/foo?id=5", nil)
		require.NoError(t, err)
		v, ok := getValue(r, "id")
		assert.True(t, ok)
		assert.Equal(t, "5", v)
	})
	t.Run("mux var", func(t *testing.T) {
		r, err := http.NewRequest("GET", "/foo", nil)
		require.NoError(t, err)
		r = mux.SetURLVars(r, map[string]string{"id": "7"})
		v, ok := getValue(r, "id")
		assert.True(t, ok)
		assert.Equal(t, "7", v)
	})
	t.Run("path value", func(t *testing.T) {
		r, err := http.NewRequest("GET", "/foo", nil)
		require.NoError(t, err)
		r.SetPathValue("id", "9")
		v, ok := getValue(r, "id")
		assert.True(t, ok)
		assert.Equal(t, "9", v)
	})
	t.Run("not found", func(t *testing.T) {
		r, err := http.NewRequest("GET", "/foo", nil)
		require.NoError(t, err)
		v, ok := getValue(r, "id")
		assert.False(t, ok)
		assert.Empty(t, v)
	})
}

func newTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	sqlite.UseSQLite()
	db, err := sqlx.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})
	_, err = db.ExecContext(context.Background(), "CREATE TABLE test_models (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
	require.NoError(t, err)
	return db
}

func registerDeps(t *testing.T, ctx context.Context, req *http.Request, db *sqlx.DB) {
	t.Helper()
	di.Register[*http.Request](ctx, func(ctx context.Context, tag string) (*http.Request, error) {
		return req, nil
	})
	di.Register[*sqlx.DB](ctx, func(ctx context.Context, tag string) (*sqlx.DB, error) {
		return db, nil
	})
}

func TestRegister(t *testing.T) {
	t.Run("not in request", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		req, err := http.NewRequest("GET", "/foo", nil)
		require.NoError(t, err)
		registerDeps(t, ctx, req, nil)
		Register[*testModel](ctx)

		deps := &modelRequest{}
		err = di.Fill(ctx, deps)
		assert.EqualError(t, err, "failed to fill: could not fetch model: id not in request")
	})

	t.Run("not found in database", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		db := newTestDB(t)
		req, err := http.NewRequest("GET", "/foo?id=99", nil)
		require.NoError(t, err)
		registerDeps(t, ctx, req, db)
		Register[*testModel](ctx)

		deps := &modelRequest{}
		err = di.Fill(ctx, deps)
		assert.ErrorIs(t, err, request.ErrStatusNotFound)
	})

	t.Run("found", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		db := newTestDB(t)
		_, err := db.ExecContext(context.Background(), "INSERT INTO test_models (name) VALUES ('a')")
		require.NoError(t, err)
		req, err := http.NewRequest("GET", "/foo?id=1", nil)
		require.NoError(t, err)
		registerDeps(t, ctx, req, db)
		Register[*testModel](ctx)

		deps := &modelRequest{}
		err = di.Fill(ctx, deps)
		require.NoError(t, err)
		m := deps.Model
		require.NotNil(t, m)
		assert.Equal(t, 1, m.ID)
		assert.Equal(t, "a", m.Name)
		assert.True(t, m.InDatabase())
	})

	t.Run("query error", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		db := newTestDB(t)
		require.NoError(t, db.Close())
		req, err := http.NewRequest("GET", "/foo?id=1", nil)
		require.NoError(t, err)
		registerDeps(t, ctx, req, db)
		Register[*testModel](ctx)

		deps := &modelRequest{}
		err = di.Fill(ctx, deps)
		assert.Error(t, err)
	})
}
