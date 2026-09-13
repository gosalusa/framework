package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/database"
	"gosalusa.com/database/builder"
	"gosalusa.com/database/dbtest"
	"gosalusa.com/database/dialects"
	"gosalusa.com/database/dialects/generic"
	"gosalusa.com/database/dialects/mysql"
	"gosalusa.com/database/dialects/postgres"
	"gosalusa.com/database/dialects/sqlite"
	"gosalusa.com/database/migrate"
	"gosalusa.com/database/model"
	"gosalusa.com/database/model/mixins"
	"gosalusa.com/di"
)

var sqliteConfig = sqlite.NewConfig(":memory:")
var mysqlConfig = &mysql.SimpleConfig{
	Host:     "localhost",
	Database: "test_db",
	Username: "root",
	Password: "root_password",
}
var pgsqlConfig = &postgres.Config{
	Host:       "localhost",
	Database:   "test_db",
	Username:   "user",
	Password:   "password",
	DisableSSL: true,
}
var sqliteRunner = dbtest.NewRunner(initDB(sqliteConfig))
var mysqlRunner = dbtest.NewRunner(initDB(mysqlConfig))
var pgsqlRunner = dbtest.NewRunner(initDB(pgsqlConfig))

type NamedRunner struct {
	Name   string
	Runner *dbtest.Runner
	Short  bool
}

var namedRunners = []NamedRunner{
	{Name: "sqlite", Runner: sqliteRunner, Short: true},
	{Name: "mysql", Runner: mysqlRunner, Short: false},
	{Name: "pgsql", Runner: pgsqlRunner, Short: false},
}

var mysqlLock *os.File
var pgsqlLock *os.File

type Case[T any] struct {
	Name               string
	Builder            T
	ExpectedSQLite     string
	ExpectedMySQL      string
	ExpectedPostgreSQL string
	ExpectedBindings   []any
}

func QueryTest(t *testing.T, testCases []Case[dialects.QueryBuilder]) {
	t.Helper()
	RawQueryTest(t, testCases, func(d dialects.Dialect, b dialects.QueryBuilder) (dialects.RawQuery, error) {
		return d.EncodeSelectQuery(b.Query())
	})
}
func DeleteQueryTest(t *testing.T, testCases []Case[dialects.DeleteQueryBuilder]) {
	t.Helper()
	RawQueryTest(t, testCases, func(d dialects.Dialect, b dialects.DeleteQueryBuilder) (dialects.RawQuery, error) {
		return d.EncodeDeleteQuery(b.DeleteQuery())
	})
}
func UpdateQueryTest(t *testing.T, testCases []Case[*dialects.UpdateQuery]) {
	t.Helper()
	RawQueryTest(t, testCases, func(d dialects.Dialect, b *dialects.UpdateQuery) (dialects.RawQuery, error) {
		return d.EncodeUpdateQuery(b)
	})
}
func CreateTableTest(t *testing.T, testCases []Case[dialects.CreateTableQueryBuilder]) {
	t.Helper()
	RawQueryTest(t, testCases, func(d dialects.Dialect, b dialects.CreateTableQueryBuilder) (dialects.RawQuery, error) {
		return d.EncodeCreateTableQuery(b.CreateTableQuery())
	})
}
func AlterTableTest(t *testing.T, testCases []Case[dialects.AlterTableQueryBuilder]) {
	t.Helper()
	RawQueryTest(t, testCases, func(d dialects.Dialect, b dialects.AlterTableQueryBuilder) (dialects.RawQuery, error) {
		return d.EncodeAlterTableQuery(b.AlterTableQuery())
	})
}
func ColumnDefinitionTest(t *testing.T, testCases []Case[*dialects.ColumnDefinition]) {
	t.Helper()
	RawQueryTest(t, testCases, func(d dialects.Dialect, b *dialects.ColumnDefinition) (dialects.RawQuery, error) {
		return d.(*generic.Generic).EncodeColumnDefinition(b)
	})
}
func RawQueryTest[T any](t *testing.T, testCases []Case[T], encoder func(d dialects.Dialect, b T) (dialects.RawQuery, error)) {
	t.Helper()
	sqliteDialect := sqlite.New()
	mysqlDialect := mysql.New()
	pgsqlDialect := postgres.New()
	for _, tc := range testCases {
		t.Run(tc.Name+"_sqlite", func(t *testing.T) {
			result, err := encoder(sqliteDialect, tc.Builder)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.ExpectedSQLite, result.SQL)
				assert.Equal(t, tc.ExpectedBindings, result.Bindings)
			}
		})
		if tc.ExpectedMySQL != "" {
			t.Run(tc.Name+"_mysql", func(t *testing.T) {
				result, err := encoder(mysqlDialect, tc.Builder)
				if assert.NoError(t, err) {
					assert.Equal(t, tc.ExpectedMySQL, result.SQL)
					assert.Equal(t, tc.ExpectedBindings, result.Bindings)
				}
			})
		}
		if tc.ExpectedPostgreSQL != "" {
			t.Run(tc.Name+"_pgsql", func(t *testing.T) {
				result, err := encoder(pgsqlDialect, tc.Builder)
				if assert.NoError(t, err) {
					assert.Equal(t, tc.ExpectedPostgreSQL, result.SQL)
					assert.Equal(t, tc.ExpectedBindings, result.Bindings)
				}
			})
		}
	}
}

