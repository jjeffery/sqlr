package sqlrow

import (
	"database/sql"
	"database/sql/driver"

	"github.com/jjeffery/sqlrow/private/dialect"
)

// Dialect is an interface used to handle differences
// in SQL dialects.
type Dialect interface {
	// Quote a table name or column name so that it does
	// not clash with any reserved words. The SQL-99 standard
	// specifies double quotes (eg "table_name"), but many
	// dialects, including MySQL use the backtick (eg `table_name`).
	// SQL server uses square brackets (eg [table_name]).
	Quote(column string) string

	// Return the placeholder for binding a variable value.
	// Most SQL dialects support a single question mark (?), but
	// PostgreSQL uses numbered placeholders (eg $1).
	Placeholder(n int) string
}

// Pre-defined dialects
var (
	Postgres Dialect // Quote: "column_name", Placeholders: $1, $2, $3
	MySQL    Dialect // Quote: `column_name`, Placeholders: ?, ?, ?
	MSSQL    Dialect // Quote: [column_name], Placeholders: ?, ?, ?
	SQLite   Dialect // Quote: `column_name`, Placeholders: ?, ?, ?
	ANSISQL  Dialect // Quote: "column_name", Placeholders: ?, ?, ?
)

// DefaultDialect is the dialect used by a schema if none is specified.
// It is chosen from the first driver in the list of drivers returned by the
// sql.Drivers() function.
//
// Many programs only load one database driver, and in this case the default
// dialect should be the correct choice.
var DefaultDialect Dialect // Depends on the DB driver loaded.

var allDialects []Dialect

func init() {
	Postgres = dialect.Postgres
	MySQL = dialect.MySQL
	MSSQL = dialect.MSSQL
	SQLite = dialect.SQLite
	ANSISQL = dialect.ANSI
	allDialects = []Dialect{Postgres, MySQL, MSSQL, SQLite, ANSISQL}

	DefaultDialect = ANSISQL

	// If one or more drivers have been loaded, choose the default dialect
	// based on the first driver in the list. If there are multiple drivers
	// the first driver is going to be the first alphabetically, as the driver
	// names are sorted.
	if drivers := sql.Drivers(); len(drivers) > 0 {
		switch drivers[0] {
		case "postgres":
			DefaultDialect = Postgres
		case "mysql":
			DefaultDialect = MySQL
		case "sqlite", "sqlite3":
			DefaultDialect = SQLite
		case "mssql":
			DefaultDialect = MSSQL
		}
	}
}

func dialectFor(db *sql.DB) Dialect {
	if db != nil {
		if drvr := db.Driver(); drvr != nil {
			for _, dlct := range allDialects {
				if matcher, ok := dlct.(interface {
					Match(driver.Driver) bool
				}); ok {
					if matcher.Match(drvr) {
						return dlct
					}
				}
			}
		}
	}
	// dialect not found for driver, use default
	return DefaultDialect
}
