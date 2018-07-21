// Package dialect handles differences in various
// SQL dialects.
package dialect

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
)

// Dialect provides information about an SQL dialect.
type Dialect struct {
	driverTypes     []string
	quoteFunc       func(name string) string
	placeholderFunc func(n int) string
}

// PostgresDialect is a dialect for PostgreSQL.
type PostgresDialect struct {
	Dialect
}

// Pre-defined dialects
var (
	ANSI     *Dialect
	MSSQL    *Dialect
	MySQL    *Dialect
	Postgres *PostgresDialect
	SQLite   *Dialect
)

// Quote quotes a column name.
func (d *Dialect) Quote(name string) string {
	return d.quoteFunc(name)
}

// Placeholder returns the string for a placeholder.
func (d *Dialect) Placeholder(n int) string {
	if d.placeholderFunc == nil {
		return "?"
	}
	return d.placeholderFunc(n)
}

// Match returns true if the dialect is appropriate for the driver.
func (d *Dialect) Match(drv driver.Driver) bool {
	driverType := fmt.Sprint(reflect.TypeOf(drv))
	for _, dt := range d.driverTypes {
		if driverType == dt {
			return true
		}
	}
	return false
}

func init() {
	ANSI = &Dialect{
		quoteFunc: quoteFunc(`"`, `"`),
	}
	MSSQL = &Dialect{
		quoteFunc:   quoteFunc("[", "]"),
		driverTypes: []string{"*mssql.MssqlDriver"},
	}
	MySQL = &Dialect{
		quoteFunc:   quoteFunc("`", "`"),
		driverTypes: []string{"*mysql.MySQLDriver"},
	}
	SQLite = &Dialect{
		quoteFunc:   quoteFunc("`", "`"),
		driverTypes: []string{"*sqlite3.SQLiteDriver"},
	}
	Postgres = &PostgresDialect{
		Dialect{
			quoteFunc:       quoteFunc(`"`, `"`),
			placeholderFunc: placeholderFunc("$%d"),
			driverTypes:     []string{"*pq.Driver"},
		},
	}
}

func quoteFunc(begin string, end string) func(name string) string {
	return func(name string) string {
		var names []string
		for _, n := range strings.Split(name, ".") {
			n = strings.TrimLeft(n, "\"`[ \t"+begin)
			n = strings.TrimRight(n, "\"`] \t"+end)
			names = append(names, begin+n+end)
		}
		return strings.Join(names, ".")
	}
}

func placeholderFunc(format string) func(n int) string {
	return func(n int) string {
		return fmt.Sprintf(format, n)
	}
}

// Postgres is a marker method that indicates the dialect is for PostgreSQL.
func (d *PostgresDialect) Postgres() {}
