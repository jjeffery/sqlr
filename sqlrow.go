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
