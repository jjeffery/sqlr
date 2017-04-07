package statement

import (
	"reflect"
	"sync"
)

type stmtKey struct {
	rowType reflect.Type
	sql     string
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

func getStmtFromCache(rowType reflect.Type, sql string, createFunc func(reflect.Type, string) (*Stmt, error)) (*Stmt, error) {
	var err error
	key := stmtKey{
		rowType: rowType,
		sql:     sql,
	}
	stmtCache.mu.RLock()
	stmt := stmtCache.stmts[key]
	stmtCache.mu.RUnlock()
	if stmt == nil {
		stmt, err = createFunc(rowType, sql)
		if err != nil {
			return nil, err
		}
		stmtCache.mu.Lock()
		// Check again once the exclusive lock has been acquired.
		if existing := stmtCache.stmts[key]; existing != nil {
			// another go-routine has beaten us to creating a new stmt; use theirs
			stmt = existing
		} else {
			stmtCache.stmts[key] = stmt
		}
		stmtCache.mu.Unlock()
	}
	return stmt, nil
}
