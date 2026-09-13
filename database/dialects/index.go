package dialects

// Index describes a CREATE INDEX statement on the columns of a table. Unique
// makes it a UNIQUE index.
type Index struct {
	Name    string
	Table   string
	Columns []string
	Unique  bool
}
