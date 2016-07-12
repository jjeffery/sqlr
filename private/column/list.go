package column

import (
	"database/sql"
	"reflect"
	"time"
)

// Standard types.
var (
	sqlScanType = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
	timeType    = reflect.TypeOf(time.Time{})
)

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
	columnName := columnNameFromTag(field.Tag)
	if columnName == "-" {
		// ignore field marked as not a column
		return
	}

	if len(field.PkgPath) != 0 && !field.Anonymous {
		// ignore unexported field
		return
	}

	fieldType := field.Type
	for fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	// ignore fields that are arrays, channels, functions, interfaces, maps, slices
	switch fieldType.Kind() {
	case reflect.Array, reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Slice:
		return
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
	state.path = state.path.Append(field.Name, columnName)

	// An embedded structure will not be mapped recursively if it meets
	// any of the following criteria:
	// * it is time.Time (special case)
	// * it implements sql.Scan (unlikely)
	// * its pointer type implements sql.Scan (more likely)
	if fieldType.Kind() == reflect.Struct &&
		fieldType != timeType &&
		!fieldType.Implements(sqlScanType) &&
		!reflect.PtrTo(fieldType).Implements(sqlScanType) {
		list.addFields(fieldType, state)
		return
	}

	info := &Info{
		Field: field,
		Index: state.index,
		Path:  state.path,
	}

	info.update()

	*list = append(*list, info)
}
