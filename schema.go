package sqlstmt

import (
	"github.com/jjeffery/sqlstmt/private/colname"
	"github.com/jjeffery/sqlstmt/private/dialect"
)

// Default is the default schema, which can be modified as required.
//
// The default schema has sensible defaults. If not explicitly
// specified, the dialect is determined by the SQL database drivers
// loaded. If the program only uses one database driver, then the default
// schema will use the correct dialect.
//
// The default naming convention uses "snake case". So a struct field
// named "GivenName" will have an associated column name of "given_name".
var Default *Schema = &Schema{}

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
	Logger SQLLogger
}

func (s *Schema) dialect() dialect.Dialect {
	if s.Dialect != nil {
		return s.Dialect
	}
	if Default.Dialect != nil {
		return Default.Dialect
	}
	return dialect.For("")
}

func (s *Schema) convention() Convention {
	if s.Convention != nil {
		return s.Convention
	}
	if Default.Convention != nil {
		return Default.Convention
	}
	return colname.Snake
}

// NewInsertRowStmt returns a new InsertRowStmt for the given
// row and SQL.
// It is safe for concurrent access by multiple goroutines.
func (s *Schema) NewInsertRowStmt(row interface{}, sql string) *InsertRowStmt {
	return newInsertRowStmt(s, row, sql)
}

// NewUpdateRowStmt returns a new ExecRowStmt for updating a single row.
func (s *Schema) NewUpdateRowStmt(row interface{}, sql string) *ExecRowStmt {
	return newUpdateRowStmt(s, row, sql)
}

// NewDeleteRowStmt returns a new ExecRowStmt for deleting a single row.
// It is safe for concurrent access by multiple goroutines.
func (s *Schema) NewDeleteRowStmt(row interface{}, sql string) *ExecRowStmt {
	return newDeleteRowStmt(s, row, sql)
}

// NewGetRowStmt executes a query that returns a single row.
// It is safe for concurrent access by multiple goroutines.
func (s *Schema) NewGetRowStmt(row interface{}, sql string) *GetRowStmt {
	return newGetRowStmt(s, row, sql)
}

// NewSelectStmt executes a query that returns multiple rows.
// It is safe for concurrent access by multiple goroutines.
func (s *Schema) NewSelectStmt(row interface{}, sql string) *SelectStmt {
	return newSelectStmt(Default, row, sql)
}
