package sqlstmt

import (
	"database/sql"
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

// Logger wraps a single method, Print, which prints a message
// for diagnostic purposes. Any implementation of this interface must
// support concurrent access by multiple goroutines.
//
// The Logger type in the standard library package "log" implements
// this interface.
//
// Note that according to the Go naming conventions for single-method
// interfaces, this interface should be called "Printer". The name
// "Logger" has been chosen because it better reflects the intention
// of this interface, and it matches the name of the Logger type in
// the log package.
type Logger interface {
	Print(v ...interface{})
}
