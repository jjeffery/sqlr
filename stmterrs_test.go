package sqlstmt

// tests for error conditions

import (
	"database/sql"
	"errors"
	"testing"
)

type FakeExecer struct {
	execErr         error
	rowsAffected    int64
	rowsAffectedErr error
	lastInsertId    int64
	lastInsertIdErr error
}

func (fe *FakeExecer) Exec(query string, args ...interface{}) (sql.Result, error) {
	if fe.execErr != nil {
		return nil, fe.execErr
	}
	return fe, nil
}

func (fe *FakeExecer) LastInsertId() (int64, error) {
	return fe.lastInsertId, fe.lastInsertIdErr
}

func (fe *FakeExecer) RowsAffected() (int64, error) {
	return fe.rowsAffected, fe.rowsAffectedErr
}

func TestInsert_CannotSetAutoIncrement(t *testing.T) {
	type Row struct {
		ID   int64 `sql:"primary key autoincrement"`
		Name string
	}
	stmt := NewInsertRowStmt(Row{}, "tablename")
	execer := &FakeExecer{}

	err := stmt.Exec(execer, Row{})
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
	execer := &FakeExecer{
		execErr: errors.New("test error condition"),
	}

	err := stmt.Exec(execer, &Row{})
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
	execer := &FakeExecer{
		lastInsertIdErr: errors.New("test: LastInsertId"),
	}

	err := stmt.Exec(execer, &Row{})
	if err == nil || err.Error() != "test: LastInsertId" {
		t.Errorf("err=%q", err)
	}
}
