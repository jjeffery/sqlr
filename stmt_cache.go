package sqlr

import (
	"reflect"
	"sync"
)

// stmtCache is a cache of statements for a schema.
type stmtCache struct {
	mu    sync.RWMutex
	stmts map[stmtKey]*Stmt
}

// stmtKey is the unique key used to identify statements within
// a statement cache. Note that two statements with the same key might
// be different for different schemas, as  the dialects and/or naming conventions
// could be different.
type stmtKey struct {
	rowType reflect.Type
	query   string
}

func (c *stmtCache) clear() {
	c.mu.Lock()
	c.stmts = nil
	c.mu.Unlock()
}

func (c *stmtCache) lookup(rowType reflect.Type, query string) (*Stmt, bool) {
	key := stmtKey{
		rowType: rowType,
		query:   query,
	}
	c.mu.RLock()
	stmt, ok := c.stmts[key]
	c.mu.RUnlock()
	return stmt, ok
}

// set the statement for the given rowType and query string. Returns the statement,
// which could be different from the input statement if another goroutine has already
// set a statement for the same row type and query.
func (c *stmtCache) set(rowType reflect.Type, query string, stmt *Stmt) *Stmt {
	key := stmtKey{
		rowType: rowType,
		query:   query,
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
