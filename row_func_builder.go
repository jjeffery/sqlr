package sqlr

import (
	"fmt"
	"reflect"

	"github.com/jjeffery/errors"
	"github.com/jjeffery/sqlr/private/column"
)

// rowFuncBuilder knows how to build row-based data access functions.
type rowFuncBuilder struct {
	schema    *Schema
	rowType   reflect.Type
	tableName string
	singular  string
	plural    string
}

func (b *rowFuncBuilder) makeQuery(funcType reflect.Type) (func(*Session) reflect.Value, error) {
	for _, maker := range []func(reflect.Type) (func(*Session) reflect.Value, error){
		b.selectFunc,
		b.getOneFunc,
		b.getManyFunc,
		b.loadOneFunc,
	} {
		f, err := maker(funcType)
		if err != nil {
			return nil, err
		}
		if f != nil {
			return f, nil
		}
	}

	// TODO(jpj): need to be able to print a function type
	return nil, newError("cannot recognize function")
}

// selectFunc returns a func implementation if the func is a select func.
// input args alternatives:
//   (query string, args ...interface{})
// output args alternatives:
//   ([]*Row, error)
//   (*Row, error)
// Returns nil if not a match, returns error if the function looks like a query
// but is not quite conformant.
func (b *rowFuncBuilder) selectFunc(funcType reflect.Type) (func(*Session) reflect.Value, error) {
	if funcType.In(0) != wellKnownTypes.stringType {
		// not a select function
		return nil, nil
	}
	const invalidInputsMsg = "expect query function inputs to be like (query string, args ...interface{})"
	if funcType.NumIn() != 2 {
		return nil, newError(invalidInputsMsg)
	}
	if funcType.In(1) != wellKnownTypes.sliceOfInterfaceType {
		return nil, newError(invalidInputsMsg)
	}
	rowTypeName := b.rowTypeName()
	invalidOutputsMsg := fmt.Sprintf("MakeQuery: expect query function outputs to be like ([]*%s, error) or (*%s, error)", rowTypeName, rowTypeName)
	if funcType.NumOut() != 2 {
		return nil, newError(invalidOutputsMsg)
	}
	if funcType.Out(1) != wellKnownTypes.errorType {
		return nil, newError(invalidOutputsMsg)
	}
	rowPtrType := reflect.PtrTo(b.rowType)
	if funcType.Out(0) == rowPtrType {
		return b.makeSelectRowFunc(funcType), nil
	}
	sliceOfRowPtrType := reflect.SliceOf(rowPtrType)
	if funcType.Out(0) == sliceOfRowPtrType {
		return b.makeSelectRowsFunc(funcType), nil
	}

	return nil, newError(invalidOutputsMsg)
}

func (b *rowFuncBuilder) makeSelectRowsFunc(funcType reflect.Type) func(*Session) reflect.Value {
	return func(sess *Session) reflect.Value {
		return reflect.MakeFunc(funcType, func(args []reflect.Value) []reflect.Value {
			rowsPtrValue := reflect.New(reflect.SliceOf(reflect.PtrTo(b.rowType)))
			query := args[0].Interface().(string)
			queryArgs := args[1].Interface().([]interface{})
			_, err := sess.Select(rowsPtrValue.Interface(), query, queryArgs...)
			if err != nil {
				err = errors.Wrap(err, fmt.Sprintf("cannot query %s", b.plural)).With(
					"query", query,
					"args", queryArgs,
				)
			}
			rowsValue := rowsPtrValue.Elem()
			errValue := reflect.ValueOf(err)
			if err == nil {
				errValue = reflect.Zero(wellKnownTypes.errorType)
			}
			return []reflect.Value{
				rowsValue,
				errValue,
			}
		})
	}
}

func (b *rowFuncBuilder) makeSelectRowFunc(funcType reflect.Type) func(*Session) reflect.Value {
	return func(sess *Session) reflect.Value {
		return reflect.MakeFunc(funcType, func(args []reflect.Value) []reflect.Value {
			rowPtrValue := reflect.New(b.rowType)
			query := args[0].Interface().(string)
			queryArgs := args[1].Interface().([]interface{})
			_, err := sess.Select(rowPtrValue.Interface(), query, queryArgs...)
			if err != nil {
				err = errors.Wrap(err, fmt.Sprintf("cannot query one %s", b.singular)).With(
					"query", query,
					"args", queryArgs,
				)
				rowPtrValue = reflect.Zero(reflect.PtrTo(b.rowType))
			}

			return []reflect.Value{
				rowPtrValue,
				errorValueFor(err),
			}
		})
	}
}

