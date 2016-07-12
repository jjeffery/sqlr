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

type SelectStmt struct {
	// todo
	err error
}

func PrepareSelect(row interface{}, sql string) *SelectStmt {
	return &SelectStmt{
		err: errNotImplemented,
	}
}

func (stmt *SelectStmt) Select(db Queryer, dest interface{}, args ...interface{}) error {
	return errNotImplemented
}

func (stmt *SelectStmt) WithConfig(cfg *Config) *SelectStmt {
	return stmt
}
