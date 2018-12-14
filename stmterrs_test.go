package sqlr

// tests for statement error conditions

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

type FakeDB struct {
	execErr         error
	rowsAffected    int64
	rowsAffectedErr error
	lastInsertId    int64
	lastInsertIdErr error
	queryErr        error
}

func (db *FakeDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	if db.execErr != nil {
		return nil, db.execErr
	}
	return db, nil
}

func (db *FakeDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.Exec(query, args...)
}

func (db *FakeDB) LastInsertId() (int64, error) {
	return db.lastInsertId, db.lastInsertIdErr
}

func (db *FakeDB) RowsAffected() (int64, error) {
	return db.rowsAffected, db.rowsAffectedErr
}

func (db *FakeDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, db.queryErr
}

func (db *FakeDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.Query(query, args...)
}

func TestSelectStmt1Errors(t *testing.T) {
	type Row struct {
		ID   int64 `sql:"primary key autoincrement"`
		Name string
	}
	type NotARow struct {
		ID            int `sql:"primary key autoincrement"`
		SomethingElse int
	}
	tests := []struct {
		row        interface{}
		sql        string
		errPrepare string
		errExec    string
	}{
		{
			row:     Row{},
			sql:     "tablename",
			errExec: "expected rows to be *[]github.com/jjeffery/sqlr.Row, *[]*github.com/jjeffery/sqlr.Row, or *github.com/jjeffery/sqlr.Row",
		},
		{
			row:     nil,
			sql:     "tablename",
			errExec: "nil pointer",
		},
		{
			row:     (*Row)(nil),
			sql:     "tablename",
			errExec: "nil pointer",
		},
		{
			row:     &NotARow{},
			sql:     "tablename",
			errExec: "expected rows to be *[]github.com/jjeffery/sqlr.Row, *[]*github.com/jjeffery/sqlr.Row, or *github.com/jjeffery/sqlr.Row",
		},
		{
			row:        Row{},
			sql:        "select {} from {} where {}",
			errPrepare: `cannot expand "{}" in "select from" clause`,
		},
		{
			row:        Row{},
			sql:        "select {dodgy¥} from xx where {}",
			errPrepare: `cannot expand "dodgy¥" in "select columns" clause: unrecognised input near "¥"`,
		},
	}

	for i, tt := range tests {
		schema := NewSchema()
		ctx := context.Background()
		stmt, err := schema.Prepare(Row{}, tt.sql)
		if tt.errPrepare == "" {
			if err != nil {
				t.Errorf("%d: expected no error, got %v", i, err)
			}
		} else {
			if err == nil {
				t.Errorf("%d: expected %q, got nil", i, tt.errPrepare)
			} else if err.Error() != tt.errPrepare {
				t.Errorf("%d: expected %q, got %v", i, tt.errPrepare, err)
			}
		}
		db := &FakeDB{}

		if err == nil {
			_, err = stmt.selectRows(ctx, db, tt.row)
			if err == nil || err.Error() != tt.errExec {
				t.Errorf("test case %d:\nwant=%q\ngot=%q", i, tt.errExec, err)
			}
		}
	}
}

func TestSelectStmt2Errors(t *testing.T) {
	type Row struct {
		ID   int64 `sql:"primary key autoincrement"`
		Name string
	}
	type NotARow struct {
		ID            int `sql:"primary key autoincrement"`
		SomethingElse int
	}
	const errorTypePtr = "expected rows to be " +
		"*[]github.com/jjeffery/sqlr.Row, " +
		"*[]*github.com/jjeffery/sqlr.Row, or " +
		"*github.com/jjeffery/sqlr.Row"
	var invalidSlice []NotARow
	var validRows []Row
	tests := []struct {
		dest       interface{}
		sql        string
		args       []interface{}
		queryErr   error
		errText    string
		errPrepare string
	}{
		{
			dest:    []Row{},
			errText: errorTypePtr,
		},
		{
			dest:    make([]Row, 0),
			errText: errorTypePtr,
		},
		{
			dest:    nil,
			errText: "nil pointer",
		},
		{
			dest:    (*Row)(nil),
			errText: "nil pointer",
		},
		{
			dest:    &NotARow{},
			errText: errorTypePtr,
		},
		{
			dest:    &invalidSlice,
			errText: errorTypePtr,
		},
		{
			dest:       &validRows,
			sql:        "select {} from table {} where {}",
			errPrepare: `cannot expand "{}" in "select from" clause`,
			args:       []interface{}{},
		},
		{
			dest:     &validRows,
			sql:      "select {} from table where name=?",
			queryErr: errors.New("test query error"),
			errText:  `test query error`,
			args:     []interface{}{"somevalue"},
		},
	}

	for i, tt := range tests {
		schema := NewSchema()
		ctx := context.Background()
		stmt, err := schema.Prepare(Row{}, tt.sql)
		if tt.errPrepare == "" {
			if err != nil {
				t.Errorf("%d: expected no error, got %q", i, err)
			}
		} else {
			if err == nil {
				t.Errorf("%d: expected %q, got no error", i, tt.errPrepare)
			} else if err.Error() != tt.errPrepare {
				t.Errorf("%d:\nexpected %q,\ngot %q", i, tt.errPrepare, err)
			}
		}
		if err != nil {
			continue
		}

		db := &FakeDB{queryErr: tt.queryErr}

		_, err = stmt.selectRows(ctx, db, tt.dest, tt.args...)
		if err == nil || err.Error() != tt.errText {
			t.Errorf("%d: want=%q\ngot=%q", i, tt.errText, err)
		}
	}
}

