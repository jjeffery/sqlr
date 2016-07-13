package sqlf

import (
	"errors"
)

var errNotImplemented = errors.New("not implemented")

type InsertRowStmt struct {
	commonStmt
}

func PrepareInsertRow(row interface{}, sql string) *InsertRowStmt {
	stmt := &InsertRowStmt{}
	stmt.err = errNotImplemented
	return stmt
}

func (stmt *InsertRowStmt) Exec(db Execer, row interface{}) error {
	return errNotImplemented
}

func (stmt *InsertRowStmt) WithConfig(cfg *Config) *InsertRowStmt {
	return stmt
}

type UpdateRowStmt struct {
	commonStmt
}

func PrepareUpdateRow(row interface{}, sql string) *UpdateRowStmt {
	stmt := &UpdateRowStmt{}
	stmt.err = errNotImplemented
	return stmt
}

func (stmt *UpdateRowStmt) Exec(db Execer, row interface{}) (int, error) {
	return 0, errNotImplemented
}

func (stmt *UpdateRowStmt) WithConfig(cfg *Config) *UpdateRowStmt {
	return stmt
}

type GetRowStmt struct {
	commonStmt
}

func PrepareGetRow(row interface{}, sql string) *GetRowStmt {
	stmt := &GetRowStmt{}
	stmt.err = errNotImplemented
	return stmt
}

func (stmt *GetRowStmt) Get(db Queryer, row interface{}) (int, error) {
	return 0, errNotImplemented
}

func (stmt *GetRowStmt) WithConfig(cfg *Config) *GetRowStmt {
	return stmt
}

type SelectRowsStmt struct {
	commonStmt
}

func PrepareSelectRows(row interface{}, sql string) *SelectRowsStmt {
	stmt := &SelectRowsStmt{}
	stmt.err = errNotImplemented
	return stmt
}

func (stmt *SelectRowsStmt) Select(db Queryer, dest interface{}, args ...interface{}) error {
	return errNotImplemented
}

func (stmt *SelectRowsStmt) WithConfig(cfg *Config) *SelectRowsStmt {
	return stmt
}

type commonStmt struct {
	err error
}

// String prints the SQL query associated with the statement.
func (stmt *commonStmt) String() string {
	return "not implemented"
}
