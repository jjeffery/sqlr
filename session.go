package sqlr

import (
	"context"
	"errors"
	"reflect"
)

// A Session is a request-scoped database session. It can execute
// queries and it can construct strongly-typed query functions.
type Session struct {
	context context.Context
	cancel  func()
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
	ctx, cancel := context.WithCancel(ctx)
	return &Session{
		context: ctx,
		cancel:  cancel,
		querier: querier,
		schema:  schema,
	}
}

// Close releases resources associated with the session. Any attempt to
// query using the session will fail after Close has been called.
//
// Because a session is request-scoped, it should never be used once
// a request has completed. Calling a session's Close method at the end
// of a request is an effective way to release resources associated with
// the session and to ensure that it can no longer be used.
//
// Close implements the io.Closer interface. It always returns nil.
func (sess *Session) Close() error {
	sess.cancel()
	return nil
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

// MakeQuery makes one or more functions that can be used to query in a type-safe manner.
// Each funcPtr is a pointer to a function that will be created by this function.
//
// If "Row" is the row type, and "RowID" is the primary key type for Row objects, then the
// function can be any one of the following signatures:
//  // Get one row based on its ID
//  func(id RowID) (*Row, error)
//
//  // Get multiple rows given multiple IDs
//  func(ids []RowID) ([]*Row, error)
//  func(ids ...RowID) ([]*Row, error)
//
//  // Get one row returning a thunk: batches multiple requests into
//  // one query using the dataloader pattern
//  func(id RowID) func() (*Row, error)
//
//  // Execute a query that will return multiple Row objects
//  func(query string, args ...interface{}) ([]*Row, error)
//
//  // Execute a query that will return a single Row object.
//  func(query string, args ...interface{}) (*Row, error)
//
// If any of the funcPtr arguments are not pointers to a function, or do not fit
// one of the known function prototypes, then this function will panic. The reasoning
// here is that if MakeQuery succeeds in a unit test, then it will always succeed
// in production.
func (sess *Session) MakeQuery(funcPtr ...interface{}) error {
	for _, fp := range funcPtr {
		if err := sess.makeQueryFunc(fp); err != nil {
			return err
		}
	}
	return nil
}

// MustMakeQuery does the same thing as MakeQuery, but panics if an error is
// encountered.
func (sess *Session) MustMakeQuery(funcPtr ...interface{}) {
	if err := sess.MakeQuery(funcPtr...); err != nil {
		panic(err)
	}
}

func (sess *Session) makeQueryFunc(funcPtr interface{}) error {
	funcPtrValue := reflect.ValueOf(funcPtr)
	if funcPtrValue.Type().Kind() != reflect.Ptr {
		return errors.New("expected pointer to function")
	}
	funcValue := funcPtrValue.Elem()
	funcType := funcValue.Type()
	if funcType.Kind() != reflect.Func {
		return errors.New("expected pointer to function")
	}

	queryFuncFactory, err := makeQuery(funcType, sess.schema)
	if err != nil {
		return err
	}

	queryFunc := queryFuncFactory(sess)
	funcValue.Set(queryFunc)
	return nil
}
