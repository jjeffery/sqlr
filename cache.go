package sqlrow

import (
	"fmt"
	"reflect"
	"sync"
)

type stmtKey struct {
	dialectName    string
	conventionName string
	rowType        reflect.Type
	sql            string
}

func (k stmtKey) String() string {
	return fmt.Sprintf("[%s,%s,%v,%v]", k.dialectName, k.conventionName, k.rowType, k.sql)
}

var stmtCache = struct {
	mu    sync.RWMutex
	stmts map[stmtKey]*Stmt
}{
	stmts: make(map[stmtKey]*Stmt),
}

// clearStmtCache is only used during testing.
func clearStmtCache() {
	stmtCache.mu.Lock()
	stmtCache.stmts = make(map[stmtKey]*Stmt)
	stmtCache.mu.Unlock()
}

func getStmtFromCache(dialect Dialect, convention Convention, rowType reflect.Type, sql string) (*Stmt, error) {
	var err error
	key := stmtKey{
		dialectName:    dialect.Name(),
		conventionName: convention.Key(),
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
