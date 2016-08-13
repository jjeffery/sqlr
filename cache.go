package sqlrow

import (
	"reflect"
	"sync"
)

type stmtKey struct {
	dialectName    string
	conventionName string
	rowType        reflect.Type
	sql            string
}

var stmtCache = struct {
	mu    sync.RWMutex
	stmts map[stmtKey]*Stmt
}{
	stmts: make(map[stmtKey]*Stmt),
}

func getStmtFromCache(dialect Dialect, convention Convention, rowType reflect.Type, sql string) (*Stmt, error) {
	var err error
	key := stmtKey{
		dialectName:    dialect.Name(),
		conventionName: convention.Name(),
		rowType:        rowType,
		sql:            sql,
	}
	stmtCache.mu.RLock()
	stmt := stmtCache.stmts[key]
	stmtCache.mu.RUnlock()
	if stmt == nil {
		stmt, err = newStmt(dialect, convention, rowType, sql)
		if err != nil {
			return nil, err
		}
		stmtCache.mu.Lock()
		// This could overwrite an existing stmt if two goroutines are
		// creating the same statement at the same time. No harm done as
		// the two stmts created should be identical.
		stmtCache.stmts[key] = stmt
		stmtCache.mu.Unlock()
	}
	return stmt, nil
}
