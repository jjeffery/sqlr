package sqlf

import (
	"github.com/jjeffery/sqlf/private/colname"
	"github.com/jjeffery/sqlf/private/dialect"
)

// DefaultSchema contains default schema settings, which can be
// overridden by the calling program.
var DefaultSchema = &Schema{}

// Schema contains common information for tables in a DB schema.
type Schema struct {
	// Dialect used for constructing SQL queries.
	Dialect Dialect

	// Convention contains conventions for inferring the name
	// of database columns from the associated Go struct field names.
	Convention Convention
}

// Table returns information about a table, the structure of whose
// rows are represented by the structure of row.
func (s *Schema) Table(name string, row interface{}) *TableInfo {
	return newTable(s, name, row)
}

func (s *Schema) dialect() dialect.Dialect {
	if s.Dialect != nil {
		return s.Dialect
	}
	if DefaultSchema.Dialect != nil {
		return DefaultSchema.Dialect
	}
	return dialect.New("")
}

func (s *Schema) convention() Convention {
	if s.Convention != nil {
		return s.Convention
	}
	if DefaultSchema.Convention != nil {
		return DefaultSchema.Convention
	}
	return colname.Snake
}
