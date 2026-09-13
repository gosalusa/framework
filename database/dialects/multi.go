package dialects

// MultiQueryBuilder is implemented by anything that produces a set of
// MultiQueries.
type MultiQueryBuilder interface {
	MultiQuery() []MultiQuery
}

// MultiQuery holds exactly one query of any kind. It is created with the Add
// methods of MultiQueryBuilderImpl.
type MultiQuery struct {
	alterTableQuery  *AlterTableQuery
	createTableQuery *CreateTableQuery
	deleteQuery      *DeleteQuery
	dropTableQuery   *DropTableQuery
	insertQuery      *InsertQuery
	selectQuery      *SelectQuery
	updateQuery      *UpdateQuery
}

// MultiQueryBuilderImpl accumulates queries in order. Each Add method appends
// a query and returns the builder so calls can be chained.
type MultiQueryBuilderImpl struct {
	queries []MultiQuery
}

// NewMultiQueryBuilder returns a new empty MultiQueryBuilderImpl.
func NewMultiQueryBuilder() *MultiQueryBuilderImpl {
	return &MultiQueryBuilderImpl{
		queries: []MultiQuery{},
	}
}

// AddAlterTableQuery appends an ALTER TABLE query to the batch.
func (b *MultiQueryBuilderImpl) AddAlterTableQuery(q *AlterTableQuery) *MultiQueryBuilderImpl {
	b.queries = append(b.queries, MultiQuery{alterTableQuery: q})
	return b
}

// AddCreateTableQuery appends a CREATE TABLE query to the batch.
func (b *MultiQueryBuilderImpl) AddCreateTableQuery(q *CreateTableQuery) *MultiQueryBuilderImpl {
	b.queries = append(b.queries, MultiQuery{createTableQuery: q})
	return b
}

// AddDeleteQuery appends a DELETE query to the batch.
func (b *MultiQueryBuilderImpl) AddDeleteQuery(q *DeleteQuery) *MultiQueryBuilderImpl {
	b.queries = append(b.queries, MultiQuery{deleteQuery: q})
	return b
}

// AddDropTableQuery appends a DROP TABLE query to the batch.
func (b *MultiQueryBuilderImpl) AddDropTableQuery(q *DropTableQuery) *MultiQueryBuilderImpl {
	b.queries = append(b.queries, MultiQuery{dropTableQuery: q})
	return b
}

// AddInsertQuery appends an INSERT query to the batch.
func (b *MultiQueryBuilderImpl) AddInsertQuery(q *InsertQuery) *MultiQueryBuilderImpl {
	b.queries = append(b.queries, MultiQuery{insertQuery: q})
	return b
}

// AddSelectQuery appends a SELECT query to the batch.
func (b *MultiQueryBuilderImpl) AddSelectQuery(q *SelectQuery) *MultiQueryBuilderImpl {
	b.queries = append(b.queries, MultiQuery{selectQuery: q})
	return b
}

// AddUpdateQuery appends an UPDATE query to the batch.
func (b *MultiQueryBuilderImpl) AddUpdateQuery(q *UpdateQuery) *MultiQueryBuilderImpl {
	b.queries = append(b.queries, MultiQuery{updateQuery: q})
	return b
}

// MultiQuery returns the queries that have been added to the batch.
func (b *MultiQueryBuilderImpl) MultiQuery(q *UpdateQuery) []MultiQuery {
	return b.queries
}

// EncodeMultiQuery encodes each query with the dialect and joins the results
// into a single multi-statement query with JoinQueries.
func EncodeMultiQuery(d Dialect, queries []MultiQuery) (RawQuery, error) {
	rawQueries := make([]RawQuery, len(queries))
	var err error
	for i, q := range queries {
		rawQueries[i], err = encodeMultiQuery(d, q)
		if err != nil {
			return RawQuery{}, err
		}
	}
	return JoinQueries(rawQueries), nil
}
func encodeMultiQuery(d Dialect, q MultiQuery) (RawQuery, error) {
	if q.alterTableQuery != nil {
		return d.EncodeAlterTableQuery(q.alterTableQuery)
	} else if q.createTableQuery != nil {
		return d.EncodeCreateTableQuery(q.createTableQuery)
	} else if q.deleteQuery != nil {
		return d.EncodeDeleteQuery(q.deleteQuery)
	} else if q.dropTableQuery != nil {
		return d.EncodeDropTableQuery(q.dropTableQuery)
	} else if q.insertQuery != nil {
		return d.EncodeInsertQuery(q.insertQuery)
	} else if q.selectQuery != nil {
		return d.EncodeSelectQuery(q.selectQuery)
	} else if q.updateQuery != nil {
		return d.EncodeUpdateQuery(q.updateQuery)
	}
	return RawQuery{}, nil
}
