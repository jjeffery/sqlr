package sqlstmt

// tests for error conditions

import (
	"database/sql"
	"testing"
)

type FakeExecer struct{}

func (fe *FakeExecer) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func TestCannotSetAutoIncrement(t *testing.T) {
	type Row struct {
		ID   int64 `sql:"primary key autoincrement"`
		Name string
	}
	stmt := NewInsertRowStmt(Row{}, "tablename")
	execer := &FakeExecer{}

	err := stmt.Exec(execer, Row{})
	if err == nil || err.Error() != "cannot set auto-increment value for type Row" {
		t.Errorf("err=%q", err)
	}
}
