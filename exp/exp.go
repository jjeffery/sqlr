package exp

import "database/sql"

type user struct {
	ID         string
	FamilyName string
	GivenName  string
	Email      string
}

// Execer implements a single method, Exec, which executes a
// query without returning any rows. The args are for any parameter
// placeholders in the query.
//
// This interface is compatible with the standard library package "database/sql".
// Both *sql.DB and *sql.Tx implement this interface.
type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// Queryer implements a single method, Query, which executes
// a query that returns zero, one or more rows. The args are
// for any parameter placeholders in the query.
//
// This interface is compatible with the standard library package "database/sql".
// Both *sql.DB and *sql.Tx implement this interface.
type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// Schema contains information that is common to a database schema.
// It is used when preparing SQL to run against a particular database.
type Schema struct {
	// Dialect used for constructing SQL queries.
	Dialect Dialect

	// ColumnNameFor converts a Go struct field name to a column name.
	ColumnNameFor func(fieldName string) string
}

// DefaultSchema provides a default schema instance.
//
// Many programs will not have to create an instance of Schema: if a
// program uses a single database schema, then the DefaultSchema can be
// used instead.
var DefaultSchema Schema

func DefineTable(name string, row interface{}) {
	DefaultSchema.DefineTable(name, row)
}

func InsertRow(execer Execer, row interface{}) error {
	return DefaultSchema.InsertRow(execer, row)
}

func UpdateRow(execer Execer, row interface{}) (int, error) {
	return DefaultSchema.UpdateRow(execer, row)
}

func DeleteRow(execer Execer, row interface{}) (int, error) {
	return DefaultSchema.DeleteRow(execer, row)
}

func SelectRow(queryer Queryer, row interface{}) error {
	return DefaultSchema.SelectRow(queryer, row)
}

func PrepareQuery(query string) (*QueryStmt, error) {
	return DefaultSchema.PrepareQuery(query)
}

func MustPrepareQuery(query string) *QueryStmt {
	return DefaultSchema.MustPrepareQuery(query)
}

func (s *Schema) DefineTable(name string, row interface{}) {}

func (s *Schema) InsertRow(execer Execer, row interface{}) error {
	return nil
}

func (s *Schema) UpdateRow(execer Execer, row interface{}) (int, error) {
	return 0, nil
}

func (s *Schema) DeleteRow(execer Execer, row interface{}) (int, error) {
	return 0, nil
}

func (s *Schema) SelectRow(queryer Queryer, row interface{}) error {
	return nil
}

func (s *Schema) PrepareQuery(query string) (*QueryStmt, error) {
	return &QueryStmt{}, nil
}

func (s *Schema) MustPrepareQuery(query string) *QueryStmt {
	stmt, err := s.PrepareQuery(query)
	if err != nil {
		panic(err)
	}
	return stmt
}

type QueryStmt struct{}

func (s *QueryStmt) Select(queryer Queryer, dest interface{}, args ...interface{}) error {
	return nil
}

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

// NewDialect returns a dialect for the specified driver. If driverName
// is blank, then the first driver is chosen from the list of drivers loaded.
// This works well for the common situation where the calling program has loaded
// only one database driver -- that database driver will be used to select the
// dialect.
//
// Supported drivers include:
//  mysql
//  postgres
//  mssql
//  sqlite3
func NewDialect(driverName string) Dialect {
	return nil
}
