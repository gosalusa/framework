package schema_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/dialects"
	"gosalusa.com/database/schema"
	"gosalusa.com/internal/test"
)

func TestBuilder(t *testing.T) {
	test.CreateTableTest(t, []test.Case[dialects.CreateTableQueryBuilder]{
		{
			Name:             "create table",
			Builder:          schema.Create("foo", func(table *schema.Blueprint) {}),
			ExpectedSQLite:   `CREATE TABLE "foo" ();`,
			ExpectedBindings: []any{},
		},
		{
			Name: "1 column",
			Builder: schema.Create("foo", func(table *schema.Blueprint) {
				table.String("bar")
			}),
			ExpectedSQLite:   `CREATE TABLE "foo" ("bar" TEXT NOT NULL);`,
			ExpectedBindings: []any{},
		},
		{
			Name: "2 columns",
			Builder: schema.Create("foo", func(table *schema.Blueprint) {
				table.Int("id")
				table.String("bar")
			}),
			ExpectedSQLite:   `CREATE TABLE "foo" ("id" INTEGER NOT NULL, "bar" TEXT NOT NULL);`,
			ExpectedBindings: []any{},
		},
		{
			Name: "size",
			Builder: schema.Create("foo", func(table *schema.Blueprint) {
				table.Int("id")
				table.String("bar").Size(100)
			}),
			ExpectedSQLite:     `CREATE TABLE "foo" ("id" INTEGER NOT NULL, "bar" TEXT NOT NULL);`,
			ExpectedPostgreSQL: `CREATE TABLE "foo" ("id" INTEGER NOT NULL, "bar" TEXT NOT NULL);`,
			ExpectedMySQL:      "CREATE TABLE `foo` (`id` INT NOT NULL, `bar` VARCHAR(100) NOT NULL);",
			ExpectedBindings:   []any{},
		},
		{
			Name: "primary key",
			Builder: schema.Create("foo", func(table *schema.Blueprint) {
				table.Int("id").Primary()
			}),
			ExpectedSQLite:   `CREATE TABLE "foo" ("id" INTEGER PRIMARY KEY NOT NULL);`,
			ExpectedBindings: []any{},
		},
		{
			Name: "composite primary key",
			Builder: schema.Create("foo", func(table *schema.Blueprint) {
				table.Int("id1")
				table.Int("id2")
				table.PrimaryKey("id1", "id2")
			}),
			ExpectedSQLite:   `CREATE TABLE "foo" ("id1" INTEGER NOT NULL, "id2" INTEGER NOT NULL, PRIMARY KEY ("id1", "id2"));`,
			ExpectedBindings: []any{},
		},
		{
			Name: "index",
			Builder: schema.Create("foo", func(table *schema.Blueprint) {
				table.Int("id")
				table.String("name")
				table.Index("name_index").AddColumn("name")
			}),
			ExpectedSQLite:   `CREATE TABLE "foo" ("id" INTEGER NOT NULL, "name" TEXT NOT NULL); CREATE INDEX IF NOT EXISTS "name_index" ON "foo" ("name");`,
			ExpectedBindings: []any{},
		},
		{
			Name: "foreign key",
			Builder: schema.Create("foo", func(table *schema.Blueprint) {
				table.Int("id")
				table.ForeignKey("id", "bar", "foo_id")
			}),
			ExpectedSQLite:   `CREATE TABLE "foo" ("id" INTEGER NOT NULL, CONSTRAINT "id-bar-foo_id" FOREIGN KEY ("id") REFERENCES "bar" ("foo_id"));`,
			ExpectedBindings: []any{},
		},
		{
			Name: "null",
			Builder: schema.Create("foo", func(table *schema.Blueprint) {
				table.Int("id").Nullable()
			}),
			ExpectedSQLite:   `CREATE TABLE "foo" ("id" INTEGER);`,
			ExpectedBindings: []any{},
		},
	})
}

func TestDefault(t *testing.T) {
	test.Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		defer schema.Drop("foo_tmp").Run(context.Background(), tx)

		c := schema.Create("foo_tmp", func(table *schema.Blueprint) {
			table.Int("id").Default(1)
			table.Bool("bool").Default(false)
		})
		err := c.Run(context.Background(), tx)
		assert.NoError(t, err)
	})
}
