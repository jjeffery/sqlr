package column

import (
	"reflect"
	"sync"
)

var typeMap = struct {
	mu sync.RWMutex
	m  map[reflect.Type][]*Info
}{
	m: make(map[reflect.Type][]*Info),
}

// ListForType returns a list of column information
// associated with the specified type, which must be a struct.
func ListForType(rowType reflect.Type) []*Info {
	typeMap.mu.RLock()
	list, ok := typeMap.m[rowType]
	typeMap.mu.RUnlock()
	if ok {
		return list
	}

	typeMap.mu.Lock()
	defer typeMap.mu.Unlock()
	list = newList(rowType)
	typeMap.m[rowType] = list
	return list
}
