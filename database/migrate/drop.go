package migrate

import (
	"fmt"

	"gosalusa.com/database"
	"gosalusa.com/database/model"
)

type dropTable string

func drop(table model.Model) dropTable {
	return dropTable(database.GetTable(table))
}

func (dt dropTable) GoString() string {
	return fmt.Sprintf("schema.DropIfExists(%#v)", string(dt))
}
