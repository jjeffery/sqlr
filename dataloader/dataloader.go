package dataloader

import (
	"fmt"
	"reflect"
)

// maxKeysPerQuery is tne maximum number of keys to include in any query
// to the query function. This is a variable so that the tests can modify it.
var maxKeysPerQuery = 100

type dataLoader struct {
	keyType         reflect.Type // type of ID, should be a string or integer type
	thunkResultType reflect.Type // type of result returned by the thunk, often same as rowType
	thunkElemType   reflect.Type // if thunkResultType is a slice, this is the type of each slice element
	thunkType       reflect.Type // type of the thunk, should be a func returning thunkResultType and an error
	queryRowType    reflect.Type // type of the row returned by the query func, usually a pointer to struct

	queryFuncValue reflect.Value // value of the function that performs the query
	keyFuncValue   reflect.Value // value of the function that extracts the key

	thunks  map[interface{}]*thunkT // map of all thunks
	pending map[interface{}]*thunkT // map of thunks that have not been called yet
}

// Call implements the loader function, whose responsibility is to locate
// the thunk for the specified key. If the corresponding thunk does not
// already exist it is created.
func (loader *dataLoader) Call(args []reflect.Value) (results []reflect.Value) {
	keyValue := args[0]
	key := keyValue.Interface()
	thunk := loader.thunks[key]
	if thunk == nil {
		thunk = &thunkT{
			loader:   loader,
			keyValue: keyValue,
			key:      key,
			pending:  true,
		}
		thunk.funcValue = []reflect.Value{reflect.MakeFunc(loader.thunkType, thunk.Call)}
		loader.thunks[key] = thunk
		loader.pending[key] = thunk
	}
	return thunk.funcValue
}

// performQuery is called by the thunk's call function. It builds a list of
// pending thunks and invokes the performQuery function with the keys associated
// with those pending thunks. It ensures that the key assocated with the thunk
// that has just been called is included in the query function arguments.
func (loader *dataLoader) performQuery(calledThunk *thunkT) {
	cap := maxKeysPerQuery
	if len(loader.pending) < cap {
		cap = len(loader.pending)
	}
	keyValues := reflect.MakeSlice(reflect.SliceOf(loader.keyType), 0, cap)
	thunks := make(map[interface{}]*thunkT)
	zeroResultValue := []reflect.Value{reflect.Zero(loader.thunkResultType), knownTypes.nilErrorValue}
	addThunk := func(thunk *thunkT) {
		if _, ok := thunks[thunk.key]; !ok {
			keyValues = reflect.Append(keyValues, thunk.keyValue)
			thunks[thunk.key] = thunk
			thunk.pending = false
			thunk.resultValue = zeroResultValue
			cap--
		}
	}

	// make sure the key for the called thunk is included in the key list
	addThunk(calledThunk)

	for _, thunk := range loader.pending {
		if cap == 0 {
			// reached the maximum keys that can be included in one query
			break
		}
		addThunk(thunk)
	}

	// remove all the keys from the pending map
	for key := range thunks {
		delete(loader.pending, key)
	}

	queryResult := loader.queryFuncValue.Call([]reflect.Value{keyValues})
	errorValue := queryResult[1]
	if !errorValue.IsNil() {
		// When there is an error, all the thunks receive the same value, which
		// is the zero value for the val type, and the common error value.
		errorResultValue := []reflect.Value{reflect.Zero(calledThunk.loader.thunkResultType), errorValue}
		for _, thunk := range thunks {
			thunk.resultValue = errorResultValue
		}
		return
	}

	// sliceValue is the value of the first output from the query function.
	// It must be a slice of query row objects.
	sliceValue := queryResult[0]

	if loader.thunkResultType == loader.thunkElemType {
		if loader.thunkElemType == loader.queryRowType {
			// keyFunc returns one value
			for i := 0; i < sliceValue.Len(); i++ {
				rowValue := sliceValue.Index(i)
				keyFuncResult := loader.keyFuncValue.Call([]reflect.Value{rowValue})
				keyValue := keyFuncResult[0]
				key := keyValue.Interface()
				if thunk, ok := thunks[key]; ok {
					thunk.resultValue = []reflect.Value{rowValue, knownTypes.nilErrorValue}
				}
			}
		} else {
			// keyFunc must return two values
			for i := 0; i < sliceValue.Len(); i++ {
				rowValue := sliceValue.Index(i)
				keyFuncResult := loader.keyFuncValue.Call([]reflect.Value{rowValue})
				keyValue := keyFuncResult[0]
				scalarValue := keyFuncResult[1]
				key := keyValue.Interface()
				if thunk, ok := thunks[key]; ok {
					thunk.resultValue = []reflect.Value{scalarValue, knownTypes.nilErrorValue}
				}
			}
		}
	} else {
		if loader.thunkElemType == loader.queryRowType {
			// keyFunc returns one value
			resultValueMap := make(map[interface{}]reflect.Value)
			for i := 0; i < sliceValue.Len(); i++ {
				rowValue := sliceValue.Index(i)
				keyFuncResult := loader.keyFuncValue.Call([]reflect.Value{rowValue})
				keyValue := keyFuncResult[0]
				key := keyValue.Interface()
				resultSliceValue, ok := resultValueMap[key]
				if !ok {
					resultSliceValue = reflect.MakeSlice(reflect.SliceOf(loader.queryRowType), 0, 4)
				}
				resultSliceValue = reflect.Append(resultSliceValue, rowValue)
				resultValueMap[key] = resultSliceValue
			}
			for key, resultValue := range resultValueMap {
				if thunk, ok := thunks[key]; ok {
					thunk.resultValue = []reflect.Value{resultValue, knownTypes.nilErrorValue}
				}
			}
		} else {
			// keyFunc must return two values
			panic("not implemented")
		}
	}
}

