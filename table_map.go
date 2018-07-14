package sqlr

import (
	"reflect"
	"sync"
)

// tableMap is used to lookup table info based on row type info.
// It is safe for concurrent acceess because table/row type info can
// be added during program execution.
//
// A sync.Map is used because this use case is one that it was designed
// for, namely a cache that is only ever added to.
type tableMap struct {
	tables sync.Map
}

// add a table to the map and return the value for the table
// in the map. The value returned will be different to tbl if
// another goroutine has already added an entry to the map for
// the table.
func (tm *tableMap) add(rowType reflect.Type, tbl *Table) *Table {
	v, _ := tm.tables.LoadOrStore(rowType, tbl)
	return v.(*Table)
}

// lookup a table based on its row type in the map. Returns nil
// if not found.
func (tm *tableMap) lookup(rowType reflect.Type) *Table {
	if v, ok := tm.tables.Load(rowType); ok {
		return v.(*Table)
	}
	return nil
}