func (b *rowFuncBuilder) getOneFunc(funcType reflect.Type) (func(*Session) reflect.Value, error) {
	if funcType.NumIn() != 1 {
		return nil, nil
	}
	if funcType.NumOut() != 2 {
		return nil, nil
	}
	if funcType.In(0).Kind() == reflect.Slice {
		return nil, nil
	}
	if funcType.Out(1) != wellKnownTypes.errorType {
		return nil, newError("expecting second return arg to be error")
	}
	if funcType.Out(0) == b.rowType {
		return nil, newError("expecting function to return (*%s, error) not (%s, error)", b.rowTypeName(), b.rowTypeName())
	}
	if funcType.Out(0) != reflect.PtrTo(b.rowType) {
		return nil, newError("expecting function to return (*%s, error)", b.rowTypeName())
	}

	pkCol, err := b.getPKCol()
	if err != nil {
		return nil, err
	}
	if pkCol == nil {
		return nil, newError("looks like a get func, but no primary key defined for %s", b.rowTypeName())
	}

	inType := funcType.In(0)
	if inType != pkCol.Field.Type {
		return nil, newError("looks like a get func, but %s has primary key type of %s", b.rowTypeName(), pkCol.Field.Type.Name())
	}

	return b.makeGetOneFunc(funcType), nil
}

func (b *rowFuncBuilder) makeGetOneFunc(funcType reflect.Type) func(*Session) reflect.Value {
	return func(sess *Session) reflect.Value {
		return reflect.MakeFunc(funcType, func(args []reflect.Value) []reflect.Value {
			rowPtrValue := reflect.New(b.rowType)
			query := fmt.Sprintf("select {} from %s where {}", b.tableName)
			queryArgs := []interface{}{args[0].Interface()}
			_, err := sess.Select(rowPtrValue.Interface(), query, queryArgs...)
			if err != nil {
				err = errors.Wrap(err, "cannot get one %s", b.singular).With(
					"query", query,
					"args", queryArgs,
				)
				rowPtrValue = reflect.Zero(reflect.PtrTo(b.rowType))
			}
			return []reflect.Value{
				rowPtrValue,
				errorValueFor(err),
			}
		})
	}
}

func (b *rowFuncBuilder) getManyFunc(funcType reflect.Type) (func(*Session) reflect.Value, error) {
	if funcType.NumIn() != 1 {
		return nil, nil
	}
	if funcType.NumOut() != 2 {
		return nil, nil
	}
	if funcType.In(0).Kind() != reflect.Slice {
		return nil, nil
	}
	if funcType.Out(1) != wellKnownTypes.errorType {
		return nil, newError("expecting second return arg to be error")
	}
	if funcType.Out(0) == b.rowType {
		return nil, newError("expecting function to return ([]*%s, error) not (%s, error)", b.rowTypeName(), b.rowTypeName())
	}
	if funcType.Out(0) == reflect.SliceOf(b.rowType) {
		return nil, newError("expecting function to return ([]*%s, error) not ([]%s, error)", b.rowTypeName(), b.rowTypeName())
	}
	if funcType.Out(0) != reflect.SliceOf(reflect.PtrTo(b.rowType)) {
		return nil, newError("expecting function to return ([]*%s, error)", b.rowTypeName())
	}

	pkCol, err := b.getPKCol()
	if err != nil {
		return nil, err
	}
	if pkCol == nil {
		return nil, newError("looks like a get func, but no primary key defined for %s", b.rowTypeName())
	}

	inType := funcType.In(0)
	if inType != reflect.SliceOf(pkCol.Field.Type) {
		return nil, newError("looks like a get func, but %s has primary key type of %s", b.rowTypeName(), pkCol.Field.Type.Name())
	}

	return b.makeGetManyFunc(funcType, b.schema.columnNamer().ColumnName(pkCol)), nil
}