type thunkT struct {
	loader      *dataLoader
	funcValue   []reflect.Value
	keyValue    reflect.Value
	rowValue    reflect.Value
	resultValue []reflect.Value
	key         interface{}
	row         interface{}
	pending     bool
}

func (thunk *thunkT) Call(args []reflect.Value) (result []reflect.Value) {
	// should not need this check
	if len(args) != 0 {
		panic(fmt.Sprintf("thunk function called with arguments: %v", args))
	}
	if thunk.pending {
		thunk.loader.performQuery(thunk)
		if thunk.pending {
			panic("thunk should no longer be pending")
		}
	}
	return thunk.resultValue
}

var knownTypes struct {
	errorType     reflect.Type
	nilErrorValue reflect.Value
}

func init() {
	knownTypes.errorType = reflect.TypeOf((*error)(nil)).Elem()
	knownTypes.nilErrorValue = reflect.Zero(knownTypes.errorType)
}

// Make a data loader function given a query function and a key function.
//
// The loaderFuncPtr arg must be a pointer to a function variable which will
// receive the created loader function.
// The queryFunc and keyFunc args are the query function and the key function,
// as described in the package description.
// If any of the arguments are not supplied correctly, this function will
// panic.
//
// See the package description and the examples for more detail.
func Make(loaderFuncPtr interface{}, queryFunc interface{}, keyFunc interface{}) {
	loader := dataLoader{
		thunks:  make(map[interface{}]*thunkT),
		pending: make(map[interface{}]*thunkT),
	}
	processLoaderFuncPtr(&loader, loaderFuncPtr)
	processQueryFunc(&loader, queryFunc)
	processKeyFunc(&loader, keyFunc)
	loaderFuncValue := reflect.MakeFunc(reflect.TypeOf(loaderFuncPtr).Elem(), loader.Call)
	loaderFuncPtrValue := reflect.ValueOf(loaderFuncPtr)
	loaderFuncPtrValue.Elem().Set(loaderFuncValue)
}

func processKeyFunc(loader *dataLoader, keyFunc interface{}) {
	loader.keyFuncValue = reflect.ValueOf(keyFunc)
	keyFuncType := loader.keyFuncValue.Type()
	if keyFuncType.Kind() != reflect.Func {
		panic("keyFunc must be a function")
	}
	if loader.keyFuncValue.IsNil() {
		panic("keyFunc is nil")
	}
	if keyFuncType.NumIn() != 1 {
		panic(fmt.Sprintf("keyFunc must have one input parameter of type %v", loader.queryRowType))
	}
	if inType := keyFuncType.In(0); inType != loader.queryRowType {
		panic(fmt.Sprintf("keyFunc must have one input parameter of type %v", loader.queryRowType))
	}
	if loader.queryRowType == loader.thunkElemType {
		if numOut := keyFuncType.NumOut(); numOut != 1 {
			panic(fmt.Sprintf("keyFunc must return one value of type %v", loader.keyType))
		}
		if keyType := keyFuncType.Out(0); keyType != loader.keyType {
			panic(fmt.Sprintf("keyFunc must return one value of type %v", loader.keyType))
		}
	} else {
		if numOut := keyFuncType.NumOut(); numOut != 2 {
			panic(fmt.Sprintf("keyFunc must return two values: (%v, %v)", loader.keyType, loader.thunkElemType))
		}
		if keyType := keyFuncType.Out(0); keyType != loader.keyType {
			panic(fmt.Sprintf("keyFunc must return two values: (%v, %v)", loader.keyType, loader.thunkElemType))
		}
		if thunkScalarType := keyFuncType.Out(1); thunkScalarType != loader.thunkElemType {
			panic(fmt.Sprintf("keyFunc must return two values: (%v, %v)", loader.keyType, loader.thunkElemType))
		}
	}
}