type EncoderTestCase[T any] struct {
	Name             string
	Builder          T
	ExpectedSQL      string
	ExpectedBindings []any
	ExpectedError    error
}

func EncoderTest[T any](t *testing.T, encoder func(v T) (dialects.RawQuery, error), testCases []EncoderTestCase[T]) {
	t.Helper()
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Helper()
			result, err := encoder(tc.Builder)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.ExpectedSQL, result.SQL)
				assert.Equal(t, tc.ExpectedBindings, result.Bindings)
			}
		})
	}
}

func lockTestDB(driver string) (*os.File, error) {
	lockFile := filepath.Join(os.TempDir(), "salusa-"+driver+"-test-db.lock")
	f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open %s test database lock file: %w", driver, err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, fmt.Errorf("lock %s test database: %w", driver, err)
	}
	return f, nil
}

func initDB(cfg database.Config) func() (*sqlx.DB, error) {
	return func() (*sqlx.DB, error) {
		ctx := di.TestDependencyProviderContext()
		database.Register(ctx, cfg, nil)
		db, err := di.Resolve[*sqlx.DB](ctx)
		if err != nil {
			return nil, fmt.Errorf("open db: %w", err)
		}

		switch db.DriverName() {
		case "sqlite", "sqlite3":
		case "mysql":
			if mysqlLock == nil {
				mysqlLock, err = lockTestDB("mysql")
				if err != nil {
					return nil, err
				}
			}
			err = dropMySQLTables(ctx, db)
			if err != nil {
				return nil, err
			}
		case "postgres":
			if pgsqlLock == nil {
				pgsqlLock, err = lockTestDB("postgres")
				if err != nil {
					return nil, err
				}
			}
			_, err = db.ExecContext(ctx, "DROP SCHEMA public CASCADE;CREATE SCHEMA public;")
			if err != nil {
				return nil, err
			}
		}

		err = migrate.RunModelCreate(ctx, db, &Foo{}, &Bar{}, &FooSoftDelete{})
		if err != nil {
			return nil, fmt.Errorf("create test tables: %s: %w", db.DriverName(), err)
		}
		return db, nil
	}
}

func dropMySQLTables(ctx context.Context, db *sqlx.DB) error {
	tables := []string{}
	err := builder.NewBuilder().
		WithContext(ctx).
		Select("table_name").
		From("information_schema.tables").
		WhereRaw("table_schema = DATABASE()").
		Load(db, &tables)
	if err != nil {
		return err
	}
	if len(tables) == 0 {
		return nil
	}

	d, err := dialects.New("mysql")
	if err != nil {
		return err
	}

	results := make([]dialects.RawQuery, len(tables))
	for i, t := range tables {
		sql, err := d.EncodeDropTableQuery(&dialects.DropTableQuery{Table: t})
		if err != nil {
			return err
		}
		results[i] = sql
	}
	query := dialects.JoinQueries(results)

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0")
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, query.SQL, query.Bindings...)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1")
	if err != nil {
		return err
	}
	return tx.Commit()
}

func Run(t *testing.T, name string, cb func(t *testing.T, tx *sqlx.Tx)) {
	t.Helper()
	runners(t, name, func(runner *dbtest.Runner, name string) {
		t.Helper()
		runner.Run(t, name, cb)
	})
}

func RunBenchmark(t *testing.B, name string, cb func(t *testing.B, tx *sqlx.Tx)) {
	t.Helper()
	runners(t, name, func(runner *dbtest.Runner, name string) {
		runner.RunBenchmark(t, name, cb)
	})
}

func runners(t testing.TB, name string, cb func(runner *dbtest.Runner, name string)) {
	t.Helper()
	for _, r := range namedRunners {
		if testing.Short() && !r.Short {
			t.Skipf("Skipping %s tests", r.Name)
		} else {
			cb(r.Runner, strings.TrimSpace(name+" "+r.Name))
		}
	}
}

type Foo struct {
	model.BaseModel
	ID   int                    `json:"id"   db:"id,primary,autoincrement"`
	Name string                 `json:"name" db:"name"`
	Bar  *builder.HasOne[*Bar]  `json:"bar"`
	Bars *builder.HasMany[*Bar] `json:"bars"`
}

func (h *Foo) Table() string {
	return "foos"
}

type Bar struct {
	model.BaseModel
	ID    int                      `json:"id"     db:"id,primary,autoincrement"`
	FooID int                      `json:"foo_id" db:"foo_id"`
	Foo   *builder.BelongsTo[*Foo] `json:"foo"`
}

func (h *Bar) Table() string {
	return "bars"
}

type FooSoftDelete struct {
	model.BaseModel
	mixins.SoftDelete
	ID   int    `json:"id"   db:"id,primary,autoincrement"`
	Name string `json:"name" db:"name"`
}

func (h *FooSoftDelete) Table() string {
	return "foo_soft_deletes"
}
