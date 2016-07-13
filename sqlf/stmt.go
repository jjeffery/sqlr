package sqlf

import (
	"errors"
)

var errNotImplemented = errors.New("not implemented")

type InsertRowStmt struct {
	// todo
	err error
}

func PrepareInsertRow(row interface{}, sql string) *InsertRowStmt {
	return &InsertRowStmt{
		err: errNotImplemented,
	}
}

func (stmt *InsertRowStmt) Exec(db Execer, row interface{}) error {
	return errNotImplemented
}

func (stmt *InsertRowStmt) WithConfig(cfg *Config) *InsertRowStmt {
	return stmt
}

type UpdateRowStmt struct {
	// todo
	err error
}

func PrepareUpdateRow(row interface{}, sql string) *UpdateRowStmt {
	return &UpdateRowStmt{
		err: errNotImplemented,
	}
}

func (stmt *UpdateRowStmt) Exec(db Execer, row interface{}) (int, error) {
	return 0, errNotImplemented
}

func (stmt *UpdateRowStmt) WithConfig(cfg *Config) *UpdateRowStmt {
	return stmt
}

type GetRowStmt struct {
	// todo
	err error
}

func PrepareGetRow(row interface{}, sql string) *GetRowStmt {
	return &GetRowStmt{
		err: errNotImplemented,
	}
}

func (stmt *GetRowStmt) Get(db Queryer, row interface{}) (int, error) {
	return 0, errNotImplemented
}

func (stmt *GetRowStmt) WithConfig(cfg *Config) *GetRowStmt {
	return stmt
}

type SelectRowsStmt struct {
	// todo
	err error
}

func PrepareSelectRows(row interface{}, sql string) *SelectRowsStmt {
	return &SelectRowsStmt{
		err: errNotImplemented,
	}
}

func (stmt *SelectRowsStmt) Select(db Queryer, dest interface{}, args ...interface{}) error {
	return errNotImplemented
}

func (stmt *SelectRowsStmt) WithConfig(cfg *Config) *SelectRowsStmt {
	return stmt
}
