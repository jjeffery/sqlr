package sqlf

import (
	"github.com/jjeffery/sqlf/private/colname"
	"github.com/jjeffery/sqlf/private/dialect"
)

// DefaultSchema contains the default schema, which can be
// modified as required.
//
// The default schema has sensible defaults. If not explicitly
// specified, the dialect is determined by the SQL database drivers
// loaded. If the program only uses one database driver, then the default
// schema will use the correct dialect.
//
// The default naming convention uses "snake case". So a struct field
// named "GivenName" will have an associated column name of "given_name".
var DefaultSchema *Schema = &Schema{}

// Schema contains configuration information that is common
// to statements prepared for the same database schema.
//
// If a program works with a single database driver (eg "mysql"),
// and columns conform to a standard naming convention, then that
// progam can use the default schema (DefaultSchema) and there is
// no need to use the Schema type directly.
//
// Programs that operate with a number of different database
// drivers and naming conventions should create a schema for each
// combination of driver and naming convention, and use the appropriate
// schema to prepare each statements
type Schema struct {
	// Dialect used for constructing SQL queries. If nil, the dialect
	// is inferred from the list of SQL drivers loaded in the program.
	Dialect Dialect

	// Convention contains methods for inferring the name
	// of database columns from the associated Go struct field names.
	Convention Convention

	// Logger is used for diagnostic logging. If set then all statements
	// created for this schema will share this logger.
	Logger Logger
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

func (s *Schema) NewInsertRowStmt(row interface{}, sql string) *InsertRowStmt {
	return newInsertRowStmt(s, row, sql)
}

func (s *Schema) NewUpdateRowStmt(row interface{}, sql string) *ExecRowStmt {
	return newUpdateRowStmt(s, row, sql)
}

func (s *Schema) NewGetRowStmt(row interface{}, sql string) *GetRowStmt {
	return newGetRowStmt(s, row, sql)
}

func (s *Schema) NewSelectStmt(row interface{}, sql string) *SelectStmt {
	return newSelectStmt(DefaultSchema, row, sql)
}
