package sqlf

import (
	"database/sql"
	"strings"
)

// Dialect is an interface used to handle differences
// in SQL dialects.
type Dialect interface {
	// Quote a table name or column name so that it does
	// not clash with any reserved words. The SQL-99 standard
	// specifies double quotes (eg "table_name"), but many
	// dialects, including MySQL use the backtick (eg `table_name`).
	// SQL serer uses square brackets (eg [table_name]).
	Quote(name string) string

	// Return the placeholder for binding a variable value.
	// Most SQL dialects support a single question mark ("?"), but
	// PostgreSQL uses numbered placeholders (eg "$1").
	Placeholder(n int) string
}

// quoteName will
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

type dialect struct {
	quoteFunc       func(name string) string
	placeholderFunc func(n int) string
}

func (d dialect) Quote(name string) string {
	return d.quoteFunc(name)
}

func (d dialect) Placeholder(n int) string {
	if d.placeholderFunc == nil {
		return "?"
	}
	return d.placeholderFunc(n)
}

// SQL Dialects
var (
	DefaultDialect Dialect // Default dialect (MySQL, can be changed)
	DialectMySQL   Dialect // MySQL dialect
	DialectMSSQL   Dialect // Microsoft SQL Server dialect
)

func init() {
	DialectMySQL = dialect{quoteFunc: quoteFunc("`", "`")}
	DialectMSSQL = dialect{quoteFunc: quoteFunc("[", "]")}
}

func defaultDialect() Dialect {
	if DefaultDialect != nil {
		return DefaultDialect
	}
	for _, d := range sql.Drivers() {
		if strings.Contains(strings.ToLower(d), "mysql") {
			return DialectMySQL
		} else if d == "mssql" {
			return DialectMSSQL
		}
	}
	panic("Cannot determine default dialect. Set DefaultDialect")
}
