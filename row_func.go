package sqlr

import (
	"reflect"
)

var wellKnownTypes = struct {
	errorType            reflect.Type
	stringType           reflect.Type
	sliceOfInterfaceType reflect.Type
	nilErrorValue        reflect.Value
}{
	errorType:            reflect.TypeOf((*error)(nil)).Elem(),
	stringType:           reflect.TypeOf((*string)(nil)).Elem(),
	sliceOfInterfaceType: reflect.SliceOf(reflect.TypeOf((*interface{})(nil)).Elem()),
}

func init() {
	wellKnownTypes.nilErrorValue = reflect.Zero(wellKnownTypes.errorType)
}

func errorValueFor(err error) reflect.Value {
	if err == nil {
		return wellKnownTypes.nilErrorValue
	}
	errValue := reflect.New(wellKnownTypes.errorType).Elem()
	errValue.Set(reflect.ValueOf(err))
	return errValue
}
