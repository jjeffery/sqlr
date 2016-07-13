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
	return &InsertRowStmt{}, nil
}

func (stmt *InsertRowStmt) Exec(db Execer, row interface{}) error {
	return errNotImplemented
}

type UpdateRowStmt struct {
	commonStmt
}

func MustPrepareUpdateRow(row interface{}, sql string) *UpdateRowStmt {
	stmt, err := PrepareUpdateRow(row, sql)
	if err != nil {
		panic(err)
	}
	return stmt
}

func PrepareUpdateRow(row interface{}, sql string) (*UpdateRowStmt, error) {
	return &UpdateRowStmt{}, nil
}

func (stmt *UpdateRowStmt) Exec(db Execer, row interface{}) (int, error) {
	return 0, errNotImplemented
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
	return &GetRowStmt{}, nil
}

func (stmt *GetRowStmt) Get(db Queryer, row interface{}) (int, error) {
	return 0, errNotImplemented
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
	return &SelectStmt{}, nil
}

func (stmt *SelectStmt) Select(db Queryer, dest interface{}, args ...interface{}) error {
	return errNotImplemented
}

type commonStmt struct {
	err error
}

// String prints the SQL query associated with the statement.
func (stmt *commonStmt) String() string {
	return "not implemented"
}
