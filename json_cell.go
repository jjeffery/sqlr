package sqlrow

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// jsonCell is used to unmarshal JSON cells into their destination type
type jsonCell struct {
	colname   string
	cellValue interface{}
	data      []byte
}

func newJSONCell(colname string, v interface{}) *jsonCell {
	return &jsonCell{
		colname:   colname,
		cellValue: v,
	}
}

func (jc *jsonCell) ScanValue() interface{} {
	return &jc.data
}

func (jc *jsonCell) Unmarshal() error {
	if len(jc.data) == 0 {
		// No JSON data to unmarshal, so set to the zero value
		// for this type. We know that jc.cellValue is a pointer,
		// so it is safe to call Elem() and set the value.
		valptr := reflect.ValueOf(jc.cellValue)
		val := valptr.Elem()
		val.Set(reflect.Zero(val.Type()))
		return nil
	}
	if err := json.Unmarshal(jc.data, jc.cellValue); err != nil {
		// TODO(jpj): if Wrap makes it into the stdlib, use it here
		return fmt.Errorf("cannot unmarshal JSON field %q: %v", jc.colname, err)
	}
	return nil
}
