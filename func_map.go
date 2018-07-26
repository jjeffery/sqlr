package sqlr

import (
	"reflect"
	"sync"
)

// funcMap is used to lookup functions based type info.
// It is safe for concurrent access because funcs can
// be added during program execution.
//
// A sync.Map is used because this use case is one that it was designed
// for, namely a cache that is only ever added to.
type funcMap struct {
	funcs sync.Map
}

// add a func to the map and return the value for the func
// in the map. The value returned will be different to f if
// another goroutine has already added an entry to the map for
// the func.
func (fm *funcMap) add(funcType reflect.Type, f func(*Session) reflect.Value) func(*Session) reflect.Value {
	v, _ := fm.funcs.LoadOrStore(funcType, f)
	return v.(func(*Session) reflect.Value)
}

// lookup a func based on its row type in the map. Returns nil
// if not found.
func (fm *funcMap) lookup(funcType reflect.Type) func(*Session) reflect.Value {
	if v, ok := fm.funcs.Load(funcType); ok {
		return v.(func(*Session) reflect.Value)
	}
	return nil
}
