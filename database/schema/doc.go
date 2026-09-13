// Package schema provides a dialect-agnostic DSL for describing and executing
// database schema changes such as creating, altering, and dropping tables.
//
// Operations are built with fluent builders that chain modifiers on column,
// index, and foreign-key builders, then executed against a database.DB with
// Run:
//
//	schema.Create("posts", func(table *schema.Blueprint) {
//		table.Int("id").AutoIncrement().Primary()
//		table.String("title").Size(255).NotNullable()
//		table.Text("body").Nullable()
//		table.DateTime("created_at").DefaultCurrentTime()
//	}).Run(ctx, tx)
//
// # Builders and the Runner contract
//
// Create starts a new table, Table alters an existing one, and Drop and
// DropIfExists remove tables. Every operation implements Runner, whose Run
// method executes the change against a transaction. This is the contract the
// migrate package relies on: a migrate.Migration holds its Up and Down
// operations as schema.Runner values.
//
// Within Create and Table the callback receives a *Blueprint, which collects
// columns (via shorthand methods such as String, Int, and Text, or OfType),
// indexes, foreign keys, primary keys, and, for Table, columns to drop.
//
// # Rendered through dialects
//
// The builders describe tables as Blueprints and convert them into the
// dialect-neutral query types defined in the dialects package
// (CreateTableQuery, AlterTableQuery, and DropTableQuery). dialects.New picks
// a Dialect based on the database driver in use, which renders the final SQL.
// The same builder therefore produces SQLite, PostgreSQL, or MySQL statements
// without any code changes.
//
// # Generating migrations
//
// migrate.CreateFromModel generates a Create builder from a model struct, and
// Migrations.update diffs a model against the current schema to produce Up and
// Down Table builders. The generated migration templates use GoString on the
// builders to serialize them back into Go source.
package schema