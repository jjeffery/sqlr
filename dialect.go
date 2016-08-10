package sqlstmt

import (
	"github.com/jjeffery/sqlstmt/private/dialect"
)

// Dialect is an interface used to handle differences
// in SQL dialects.
type Dialect interface {
	// Name of the dialect. This name is used as
	// a key for caching, so if If two dialects have
	// the same name, then they should be identical.
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

// DialectFor returns the dialect for the specified database driver.
// If name is blank, then the dialect returned is for the first
// driver returned by sql.Drivers(). If only one SQL driver has
// been loaded by the calling program then this will return the
// correct dialect. If the driver name is unknown, the default
// dialect is returned.
//
// Supported dialects include:
//
//  name      alternative names
//  ----      -----------------
//  mssql
//  mysql
//  postgres  pq, postgresql
//  sqlite3   sqlite
//  ql        ql-mem
func DialectFor(name string) Dialect {
	return dialect.For(name)
}
