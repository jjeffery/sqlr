package sqlstmt

import (
	"github.com/jjeffery/sqlstmt/private/dialect"
)

// Dialect is an interface used to handle differences
// in SQL dialects.
type Dialect interface {
	// Name of the dialect.
	Name() string

	// Quote a table name or column name so that it does
	// not clash with any reserved words. The SQL-99 standard
	// specifies double quotes (eg "table_name"), but many
	// dialects, including MySQL use the backtick (eg `table_name`).
	// SQL server uses square brackets (eg [table_name]).
	Quote(name string) string

	// Return the placeholder for binding a variable value.
	// Most SQL dialects support a single question mark (?), but
	// PostgreSQL uses numbered placeholders (eg $1).
	Placeholder(n int) string
}

// New creates a dialect based on the name. Supported dialects include:
//
//  mssql
//  mysql
//  postgres (pq, postgresql)
//  sqlite3 (sqlite)
func NewDialect(name string) Dialect {
	return dialect.New(name)
}
