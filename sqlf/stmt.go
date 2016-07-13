package sqlf

import (
	"errors"
)

var errNotImplemented = errors.New("not implemented")

type InsertRowStmt struct {
	commonStmt
}

func MustPrepareInsertRow(row interface{}, sql string) *InsertRowStmt {
	stmt, err := PrepareInsertRow(row, sql)
	if err != nil {
		panic(err)
	}
	return stmt
}

func PrepareInsertRow(row interface{}, sql string) (*InsertRowStmt, error) {
	return nil, errNotImplemented
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

func MustPrepareUpdateRow(row interface{}, sql string) *UpdateRowStmt {
	stmt := &UpdateRowStmt{}
	stmt.err = errNotImplemented
	return stmt
}

func PrepareUpdateRow(row interface{}, sql string) (*UpdateRowStmt, error) {
	stmt := &UpdateRowStmt{}
	stmt.err = errNotImplemented
	return stmt, errNotImplemented
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

func MustPrepareGetRow(row interface{}, sql string) *GetRowStmt {
	stmt, err := PrepareGetRow(row, sql)
	if err != nil {
		panic(err)
	}
	return stmt
}

func PrepareGetRow(row interface{}, sql string) (*GetRowStmt, error) {
	return nil, errNotImplemented
}

func (stmt *GetRowStmt) Get(db Queryer, row interface{}) (int, error) {
	return 0, errNotImplemented
}

func (stmt *GetRowStmt) WithConfig(cfg *Config) *GetRowStmt {
	return stmt
}

type SelectStmt struct {
	commonStmt
}

func MustPrepareSelect(row interface{}, sql string) *SelectStmt {
	stmt, err := PrepareSelect(row, sql)
	if err != nil {
		panic(err)
	}
	return stmt
}

func PrepareSelect(row interface{}, sql string) (*SelectStmt, error) {
	return nil, errNotImplemented
}

func (stmt *SelectStmt) Select(db Queryer, dest interface{}, args ...interface{}) error {
	return errNotImplemented
}

func (stmt *SelectStmt) WithConfig(cfg *Config) *SelectStmt {
	return stmt
}

type commonStmt struct {
	err error
}

// String prints the SQL query associated with the statement.
func (stmt *commonStmt) String() string {
	return "not implemented"
}
