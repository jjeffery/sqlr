package sqlr

import (
	"errors"
	"fmt"
	"reflect"
)

// RowFunc creates useful, strongly-typed, data access functions
// for a specified struct type that represents a row in a database table.
// The functions created are attached to a session.
//
// Although a little obscure, RowFunc makes it easy to create data access
// objects.
//
// TODO: example is not very idiomatic, see https://todo.todo/todo for a
// more complete, idiomatic example.
type RowFunc struct {
	sess    *Session
	builder rowFuncBuilder
}

// BUG(jpj): RowFunc only works with rows that have a single primary key column. This may change in future.

// NewRowFunc creates a SessionRow. TODO: better comment here please.
func NewRowFunc(sess *Session, rowType interface{}, opts ...RowFuncOpt) *RowFunc {
	var t reflect.Type
	if tt, ok := rowType.(reflect.Type); ok {
		// passed a reflect.Type
		t = tt
	} else {
		t = reflect.TypeOf(rowType)
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("expected row to be a struct: got %v", t))
	}

	rf := &RowFunc{
		sess: sess,
		builder: rowFuncBuilder{
			rowType: t,
			schema:  sess.schema,
		},
	}
	for _, opt := range opts {
		opt(rf)
	}

	if rf.builder.tableName == "" {
		rf.builder.tableName = sess.schema.convention.Convert(rf.builder.rowType.Name())
	}
	if rf.builder.singular == "" {
		rf.builder.singular = rf.builder.rowType.Name()
	}
	if rf.builder.plural == "" {
		// TODO: is it worth including a more sophisticated inflector?
		rf.builder.plural = rf.builder.singular + "s"
	}

	return rf
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
func (sr *RowFunc) MakeQuery(funcPtr ...interface{}) error {
	for _, fp := range funcPtr {
		if err := sr.makeQueryFunc(fp); err != nil {
			return err
		}
	}
	return nil
}

// MustMakeQuery does the same thing as MakeQuery, but panics if an error is
// encountered.
func (sr *RowFunc) MustMakeQuery(funcPtr ...interface{}) {
	if err := sr.MakeQuery(funcPtr...); err != nil {
		panic(err)
	}
}

func (sr *RowFunc) makeQueryFunc(funcPtr interface{}) error {
	funcPtrValue := reflect.ValueOf(funcPtr)
	if funcPtrValue.Type().Kind() != reflect.Ptr {
		return errors.New("expected pointer to function")
	}
	funcValue := funcPtrValue.Elem()
	funcType := funcValue.Type()
	if funcType.Kind() != reflect.Func {
		return errors.New("expected pointer to function")
	}

	queryFuncFactory, err := sr.builder.makeQuery(funcType)
	if err != nil {
		return err
	}

	queryFunc := queryFuncFactory(sr.sess)
	funcValue.Set(queryFunc)
	return nil
}

var wellKnownTypes = struct {
	errorType            reflect.Type
	stringType           reflect.Type
	sliceOfInterfaceType reflect.Type
}{
	errorType:            reflect.TypeOf((*error)(nil)).Elem(),
	stringType:           reflect.TypeOf((*string)(nil)).Elem(),
	sliceOfInterfaceType: reflect.SliceOf(reflect.TypeOf((*interface{})(nil)).Elem()),
}

// RowFuncOpt is an option used when creating a new RowFunc.
type RowFuncOpt func(*RowFunc)

// TableName specifies the database table name. If not specified, the table
// name is derived from the row type name, using the schema's naming convention.
func TableName(name string) RowFuncOpt {
	return func(s *RowFunc) {
		s.builder.tableName = name
	}
}

// Singular specifies the name to use when describing a single row object.
// This is used for constructing useful, readable error messages. If not
// specified, singular is the row's type name.
func Singular(name string) RowFuncOpt {
	return func(s *RowFunc) {
		s.builder.singular = name
	}
}

// Plural specifies the name to use when describing more than one row object.
// This is used for constructing useful, readable error messages. If not specified,
// plural is derived from the singular name (which usually involves just adding an 's').
func Plural(name string) RowFuncOpt {
	return func(s *RowFunc) {
		s.builder.singular = name
	}
}
