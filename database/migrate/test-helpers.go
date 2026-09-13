package migrate

import (
	"context"

	"gosalusa.com/database"
	"gosalusa.com/database/model"
)

func MustMigrateModel(db database.DB, m model.Model) {
	migration, err := CreateFromModel(m)
	if err != nil {
		panic(err)
	}
	err = migration.Run(context.Background(), db)
	if err != nil {
		panic(err)
	}
}
