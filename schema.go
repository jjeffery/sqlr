package sqlrow

import (
	"github.com/jjeffery/sqlrow/private/dialect"
	"github.com/jjeffery/sqlrow/private/naming"
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
}

// Insert inserts a row. If the row has an auto-increment column
// defined, then the generated value is retrieved and inserted into the
// row. (If the database driver provides the necessary support).
func (s Schema) Insert(db DB, row interface{}, sql string) error {
	_, err := s.execCommon(db, row, checkSQL(sql, insertFormat), nil)
	return err
}

// Update updates a row. Returns the number of rows affected,
// which should be zero or one.
func (s Schema) Update(db DB, row interface{}, sql string, args ...interface{}) (int, error) {
	return s.execCommon(db, row, checkSQL(sql, updateFormat), args)
}

// Delete deletes a row. Returns the number of rows affected,
// which should be zero or one.
func (s Schema) Delete(db DB, row interface{}, sql string, args ...interface{}) (int, error) {
	return s.execCommon(db, row, checkSQL(sql, deleteFormat), args)
}

// Prepare creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the returned
// statement.
func (s Schema) Prepare(row interface{}, sql string) (*Stmt, error) {
	rowType, err := inferRowType(row, "row")
	if err != nil {
		return nil, err
	}
	// does not reference the cache because the calling program is
	// taking responsibility for keeping track of this stmt
	return newStmt(s.dialect(), s.convention(), rowType, sql)
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
	sql = checkSQL(sql, selectFormat)
	rowType, err := inferRowType(rows, "rows")
	if err != nil {
		return 0, err
	}
	//sql = checkSQL(sql, selectFormat)
	stmt, err := getStmtFromCache(s.dialect(), s.convention(), rowType, sql)
	if err != nil {
		return 0, err
	}
	return stmt.Select(db, rows, args...)
}

func (s Schema) dialect() Dialect {
	if s.Dialect != nil {
		return s.Dialect
	}
	if Default.Dialect != nil {
		return Default.Dialect
	}
	return dialect.For("default")
}

func (s Schema) convention() Convention {
	if s.Convention != nil {
		return s.Convention
	}
	if Default.Convention != nil {
		return Default.Convention
	}
	return naming.Snake
}

func (s Schema) execCommon(db DB, row interface{}, sql string, args []interface{}) (int, error) {
	rowType, err := inferRowType(row, "row")
	if err != nil {
		return 0, err
	}
	stmt, err := getStmtFromCache(s.dialect(), s.convention(), rowType, sql)
	if err != nil {
		return 0, err
	}
	return stmt.Exec(db, row, args...)
}
