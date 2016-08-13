package sqlrow

import (
	"database/sql"
	"errors"
)

var (
	errNotImplemented = errors.New("not implemented")
)

// DB is the interface that wraps the database access methods
// used by this package.
//
// The *DB and *Tx types in the standard library package "database/sql"
// both implement this interface.
type DB interface {
	// Exec executes a query without returning any rows.
	// The args are for any placeholder parameters in the query.
	Exec(query string, args ...interface{}) (sql.Result, error)

	// Query executes a query that returns rows, typically a SELECT.
	// The args are for any placeholder parameters in the query.
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// Insert inserts a row. If the row has an auto-increment column
// defined, then the generated value is retrieved and inserted into the
// row. (If the database driver provides the necessary support).
func Insert(db DB, row interface{}, sql string) error {
	return errNotImplemented
}

// Update updates a row. Returns the number of rows affected,
// which should be zero or one.
func Update(db DB, row interface{}, sql string, args ...interface{}) (int, error) {
	return 0, errNotImplemented
}

// Delete deletes a row. Returns the number of rows affected,
// which should be zero or one.
func Delete(db DB, row interface{}, sql string, args ...interface{}) (int, error) {
	return 0, errNotImplemented
}

// Select executes a SELECT query and stores the result in rows.
// The argument passed to rows can be one of the following:
//  (a) A pointer to a slice of structs; or
//  (b) A pointer to a slice of struct pointers; or
//  (c) A pointer to a struct.
// When rows is a pointer to a slice, it is populated with
// one item for each row returned by the SELECT query.
//
// When rows is a pointer to a struct, it is populated with
// the first row returned from the query. This is a good
// option when the query will only return one row.
//
// Select returns the number of rows returned by the SELECT
// query.
func Select(db DB, rows interface{}, sql string, args ...interface{}) (int, error) {
	return 0, errNotImplemented
}

// Dialect is an interface used to handle differences
// in SQL dialects.
type Dialect interface {
	// Name of the dialect. This name is used as
	// a key for caching, so if If two dialects have
	// the same name, then they should be identical.
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

// DialectFor returns the dialect for the specified database driver.
// If name is blank, then the dialect returned is for the first
// driver returned by sql.Drivers(). If only one SQL driver has
// been loaded by the calling program then this will return the
// correct dialect. If the driver name is unknown, the default
// dialect is returned.
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
func DialectFor(name string) Dialect {
	return nil
}

// Convention provides naming convention methods for
// inferring database column names from Go struct field names.
type Convention interface {
	// The name of the convention. This name is used as
	// a key for caching, so if If two conventions have
	// the same name, then they should be identical.
	Name() string

	// ColumnName returns the name of a database column based
	// on the name of a Go struct field.
	ColumnName(fieldName string) string

	// Join joins a prefix with a name to form a column name.
	// Used for naming columns based on fields within embedded
	// structures. The column name will be based on the name of
	// the Go struct field and its enclosing embedded struct fields.
	Join(prefix, name string) string
}

// ConventionSnake is the default, 'snake_case' naming convention
var ConventionSnake Convention

// ConventionSame is a naming convention where the column name
// is identical to the Go struct field name.
var ConventionSame Convention

// Default is the default schema, which can be modified as required.
//
// The default schema has sensible defaults. If not explicitly
// specified, the dialect is determined by the first of the SQL database
// drivers to be loaded. If the program only uses one database driver,
// then the default schema will use the correct dialect.
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
}

// Insert inserts a row. If the row has an auto-increment column
// defined, then the generated value is retrieved and inserted into the
// row. (If the database driver provides the necessary support).
func (s Schema) Insert(db DB, row interface{}, sql string) error {
	return errNotImplemented
}

// Update updates a row. Returns the number of rows affected,
// which should be zero or one.
func (s Schema) Update(db DB, row interface{}, sql string, args ...interface{}) (int, error) {
	return 0, errNotImplemented
}

// Delete deletes a row. Returns the number of rows affected,
// which should be zero or one.
func (s Schema) Delete(db DB, row interface{}, sql string, args ...interface{}) (int, error) {
	return 0, errNotImplemented
}

// Prepare creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the returned
// statement.
func (s Schema) Prepare(row interface{}, sql string) (*Stmt, error) {
	return nil, errNotImplemented
}

// Select executes a SELECT query and stores the result in rows.
// The argument passed to rows can be one of the following:
//  A pointer to an array of structs; or
//  a pointer to an array of struct pointers; or
//  a pointer to a struct.
// When rows is a pointer to an array it is populated with
// one item for each row returned by the SELECT query.
//
// When rows is a pointer to a struct, it is populated with
// the first row returned from the query. This is a good
// option when the query will only return one row.
//
// Select returns the number of rows returned by the SELECT
// query.
func (s Schema) Select(db DB, rows interface{}, sql string, args ...interface{}) (int, error) {
	return 0, errNotImplemented
}

type Stmt struct {
	reserve int
}

// Prepare creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the returned
// statement. The row parameter should be a structure or a pointer to a structure
// and is used to determine the type of the row used when executing the statement.
func Prepare(row interface{}, sql string) (*Stmt, error) {
	return nil, errNotImplemented
}

// Select executes the prepared query statement with the given arguments and
// returns the query results in rows. If rows is a pointer to a slice of structs
// then one item is added to the slice for each row returned by the query. If row
// is a pointer to a struct then that struct is filled with the result of the first
// row returned by the query. In both cases Select returns the number of rows returned
// by the query.
func (s *Stmt) Select(db DB, rows interface{}, args ...interface{}) (int, error) {
	return 0, errNotImplemented
}

// Exec executes the prepared statement with the given row and optional arguments.
// It returns the number of rows affected by the statement.
//
// If the statement is an INSERT statement and the row has an auto-increment field,
// then the row is updated with the value of the auto-increment column as long as
// the SQL driver supports this functionality.
func (s *Stmt) Exec(db DB, row interface{}, args ...interface{}) (int, error) {
	return 0, errNotImplemented
}
