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
}

func (cfg *Schema) dialect() dialect.Dialect {
	if cfg.Dialect != nil {
		return cfg.Dialect
	}
	if DefaultSchema.Dialect != nil {
		return DefaultSchema.Dialect
	}
	return dialect.New("")
}

func (cfg *Schema) convention() Convention {
	if cfg.Convention != nil {
		return cfg.Convention
	}
	if DefaultSchema.Convention != nil {
		return DefaultSchema.Convention
	}
	return colname.Snake
}

func (s *Schema) MustPrepareInsertRow(row interface{}, sql string) *InsertRowStmt {
	stmt, err := PrepareInsertRow(row, sql)
	if err != nil {
		panic(err)
	}
	return stmt
}

func (s *Schema) PrepareInsertRow(row interface{}, sql string) (*InsertRowStmt, error) {
	return &InsertRowStmt{}, nil
}

func (s *Schema) MustPrepareUpdateRow(row interface{}, sql string) *UpdateRowStmt {
	stmt, err := PrepareUpdateRow(row, sql)
	if err != nil {
		panic(err)
	}
	return stmt
}

func (s *Schema) PrepareUpdateRow(row interface{}, sql string) (*UpdateRowStmt, error) {
	return &UpdateRowStmt{}, nil
}

func (s *Schema) MustPrepareGetRow(row interface{}, sql string) *GetRowStmt {
	stmt, err := PrepareGetRow(row, sql)
	if err != nil {
		panic(err)
	}
	return stmt
}

func (s *Schema) PrepareGetRow(row interface{}, sql string) (*GetRowStmt, error) {
	return &GetRowStmt{}, nil
}

func (s *Schema) MustPrepareSelect(row interface{}, sql string) *SelectStmt {
	stmt, err := PrepareSelect(row, sql)
	if err != nil {
		panic(err)
	}
	return stmt
}

func (s *Schema) PrepareSelect(row interface{}, sql string) (*SelectStmt, error) {
	return &SelectStmt{}, nil
}
