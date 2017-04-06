package statement

import (
	"errors"
	"reflect"
	"testing"
)

func TestCache(t *testing.T) {
	// clear the stmt cache before and after the test
	clearStmtCache()
	defer clearStmtCache()

	// function to create a factory function
	factoryFor := func(stmt *Stmt, err error) func(reflect.Type, string) (*Stmt, error) {
		return func(rowType reflect.Type, sql string) (*Stmt, error) {
			return stmt, err
		}
	}

	// Expected statements and error
	expected := struct {
		stmts []*Stmt
		errs  []error
	}{
		stmts: []*Stmt{&Stmt{}, &Stmt{}, &Stmt{}},
		errs:  []error{errors.New("1"), errors.New("2")},
	}

	type rowType1 struct {
		id   int
		col2 int
	}
	type rowType2 struct {
		id   string
		col2 float32
	}

	tests := []struct {
		rowType reflect.Type
		sql     string
		factory func(reflect.Type, string) (*Stmt, error)
		stmt    *Stmt
		err     error
	}{
		{
			rowType: reflect.TypeOf(rowType1{}),
			sql:     "sql 0",
			factory: factoryFor(expected.stmts[0], nil),
			stmt:    expected.stmts[0],
			err:     nil,
		},
		{
			rowType: reflect.TypeOf(rowType1{}),
			sql:     "sql 0",
			// should return the same stmt as previous test case, so factory returns error
			factory: factoryFor(nil, errors.New("expected not to be called")),
			stmt:    expected.stmts[0],
			err:     nil,
		},
		{
			rowType: reflect.TypeOf(rowType1{}),
			sql:     "sql 1",
			// should return the same stmt as previous test case, so factory returns error
			factory: factoryFor(expected.stmts[1], nil),
			stmt:    expected.stmts[1],
			err:     nil,
		},
		{
			rowType: reflect.TypeOf(rowType2{}),
			sql:     "sql 0",
			// should return the same stmt as previous test case, so factory returns error
			factory: factoryFor(expected.stmts[2], nil),
			stmt:    expected.stmts[2],
			err:     nil,
		},
		{
			rowType: reflect.TypeOf(rowType1{}),
			sql:     "sql err",
			// should return the same stmt as previous test case, so factory returns error
			factory: factoryFor(nil, expected.errs[0]),
			stmt:    nil,
			err:     expected.errs[0],
		},
	}

	for i, tt := range tests {
		stmt, err := getStmtFromCache(tt.rowType, tt.sql, tt.factory)
		if want, got := tt.stmt, stmt; want != got {
			t.Errorf("%d: want=%v got=%v", i, want, got)
		}
		if want, got := tt.err, err; want != got {
			t.Errorf("%d: want=%v got=%v", i, want, got)
		}
	}
}
