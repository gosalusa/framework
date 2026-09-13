package schema_test

import (
	"testing"

	"gosalusa.com/database/dialects"
	"gosalusa.com/database/schema"
	"gosalusa.com/internal/test"
)

func TestUpdateTable(t *testing.T) {
	test.AlterTableTest(t, []test.Case[dialects.AlterTableQueryBuilder]{
		{
			Name:             "empty update",
			Builder:          schema.Table("foo", func(table *schema.Blueprint) {}),
			ExpectedSQLite:   "",
			ExpectedBindings: []any{},
		},
		{
			Name: "add column",
			Builder: schema.Table("foo", func(table *schema.Blueprint) {
				table.String("bar")
			}),
			ExpectedSQLite:   "ALTER TABLE \"foo\" ADD \"bar\" TEXT NOT NULL;",
			ExpectedBindings: []any{},
		},
		{
			Name: "change column",
			Builder: schema.Table("foo", func(table *schema.Blueprint) {
				table.Int("id").Change()
			}),
			ExpectedSQLite:   "ALTER TABLE \"foo\" MODIFY COLUMN \"id\" INTEGER NOT NULL;",
			ExpectedBindings: []any{},
		},
		{
			Name: "drop column",
			Builder: schema.Table("foo", func(table *schema.Blueprint) {
				table.DropColumn("id")
			}),
			ExpectedSQLite:   "ALTER TABLE \"foo\" DROP COLUMN \"id\";",
			ExpectedBindings: []any{},
		},
		{
			Name: "add foreign key",
			Builder: schema.Table("foo", func(table *schema.Blueprint) {
				table.ForeignKey("id", "bar", "foo_id")
			}),
			ExpectedSQLite:   "ALTER TABLE \"foo\" ADD CONSTRAINT \"id-bar-foo_id\" FOREIGN KEY (\"id\") REFERENCES \"bar\" (\"foo_id\");",
			ExpectedBindings: []any{},
		},
		// {
		// 	Name: "drop foreign key",
		// 	Builder: schema.Table("foo", func(table *schema.Blueprint) {
		// 		table.ForeignKey("id", "bar", "foo_id")
		// 	}),
		// 	ExpectedSQL:      "",
		// 	ExpectedBindings: []any{},
		// },
		{
			Name: "add index",
			Builder: schema.Table("foo", func(table *schema.Blueprint) {
				table.Index("index-name").AddColumn("foo").AddColumn("bar")
			}),
			ExpectedSQLite:   "CREATE INDEX IF NOT EXISTS \"index-name\" ON \"foo\" (\"foo\", \"bar\");",
			ExpectedBindings: []any{},
		},
		// {
		// 	Name: "drop index",
		// 	Builder: schema.Table("foo", func(table *schema.Blueprint) {
		// 		table.Index("index-name").AddColumn("foo").AddColumn("bar")
		// 	}),
		// 	ExpectedSQL:      "",
		// 	ExpectedBindings: []any{},
		// },
	})
}
