package migrate

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/jmoiron/sqlx"
	"gosalusa.com/database"
	"gosalusa.com/database/builder"
	"gosalusa.com/database/dialects"
	"gosalusa.com/database/model"
	"gosalusa.com/database/schema"
	"gosalusa.com/di"
	"gosalusa.com/extra/sets"
)

type DBMigration struct {
	model.BaseModel
	Name  string `db:"name,primary"`
	Run   bool   `db:"run"`
	table string
}

func (m *DBMigration) Table() string {
	return m.table
}

type Migrations struct {
	table      string
	migrations []*Migration
}

func New() *Migrations {
	return &Migrations{
		table:      "migrations",
		migrations: []*Migration{},
	}
}

func (m *Migrations) Add(migrations ...*Migration) {
	m.migrations = append(m.migrations, migrations...)
}

func (m *Migrations) isTableCreated(table string) bool {
	for _, m := range m.migrations {
		blueprinter, ok := m.Up.(schema.Blueprinter)
		if !ok {
			continue
		}
		if blueprinter.Type() == schema.BlueprintTypeCreate {
			if blueprinter.GetBlueprint().TableName() == table {
				return true
			}
		}
	}

	return false
}

func (m *Migrations) GenerateMigration(migrationName, packageName string, model model.Model) (string, error) {
	if !m.isTableCreated(database.GetTable(model)) {
		up, err := CreateFromModel(model)
		if err != nil {
			return "", err
		}
		return SrcFile(migrationName, packageName, up, drop(model))
	}

	up, down, err := m.update(model)
	if err != nil {
		return "", err
	}
	return SrcFile(migrationName, packageName, up, down)
}

func (m *Migrations) Blueprint(tableName string) *schema.Blueprint {
	result := &schema.Blueprint{}

	for _, migration := range m.migrations {
		blueprinter, ok := migration.Up.(schema.Blueprinter)
		if !ok {
			continue
		}
		blueprint := blueprinter.GetBlueprint()
		if blueprint.TableName() != tableName {
			continue
		}

		if blueprinter.Type() == schema.BlueprintTypeCreate {
			result = blueprint
		} else {
			result.Merge(blueprint)
		}
	}
	return result
}

func (m *Migrations) Up(ctx context.Context, db database.DB) error {
	q := schema.Create(m.table, func(b *schema.Blueprint) {
		b.String("name").Primary()
		b.Bool("run")
	}).IfNotExists()

	d, err := dialects.New(db.DriverName())
	if err != nil {
		return err
	}
	createTable, err := d.EncodeCreateTableQuery(q.CreateTableQuery())
	if err != nil {
		return err
	}
	_, err = database.Exec(ctx, db, createTable.SQL, createTable.Bindings)
	if err != nil {
		return err
	}

	migrations, err := builder.New[*DBMigration]().
		Select("*").
		From(m.table).
		OrderBy("name").
		WithContext(ctx).
		Get(db)
	if err != nil {
		return err
	}

	runMigrations := sets.New[string]()
	for _, migration := range migrations {
		if migration.Run {
			runMigrations.Add(migration.Name)
		}
	}
	update := database.NewUpdate(ctx, nil, db)
	logger, err := di.Resolve[*slog.Logger](ctx)
	if err != nil {
		logger = slog.Default()
	}

	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Name < m.migrations[j].Name
	})
	for _, migration := range m.migrations {
		err = update(func(tx *sqlx.Tx) error {
			if runMigrations.Has(migration.Name) {
				return nil
			}

			logger.Info("starting migration", "name", migration.Name)

			m := &DBMigration{
				Name:  migration.Name,
				Run:   false,
				table: m.table,
			}
			err = model.SaveContext(ctx, tx, m)
			if err != nil {
				return err
			}
			err := migration.Up.Run(ctx, tx)
			if err != nil {
				return fmt.Errorf("failed to prepare migration %s: %w", migration.Name, err)
			}

			m.Run = true
			err = model.SaveContext(ctx, tx, m)
			if err != nil {
				return err
			}
			logger.Info("finished migration", "name", migration.Name)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
