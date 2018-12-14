package sqlr

import (
	"context"
	"database/sql"
)

// The Querier interface defines the SQL database access methods used by this package.
//
// The *DB, *Tx and *Conn types in the standard library package "database/sql" all implement this interface.
// This interface is based on https://godoc.org/github.com/golang-sql/sqlexp#Querier.
type Querier interface {
	// ExecContext executes a query without returning any rows.
	// The args are for any placeholder parameters in the query.
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// QueryContext executes a query that returns rows, typically a SELECT.
	// The args are for any placeholder parameters in the query.
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

var (
	_ Querier = &sql.DB{}
	_ Querier = &sql.Tx{}
	_ Querier = &sql.Conn{}
)
