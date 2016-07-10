package sqlf

import (
	"database/sql"
)

// Execer implements a single method, Exec, which executes a
// query without returning any rows. The args for any parameter
// placeholders in the query.
type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// Queryer implements a single method, Query, which executes
// a query that returns zero, one or more rows. The args are
// for any parameter placeholders in the query.
type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}
