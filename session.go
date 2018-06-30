package sqlr

import (
	"context"
)

// A Session is a request-scoped database session. It can execute
// queries and it can construct strongly-typed query functions.
type Session struct {
	context context.Context
	querier Querier
	schema  *Schema
}

// NewSession creates a new, request-scoped session for performing queries.
func (s *Schema) NewSession(ctx context.Context, querier Querier) *Session {
	return NewSession(ctx, querier, s)
}

// NewSession returns a new, request-scoped session.
func NewSession(ctx context.Context, querier Querier, schema *Schema) *Session {
	if ctx == nil {
		ctx = context.Background()
	}
	if querier == nil {
		panic("querier cannot be nil")
	}
	if schema == nil {
		panic("schema cannot be nil")
	}
	return &Session{
		context: ctx,
		querier: querier,
		schema:  schema,
	}
}

// Exec executes the query with the given row and optional arguments.
// It returns the number of rows affected by the statement.
//
// If the statement is an INSERT statement and the row has an auto-increment field,
// then the row is updated with the value of the auto-increment column, as long as
// the SQL driver supports this functionality.
func (sess *Session) Exec(row interface{}, query string, args ...interface{}) (int, error) {
	stmt, err := sess.schema.Prepare(row, query)
	if err != nil {
		return 0, err
	}
	return stmt.Exec(sess.context, sess.querier, row, args...)
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
func (sess *Session) Select(rows interface{}, query string, args ...interface{}) (int, error) {
	stmt, err := sess.schema.Prepare(rows, query)
	if err != nil {
		return 0, err
	}
	return stmt.Select(sess.context, sess.querier, rows, args...)
}

// Querier returns the database querier used for this session.
func (sess *Session) Querier() Querier {
	return sess.querier
}

// Context returns the request-scoped context used by this session.
func (sess *Session) Context() context.Context {
	return sess.context
}

// Schema returns the schema used for this session.
func (sess *Session) Schema() *Schema {
	return sess.schema
}
