package sqlrow

import (
	"errors"
)

var (
	errNotImplemented = errors.New("not implemented")
)

// Insert inserts a row. If the row has an auto-increment column
// defined, then the generated value is retrieved and inserted into the
// row. (If the database driver provides the necessary support).
func Insert(db DB, row interface{}, sql string) error {
	return Default.Insert(db, row, sql)
}

// Update updates a row. Returns the number of rows affected,
// which should be zero or one.
func Update(db DB, row interface{}, sql string, args ...interface{}) (int, error) {
	return Default.Update(db, row, sql, args...)
}

// Delete deletes a row. Returns the number of rows affected,
// which should be zero or one.
func Delete(db DB, row interface{}, sql string, args ...interface{}) (int, error) {
	return Default.Delete(db, row, sql, args...)
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
	return Default.Select(db, rows, sql, args...)
}

// Prepare creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the returned
// statement. The row parameter should be a structure or a pointer to a structure
// and is used to determine the type of the row used when executing the statement.
func Prepare(row interface{}, sql string) (*Stmt, error) {
	return Default.Prepare(row, sql)
}
