package sqlr

import (
	"context"
	"database/sql"
)

// The Querier interface defines the SQL database access methods used by this package.
//
// The *DB, *Tx and *Conn types in the standard library package "database/sql" all implement this interface.
type Querier interface {
	// ExecContext executes a query without returning any rows.
	// The args are for any placeholder parameters in the query.
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// QueryContext executes a query that returns rows, typically a SELECT.
	// The args are for any placeholder parameters in the query.
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// The DB interface defines the SQL database access methods used by this package.
//
// Deprecated: use Querier instead.
type DB interface {
	Querier

	// Exec executes a query without returning any rows.
	// The args are for any placeholder parameters in the query.
	//
	// Deprecated: use ExecContext instead
	Exec(query string, args ...interface{}) (sql.Result, error)

	// Query executes a query that returns rows, typically a SELECT.
	// The args are for any placeholder parameters in the query.
	//
	// Deprecated: use QueryContext instead.
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

var (
	_ Querier = &sql.DB{}
	_ Querier = &sql.Tx{}
	_ Querier = &sql.Conn{}

	_ DB = &sql.DB{}
	_ DB = &sql.Tx{}
)
