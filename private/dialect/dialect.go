// Package dialect handles differences in various
// SQL dialects.
package dialect

import (
	"database/sql"
	"fmt"
	"strings"
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

// New creates a dialect based on the name.
func New(name string) Dialect {
	if name == "" {
		drivers := sql.Drivers()
		if len(drivers) > 0 {
			name = drivers[0]
		}
	}
	//println("dialect name =", name)
	switch strings.ToLower(name) {
	case "pq", "postgres", "postgresql":
		return dialectPG
	case "mysql":
		return dialectMySQL
	case "mssql":
		return dialectMSSQL
	case "sqlite3", "sqlite":
		return dialectSQLite
	default:
		return dialectDefault
	}
}

// dialectT implements the Dialect interface.
type dialectT struct {
	name            string
	quoteFunc       func(name string) string
	placeholderFunc func(n int) string
}

func (d dialectT) Name() string {
	return d.name
}

func (d dialectT) Quote(name string) string {
	return d.quoteFunc(name)
}

func (d dialectT) Placeholder(n int) string {
	if d.placeholderFunc == nil {
		return "?"
	}
	return d.placeholderFunc(n)
}

// SQL Dialects for supported database servers.
var (
	dialectDefault = dialectT{name: "default", quoteFunc: quoteFunc(`"`, `"`)}
	dialectMySQL   = dialectT{name: "mysql", quoteFunc: quoteFunc("`", "`")}
	dialectSQLite  = dialectT{name: "sqlite", quoteFunc: quoteFunc("`", "`")}
	dialectMSSQL   = dialectT{name: "mssql", quoteFunc: quoteFunc("[", "]")}
	dialectPG      = dialectT{
		name:            "postgres",
		quoteFunc:       quoteFunc(`"`, `"`),
		placeholderFunc: placeholderFunc("$%d"),
	}
)

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
