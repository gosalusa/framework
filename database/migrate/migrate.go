package migrate

import (
	"fmt"
	"go/format"

	"golang.org/x/tools/imports"
	"gosalusa.com/database/schema"
)

type Migration struct {
	Name string
	Up   schema.Runner
	Down schema.Runner
}

func SrcFile(migrationName, packageName string, up, down fmt.GoStringer) (string, error) {
	outFile := "migration.go"
	initSrc := `package %s
	
import (
	"gosalusa.com/database"
	"gosalusa.com/database/migrate"
	"gosalusa.com/database/builder"
	"gosalusa.com/database/schema"
)

func init() {
	migrations.Add(&migrate.Migration{
		Name: %#v,
		Up: %s,
		Down: %s,
	})
}`

	src := []byte(fmt.Sprintf(initSrc, packageName, migrationName, up.GoString(), down.GoString()))
	// fmt.Printf("%s\n", src)
	src, err := imports.Process(outFile, src, nil)
	if err != nil {
		return "", err
	}

	src, err = format.Source(src)
	if err != nil {
		return "", err
	}
	return string(src), nil
}
