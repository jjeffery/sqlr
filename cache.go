package sqlrow

import (
	"reflect"
	"sync"
)

type statementCache struct {
	mu    sync.RWMutex
	stmts map[stmtKey]*Stmt
}

type stmtKey struct {
	rowType reflect.Type
	sql     string
}

func (c *statementCache) clear() {
	c.mu.Lock()
	c.stmts = nil
	c.mu.Unlock()
}

func (c *statementCache) lookup(rowType reflect.Type, sql string) (*Stmt, bool) {
	key := stmtKey{
		rowType: rowType,
		sql:     sql,
	}
	c.mu.RLock()
	stmt, ok := c.stmts[key]
	c.mu.RUnlock()
	return stmt, ok
}

func (c *statementCache) set(rowType reflect.Type, sql string, stmt *Stmt) *Stmt {
	key := stmtKey{
		rowType: rowType,
		sql:     sql,
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stmts == nil {
		c.stmts = make(map[stmtKey]*Stmt)
	}
	if existing, ok := c.stmts[key]; ok {
		// another goroutine beat us to adding the stmt, use its value
		stmt = existing
	} else {
		c.stmts[key] = stmt
	}
	return stmt
}
