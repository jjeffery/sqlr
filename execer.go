package sqlstmt

import (
	"database/sql"
)

// Execer implements a single method, Exec, which executes a
// query without returning any rows. The args are for any parameter
// placeholders in the query.
//
// The *DB and *Tx types in the standard library package "database/sql"
// both implement this interface.
type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// Queryer implements a single method, Query, which executes
// a query that returns zero, one or more rows. The args are
// for any parameter placeholders in the query.
//
// The *DB and *Tx types in the standard library package "database/sql"
// both implement this interface.
type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// Logger implements a single method, Print, which prints a message
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
