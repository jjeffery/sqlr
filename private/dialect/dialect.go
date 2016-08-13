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

// For returns a dialect for the specified database driver.
// If name is blank, then the dialect returned is for the first
// driver returned by sql.Drivers(). If the driver name is
// unknown, the default dialect is returned.
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
func For(name string) Dialect {
	if name == "" {
		drivers := sql.Drivers()
		if len(drivers) > 0 {
			name = drivers[0]
		}
	}
	name = strings.TrimSpace(strings.ToLower(name))

	d := dialects[name]
	if d == nil {
		d = defaultDialect
	}

	return d
}

// dialectT implements the Dialect interface.
type dialectT struct {
	name            string
	altnames        []string
	quoteFunc       func(name string) string
	placeholderFunc func(n int) string
}

func (d *dialectT) Name() string {
	return d.name
}

func (d *dialectT) Quote(name string) string {
	if d.quoteFunc == nil {
		return name
	}
	return d.quoteFunc(name)
}

func (d *dialectT) Placeholder(n int) string {
	if d.placeholderFunc == nil {
		return "?"
	}
	return d.placeholderFunc(n)
}

// SQL Dialects for supported database servers.
var (
	dialects       map[string]*dialectT
	defaultDialect *dialectT
)

func init() {
	dialects = make(map[string]*dialectT)
	defaultDialect = &dialectT{name: "default", quoteFunc: quoteFunc(`"`, `"`)}

	for _, d := range []*dialectT{
		&dialectT{
			name:      "mysql",
			quoteFunc: quoteFunc("`", "`"),
		},
		&dialectT{
			name:      "sqlite",
			altnames:  []string{"sqlite3"},
			quoteFunc: quoteFunc("`", "`"),
		},
		&dialectT{
			name:      "mssql",
			quoteFunc: quoteFunc("[", "]"),
		},
		&dialectT{
			name:            "postgres",
			altnames:        []string{"pq", "postgresql"},
			quoteFunc:       quoteFunc(`"`, `"`),
			placeholderFunc: placeholderFunc("$%d"),
		},
		&dialectT{
			name:            "ql",
			altnames:        []string{"ql-mem"},
			placeholderFunc: placeholderFunc("?%d"),
		},
	} {
		dialects[d.name] = d
		for _, altname := range d.altnames {
			dialects[altname] = d
		}
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
