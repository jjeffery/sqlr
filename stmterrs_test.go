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

func TestInsert_CannotSetAutoIncrement(t *testing.T) {
	type Row struct {
		ID   int64 `sql:"primary key autoincrement"`
		Name string
	}
	stmt := NewInsertRowStmt(Row{}, "tablename")
	db := &FakeDB{}

	err := stmt.Exec(db, Row{})
	if err == nil || err.Error() != "cannot set auto-increment value for type Row" {
		t.Errorf("err=%v", err)
	}
}

func TestInsert_ExecerExecFails(t *testing.T) {
	type Row struct {
		ID   int64 `sql:"primary key autoincrement"`
		Name string
	}
	stmt := NewInsertRowStmt(Row{}, "tablename")
	db := &FakeDB{
		execErr: errors.New("test error condition"),
	}

	err := stmt.Exec(db, &Row{})
	if err == nil || err.Error() != "test error condition" {
		t.Errorf("err=%q", err)
	}
}

func TestInsert_LastInsertIdFails(t *testing.T) {
	type Row struct {
		ID   int64 `sql:"primary key autoincrement"`
		Name string
	}
	stmt := NewInsertRowStmt(Row{}, "tablename")
	db := &FakeDB{
		lastInsertIdErr: errors.New("test: LastInsertId"),
	}

	err := stmt.Exec(db, &Row{})
	if err == nil || err.Error() != "test: LastInsertId" {
		t.Errorf("err=%q", err)
	}
}

func TestExec_ExecerExecFails(t *testing.T) {
	type Row struct {
		ID   int64 `sql:"primary key autoincrement"`
		Name string
	}
	stmt := NewUpdateRowStmt(Row{}, "tablename")
	db := &FakeDB{
		execErr: errors.New("test error condition"),
	}

	_, err := stmt.Exec(db, &Row{})
	if err == nil || err.Error() != "test error condition" {
		t.Errorf("err=%q", err)
	}
}

func TestExec_RowsAffectedFails(t *testing.T) {
	type Row struct {
		ID   int64 `sql:"primary key autoincrement"`
		Name string
	}
	stmt := NewUpdateRowStmt(Row{}, "tablename")
	db := &FakeDB{
		rowsAffectedErr: errors.New("test: RowsAffected"),
	}

	_, err := stmt.Exec(db, &Row{})
	if err == nil || err.Error() != "test: RowsAffected" {
		t.Errorf("err=%q", err)
	}
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
		errText string
	}{
		{
			row:     Row{},
			errText: "expected dest to be *github.com/jjeffery/sqlstmt.Row",
		},
		{
			row:     nil,
			errText: "expected dest to be *github.com/jjeffery/sqlstmt.Row",
		},
		{
			row:     (*Row)(nil),
			errText: "nil pointer passed",
		},
		{
			row:     &NotARow{},
			errText: "expected dest to be *github.com/jjeffery/sqlstmt.Row",
		},
	}

	for _, tt := range tests {
		stmt := NewGetRowStmt(Row{}, "tablename")
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
	}

	for _, tt := range tests {
		stmt := NewSelectStmt(Row{}, tt.sql)
		db := &FakeDB{queryErr: tt.queryErr}

		err := stmt.Select(db, tt.dest, tt.args...)
		if err == nil || err.Error() != tt.errText {
			t.Errorf("err=%q", err)
		}
	}
}
