package statement

import "database/sql"

// Dialect is an interface used to handle differences in SQL dialects.
type Dialect interface {
	Quote(name string) string
	Placeholder(n int) string
}

// ColumnNamer provides column names based on the structure field name.
// Multiple field names indicate one or more embedded structures.
type ColumnNamer interface {
	ColumnName(fields ...string) string
}

type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}
