package statement

import (
	"database/sql"

	"github.com/jjeffery/sqlr/private/column"
)

// Dialect is an interface used to handle differences in SQL dialects.
type Dialect interface {
	Quote(name string) string
	Placeholder(n int) string
}

// ColumnNamer provides the column name for a column.
type ColumnNamer interface {
	ColumnName(*column.Info) string
}

type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}
