package column

import (
	"database/sql"
	"reflect"
	"strings"
	"sync"
	"time"
)

// Standard types.
var (
	sqlScanType = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
	timeType    = reflect.TypeOf(time.Time{})
)

// typeMap contains a map of type to column information used
// to cache results for ListForType.
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

// newList returns a list of column information for the row type.
func newList(rowType reflect.Type) []*Info {
	var list columnList
	var state = stateT{}
	list.addFields(rowType, state)
	return list
}

type stateT struct {
	index Index
	path  Path
}

type columnList []*Info

func (list *columnList) addFields(rowType reflect.Type, state stateT) {
	for i := 0; i < rowType.NumField(); i++ {
		field := rowType.Field(i)
		list.addField(field, i, state)
	}
}

func (list *columnList) addField(field reflect.StructField, i int, state stateT) {
	// Search for a key in the struct tag that starts with "-", which indicates
	// that the field should not be mapped to a column. Eg `sql:"-"`
	for _, key := range structTagKeys {
		if value := field.Tag.Get(key); strings.HasPrefix(value, "-") {
			// ignore field marked as not a column
			return
		}
	}

	if len(field.PkgPath) != 0 && !field.Anonymous {
		// ignore unexported field
		return
	}

	fieldType := field.Type
	for fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	// ignore fields that are channels or functions, as these cannot be
	// marshalled as JSON
	switch fieldType.Kind() {
	case reflect.Chan, reflect.Func:
		return
	}

	// Construct the info and parse the tag. This is done now because
	// it is necessary to know if the field will be serialized as JSON
	// in order to decide whether to include the field or not.
	info := newInfo(field)

	// Ignore certain types unless they are marked as JSON serialized.
	if !info.Tag.JSON {
		// ignore fields that are arrays, interfaces, maps
		switch fieldType.Kind() {
		case reflect.Array, reflect.Interface, reflect.Map:
			return
		}

		// ignore slices that are not byte slices
		if fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() != reflect.Uint8 {
			return
		}
	}

	// update the state's field index to point to this field
	state.index = state.index.Append(i)

	if fieldType.Kind() == reflect.Struct && field.Anonymous {
		// Any anonymous structure is automatically added.
		list.addFields(fieldType, state)
		return
	}

	// The field is not anonymous, and is not ignored, so it
	// is either a field assocated with a column, or a struct
	// with embedded fields.
	state.path = state.path.Append(field.Name, field.Tag)

	// An embedded structure will not be mapped recursively if it meets
	// any of the following criteria:
	// * it is time.Time (special case)
	// * it implements sql.Scan (unlikely)
	// * its pointer type implements sql.Scan (more likely)
	// * it is marked as serialize to JSON
	if fieldType.Kind() == reflect.Struct &&
		fieldType != timeType &&
		!fieldType.Implements(sqlScanType) &&
		!reflect.PtrTo(fieldType).Implements(sqlScanType) &&
		!info.Tag.JSON {
		list.addFields(fieldType, state)
		return
	}

	info.Index = state.index
	info.Path = state.path
	info.FieldNames = info.Path.String()

	*list = append(*list, info)
}
