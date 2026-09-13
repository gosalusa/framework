package migrate

import (
	"fmt"

	"gosalusa.com/database"
	"gosalusa.com/database/model"
	"gosalusa.com/database/schema"
	"gosalusa.com/internal/relationship"
)

var (
	ErrNoChanges = fmt.Errorf("no changes")
)

func (m *Migrations) update(mod model.Model) (*schema.UpdateTableBuilder, *schema.UpdateTableBuilder, error) {
	err := relationship.InitializeRelationships(mod)
	if err != nil {
		return nil, nil, err
	}

	tableName := database.GetTable(mod)
	fields, err := getFields(mod)
	if err != nil {
		return nil, nil, err
	}
	oldTable := m.Blueprint(tableName)
	newTable := blueprintFromFields(tableName, fields)

	hasChanges := false

	up := schema.Table(tableName, func(table *schema.Blueprint) {
		hasChanges = table.Update(oldTable, newTable)
	})
	down := schema.Table(tableName, func(table *schema.Blueprint) {
		table.Update(newTable, oldTable)
	})
	if !hasChanges {
		return nil, nil, ErrNoChanges
	}

	return up, down, nil
}
