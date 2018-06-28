package sqlr

import (
	"database/sql"
)

// The Querier interface defines the SQL database access methods used by this package.
//
// The *DB and *Tx types in the standard library package "database/sql"
// both implement this interface.
type Querier interface {
	// Exec executes a query without returning any rows.
	// The args are for any placeholder parameters in the query.
	Exec(query string, args ...interface{}) (sql.Result, error)

	// Query executes a query that returns rows, typically a SELECT.
	// The args are for any placeholder parameters in the query.
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// DB is the deprecated name for Querier.
//
// Deprecated: use Querier instead.
type DB = Querier
