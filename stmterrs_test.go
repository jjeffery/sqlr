package sqlstmt

// tests for error conditions

import (
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

func (db *FakeDB) LastInsertId() (int64, error) {
	return db.lastInsertId, db.lastInsertIdErr
}

func (db *FakeDB) RowsAffected() (int64, error) {
	return db.rowsAffected, db.rowsAffectedErr
}

func (db *FakeDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, db.queryErr
}

func TestGetRowStmtErrors(t *testing.T) {
	type Row struct {
		ID   int64 `sql:"primary key autoincrement"`
		Name string
	}
	type NotARow struct {
		ID            int `sql:"primary key autoincrement"`
		SomethingElse int
	}
	tests := []struct {
		row     interface{}
		sql     string
		errText string
	}{
		{
			row:     Row{},
			sql:     "tablename",
			errText: "expected dest to be *github.com/jjeffery/sqlstmt.Row",
		},
		{
			row:     nil,
			sql:     "tablename",
			errText: "expected dest to be *github.com/jjeffery/sqlstmt.Row",
		},
		{
			row:     (*Row)(nil),
			sql:     "tablename",
			errText: "nil pointer passed",
		},
		{
			row:     &NotARow{},
			sql:     "tablename",
			errText: "expected dest to be *github.com/jjeffery/sqlstmt.Row",
		},
		{
			row:     Row{},
			sql:     "select {} from {} where {}",
			errText: `cannot expand "{}" in "select from" clause`,
		},
		{
			row:     Row{},
			sql:     "select {dodgy!} from xx where {}",
			errText: `cannot expand "dodgy!" in "select columns" clause: illegal char: "!"`,
		},
	}

	for _, tt := range tests {
		stmt := NewGetRowStmt(Row{}, tt.sql)
		db := &FakeDB{}

		_, err := stmt.Get(db, tt.row)
		if err == nil || err.Error() != tt.errText {
			t.Errorf("err=%q", err)
		}
	}
}

func TestSelectStmtErrors(t *testing.T) {
	type Row struct {
		ID   int64 `sql:"primary key autoincrement"`
		Name string
	}
	type NotARow struct {
		ID            int `sql:"primary key autoincrement"`
		SomethingElse int
	}
	const errorTypePtr = "Expected dest to be pointer to " +
		"[]github.com/jjeffery/sqlstmt.Row or " +
		"[]*github.com/jjeffery/sqlstmt.Row"
	var invalidSlice []NotARow
	var validRows []Row
	tests := []struct {
		dest     interface{}
		sql      string
		args     []interface{}
		queryErr error
		errText  string
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
			errText: errorTypePtr,
		},
		{
			dest:    (*Row)(nil),
			errText: "Select: nil pointer passed as dest",
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
			dest:    &validRows,
			sql:     "select {} from table {} where {}",
			errText: `cannot expand "{}" in "select from" clause`,
			args:    []interface{}{},
		},
		{
			dest:     &validRows,
			sql:      "select {} from table where name=?",
			queryErr: errors.New("test query error"),
			errText:  `test query error`,
			args:     []interface{}{"somevalue"},
		},
		{
			dest:    &validRows,
			sql:     "select {} from table where {}",
			errText: `unexpected inputs in query`,
			args:    []interface{}{"somevalue"},
		},
	}

	for _, tt := range tests {
		stmt := NewSelectStmt(Row{}, tt.sql)
		db := &FakeDB{queryErr: tt.queryErr}

		err := stmt.Select(db, tt.dest, tt.args...)
		if err == nil || err.Error() != tt.errText {
			t.Errorf("expected=%q, actual=%q", tt.errText, err)
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
			sql:     "tablename",
			row:     Row{},
			errText: "cannot set auto-increment value for type Row",
		},
		{
			sql:     "tablename",
			row:     &Row{},
			execErr: errors.New("test error condition"),
			errText: "test error condition",
		},
		{
			sql:             "tablename",
			row:             &Row{},
			lastInsertIdErr: errors.New("test LastInsertId"),
			errText:         "test LastInsertId",
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

	for _, tt := range tests {
		stmt := NewInsertRowStmt(Row{}, tt.sql)
		db := &FakeDB{
			execErr:         tt.execErr,
			lastInsertIdErr: tt.lastInsertIdErr,
		}

		err := stmt.Exec(db, tt.row)
		if err == nil || err.Error() != tt.errText {
			t.Errorf("expected=%q, actual=%q", tt.errText, err)
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
		sql             string
		row             interface{}
		execErr         error
		rowsAffectedErr error
		errText         string
	}{
		{
			sql:     "tablename",
			row:     &Row{},
			execErr: errors.New("test error condition"),
			errText: "test error condition",
		},
		{
			sql:             "tablename",
			row:             &Row{},
			rowsAffectedErr: errors.New("test RowsAffected"),
			errText:         "test RowsAffected",
		},
		{
			sql:     "update table {}",
			row:     &Row{},
			errText: `cannot expand "{}" in "update table" clause`,
		},
		{
			sql:     "select {} from tablename where {}",
			row:     &Row{},
			errText: `unexpected query columns in exec statement`,
		},
	}

	for _, tt := range tests {
		stmt := NewUpdateRowStmt(&Row{}, tt.sql)
		db := &FakeDB{
			execErr:         tt.execErr,
			rowsAffectedErr: tt.rowsAffectedErr,
		}

		_, err := stmt.Exec(db, tt.row)
		if err == nil || err.Error() != tt.errText {
			t.Errorf("expected=%q, actual=%q", tt.errText, err)
		}
	}
}
