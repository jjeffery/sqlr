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

// SQLLogger is an interface for logging SQL statements executed
// by the sqlstmt package.
type SQLLogger interface {
	// LogSQL is called by the sqlstmt package after it executes
	// an SQL query or statement.
	//
	// The query and args variables provide the query and associated
	// arguments supplied to the database server.  The rowsAffected
	// and err variables provide a summary of the query results.
	// If the number of rows affected cannot be determined for any reason,
	// then rowsAffected is set to -1.
	LogSQL(query string, args []interface{}, rowsAffected int, err error)
}

// The SQLLoggerFunc type is an adapter to allow the use of ordinary
// functions as SQLLoggers. If f is a function with the appropriate
// signature, SQLLoggerFunc(f) is an SQLLogger that calls f.
type SQLLoggerFunc func(query string, args []interface{}, rowsAffected int, err error)

// LogSQL calls f(query, args, rowsAffected, err).
func (f SQLLoggerFunc) LogSQL(query string, args []interface{}, rowsAffected int, err error) {
	f(query, args, rowsAffected, err)
}