func (b *rowFuncBuilder) makeGetManyFunc(funcType reflect.Type, pkColName string) func(*Session) reflect.Value {
	return func(sess *Session) reflect.Value {
		return reflect.MakeFunc(funcType, func(args []reflect.Value) []reflect.Value {
			var err error
			rowsPtrValue := reflect.New(reflect.SliceOf(reflect.PtrTo(b.rowType)))
			if len(args) > 0 {
				query := fmt.Sprintf("select {} from %s where `%s` in (?)", b.tableName, pkColName)
				queryArgs := make([]interface{}, len(args))
				for i, arg := range args {
					queryArgs[i] = arg.Interface()
				}
				_, err = sess.Select(rowsPtrValue.Interface(), query, queryArgs...)
				if err != nil {
					err = errors.Wrap(err, "cannot get one %s", b.singular).With(
						"query", query,
						"args", queryArgs,
					)
				}
			}
			rowsValue := rowsPtrValue.Elem()
			return []reflect.Value{
				rowsValue,
				errorValueFor(err),
			}
		})
	}
}

func (b *rowFuncBuilder) loadOneFunc(funcType reflect.Type) (func(*Session) reflect.Value, error) {
	if funcType.NumIn() != 1 {
		return nil, nil
	}
	if funcType.NumOut() != 1 {
		return nil, nil
	}
	if funcType.In(0).Kind() == reflect.Slice {
		return nil, nil
	}

	thunkType := funcType.Out(0)
	if thunkType.Kind() != reflect.Func {
		return nil, newError("looks like a load function, but does not return a thunk (function): %s", funcType.String())
	}
	if thunkType.NumIn() != 0 {
		return nil, newError("thunk function cannot accept arguments: %s", funcType.String())
	}
	invalidOutputsErr := newError("expect thunk to return (*%s error)", b.rowTypeName())
	if thunkType.NumOut() != 2 {
		return nil, invalidOutputsErr
	}
	if thunkType.Out(1) != wellKnownTypes.errorType {
		return nil, invalidOutputsErr
	}
	if thunkType.Out(0) != reflect.PtrTo(b.rowType) {
		return nil, invalidOutputsErr
	}

	pkCol, err := b.getPKCol()
	if err != nil {
		return nil, err
	}
	if pkCol == nil {
		return nil, newError("looks like a load func, but no primary key defined for %s", b.rowTypeName())
	}

	inType := funcType.In(0)
	if inType != pkCol.Field.Type {
		return nil, newError("looks like a load func, but %s has primary key type of %s", b.rowTypeName(), pkCol.Field.Type.String())
	}

	return b.makeLoadOneFunc(funcType, pkCol), nil
}

func (b *rowFuncBuilder) makeLoadOneFunc(funcType reflect.Type, pkCol *column.Info) func(*Session) reflect.Value {
	thunkType := funcType.Out(0)
	return func(sess *Session) reflect.Value {
		return reflect.MakeFunc(funcType, func(args []reflect.Value) []reflect.Value {
			thunkValue := reflect.MakeFunc(thunkType, func([]reflect.Value) []reflect.Value {
				// TODO: need to implement this
				rowPtrValue := reflect.Zero(reflect.PtrTo(b.rowType))
				err := errors.New("not implemented")
				return []reflect.Value{rowPtrValue, errorValueFor(err)}
			})

			return []reflect.Value{thunkValue}
		})
	}
}

func (b *rowFuncBuilder) getPKCol() (*column.Info, error) {
	var pkCol *column.Info
	cols := column.ListForType(b.rowType)
	for _, col := range cols {
		if col.Tag.PrimaryKey {
			if pkCol != nil {
				return nil, newError("composite primary key not supported %s", b.rowType.Name())
			}
			pkCol = col
		}
	}
	return pkCol, nil
}

func (b *rowFuncBuilder) rowTypeName() string {
	return b.rowType.String()
}

type rowFuncError string

func newError(format string, args ...interface{}) rowFuncError {
	msg := fmt.Sprintf(format, args...)
	return rowFuncError(msg)
}

func (e rowFuncError) Error() string {
	return string(e)
}
