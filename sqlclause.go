package sqlf

// sqlClause represents a specific SQL clause. Column lists
// and table names are represented differently depending on
// which SQL clause they appear in.
type sqlClause int

// All of the different clauses of an SQL statement where columns
// and table names can be found.
const (
	clauseSelectColumns sqlClause = iota
	clauseSelectFrom
	clauseSelectOrderBy
	clauseInsertInto
	clauseInsertColumns
	clauseInsertValues
	clauseUpdateTable
	clauseUpdateSet
	clauseUpdateWhere
	clauseDeleteTable
	clauseDeleteWhere
)

// isInput identifies whether the SQL clause contains placeholders
// for variable input.
func (c sqlClause) isInput() bool {
	return c == clauseInsertValues ||
		c == clauseUpdateSet ||
		c == clauseUpdateWhere
}