func processQueryFunc(loader *dataLoader, queryFunc interface{}) {
	loader.queryFuncValue = reflect.ValueOf(queryFunc)
	queryFuncType := loader.queryFuncValue.Type()
	if queryFuncType.Kind() != reflect.Func {
		panic("queryFunc must be a function")
	}
	if loader.queryFuncValue.IsNil() {
		panic("queryFunc is nil")
	}
	if queryFuncType.NumIn() != 1 {
		panic(fmt.Sprintf("queryFunc should accept one parameter, of type []%v", loader.keyType))
	}
	inSliceType := queryFuncType.In(0)
	if inSliceType.Kind() != reflect.Slice {
		panic(fmt.Sprintf("queryFunc should accept one parameter, of type []%v", loader.keyType))
	}
	keyType := inSliceType.Elem()
	if keyType != loader.keyType {
		panic(fmt.Sprintf("queryFunc should accept one parameter, of type []%v", loader.keyType))
	}
	if queryFuncType.NumOut() != 2 {
		panic("queryFunc should return two values, a slice and an error")
	}
	if queryFuncType.Out(1) != knownTypes.errorType {
		panic("queryFunc should return two values, the second of which is an error")
	}
	outSliceType := queryFuncType.Out(0)
	if outSliceType.Kind() != reflect.Slice {
		panic("queryFunc should two values, the first of which is a slice")
	}
	loader.queryRowType = outSliceType.Elem()
}

func processLoaderFuncPtr(loader *dataLoader, loaderFuncPtr interface{}) {
	// figure out the types
	loaderFuncPtrValue := reflect.ValueOf(loaderFuncPtr)
	loaderFuncPtrType := loaderFuncPtrValue.Type()
	if loaderFuncPtrType.Kind() != reflect.Ptr {
		panic("loaderFuncPtr must be a pointer to a function")
	}
	if loaderFuncPtrValue.IsNil() {
		panic("loaderFuncPtr is nil")
	}
	loaderFuncType := loaderFuncPtrType.Elem()
	if loaderFuncType.Kind() != reflect.Func {
		panic("loaderFuncPtr must be a pointer to a function")
	}
	if loaderFuncType.NumIn() != 1 {
		panic("loaderFuncPtr should have one input parameter, of string or integral type")
	}
	loader.keyType = loaderFuncType.In(0)
	if !isKeyType(loader.keyType) {
		panic("loaderFuncPtr should accept one input parameter, of string or integral type")
	}
	if loaderFuncType.NumOut() != 1 {
		panic("loaderFuncPtr should have one return value: a function thunk")
	}
	loader.thunkType = loaderFuncType.Out(0)
	if loader.thunkType.Kind() != reflect.Func {
		panic("loaderFuncPtr should have one return value: a function thunk")
	}
	if loader.thunkType.NumIn() != 0 {
		panic("loaderFuncPtr should return a function thunk that has no input parameters")
	}
	if loader.thunkType.NumOut() != 2 {
		panic("loaderFuncPtr should return a function thunk that returns two values")
	}
	if loader.thunkType.Out(1) != knownTypes.errorType {
		panic("loaderFuncPtr should return a function thunk whose second return value should be an error")
	}
	loader.thunkResultType = loader.thunkType.Out(0)
	if loader.thunkResultType.Kind() == reflect.Slice {
		loader.thunkElemType = loader.thunkResultType.Elem()
	} else {
		loader.thunkElemType = loader.thunkResultType
	}
}

func isKeyType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Array:
		// arrays are acceptable key types, for example [16]byte
		return isKeyType(t.Elem())
	case reflect.String, reflect.Int:
		return true
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return true
	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return true
	}
	return false
}

type rowCacheKey struct {
	idType  reflect.Type
	rowType reflect.Type
}

type rowCache struct {
	cache map[rowCacheKey]map[interface{}]interface{}
}