func TestInsertRowStmtErrors(t *testing.T) {
	type Row struct {
		ID   int64 `sql:"primary key autoincrement"`
		Name string
	}
	type NotARow struct {
		ID            int `sql:"primary key autoincrement"`
		SomethingElse int
	}
	tests := []struct {
		sql             string
		row             interface{}
		execErr         error
		lastInsertIdErr error
		errText         string
	}{
		{
			sql:     "insert into tablename({}) values({})",
			row:     &Row{},
			execErr: errors.New("test error condition"),
			errText: "test error condition",
		},
		{
			sql:     "insert into table values {}",
			row:     &Row{},
			errText: `cannot expand "insert values" clause because "insert columns" clause is missing`,
		},
		{
			sql:     "insert into table({}) values({all})",
			row:     &Row{},
			errText: `columns for "insert values" clause must match the "insert columns" clause`,
		},
	}

	for i, tt := range tests {
		schema := NewSchema()
		ctx := context.Background()
		db := &FakeDB{
			execErr:         tt.execErr,
			lastInsertIdErr: tt.lastInsertIdErr,
		}

		sess := NewSession(ctx, db, schema)
		_, err := sess.Row(tt.row).Exec(tt.sql)
		if err == nil || err.Error() != tt.errText {
			t.Errorf("%d: expected=%q, actual=%v", i, tt.errText, err)
		}
	}
}

func TestExecRowStmtErrors(t *testing.T) {
	type Row struct {
		ID   int64 `sql:"primary key autoincrement"`
		Name string
	}
	type NotARow struct {
		ID            int `sql:"primary key autoincrement"`
		SomethingElse int
	}
	tests := []struct {
		sql     string
		row     interface{}
		execErr error
		errText string
	}{
		{
			sql:     "update tablename set {} where {}",
			row:     &Row{},
			execErr: errors.New("test error condition"),
			errText: "test error condition",
		},
		{
			sql:     "update table {}",
			row:     &Row{},
			errText: `cannot expand "{}" in "update table" clause`,
		},
	}

	for i, tt := range tests {
		schema := NewSchema()
		ctx := context.Background()
		db := &FakeDB{
			execErr: tt.execErr,
		}
		sess := NewSession(ctx, db, schema)

		_, err := sess.Row(tt.row).Exec(tt.sql)
		if err == nil || err.Error() != tt.errText {
			t.Errorf("%d: expected=%q, actual=%q", i, tt.errText, err)
		}
	}
}

func TestInvalidStmts(t *testing.T) {
	ctx := context.Background()
	type Row struct {
		ID     int64 `sql:"primary key"`
		Name   string
		Number int
	}

	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	schema := NewSchema(ForDB(db))
	sess := NewSession(ctx, db, schema)
	defer sess.Close()

	var row Row
	var notRow int

	tests := []struct {
		fn   func() (interface{}, error)
		want string
	}{
		{
			fn:   func() (interface{}, error) { return sess.Row(&notRow).Exec("insert into rows({}) values({})") },
			want: `expected row type to be a struct, found int`,
		},
		{
			fn:   func() (interface{}, error) { return sess.Row(&row).Exec("insert into xyz values({})") },
			want: `cannot expand "insert values" clause because "insert columns" clause is missing`,
		},
		{
			fn:   func() (interface{}, error) { return sess.Row(&row).Exec("insert into xyz({}) values({pk})") },
			want: `columns for "insert values" clause must match the "insert columns" clause`,
		},
		{
			fn:   func() (interface{}, error) { return sess.Row(&row).Exec("update {} this is not valid SQL") },
			want: `cannot expand "{}" in "update table" clause`,
		},
		{
			fn:   func() (interface{}, error) { return sess.Row(&row).Exec("update rows set {} where {} and number=?") },
			want: `expected arg count=1, actual=0`,
		},
		{
			fn:   func() (interface{}, error) { return sess.Row(&row).Exec("delete from rows where {}") },
			want: `no such table: rows`,
		},
		{
			fn:   func() (interface{}, error) { return sess.Row(&row).Exec("select {alias} from rows") },
			want: `cannot expand "alias" in "select columns" clause: missing ident after 'alias'`,
		},
		{
			fn:   func() (interface{}, error) { return sess.Select(&row, "select {'col1} from rows") },
			want: `cannot expand "'col1" in "select columns" clause: unrecognised input near "'col1"`,
		},
		{
			fn:   func() (interface{}, error) { return sess.Select(&notRow, "select {} from rows") },
			want: `expected row type to be a struct, found int`,
		},
	}

	for i, tt := range tests {
		_, err := tt.fn()
		if err != nil {
			if tt.want != err.Error() {
				t.Errorf("%d: want %s, got %v", i, tt.want, err.Error())
			}
			continue
		}
		t.Errorf("%d: want %s, got nil", i, tt.want)
	}
}

func TestInvalidPrepare(t *testing.T) {
	schema := NewSchema(WithDialect(ANSISQL))
	var notRow []int
	_, err := schema.Prepare(notRow, "select {} from rows")
	want := `expected row type to be a struct, found int`
	if err != nil {
		if want != err.Error() {
			t.Errorf("want %s, got %v", want, err)
		}
	} else {
		t.Errorf("want %s, got nil", want)
	}
}
