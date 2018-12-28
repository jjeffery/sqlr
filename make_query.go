package sqlr

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/jjeffery/kv"
	"github.com/jjeffery/sqlr/dataloader"
)

func makeQuery(funcType reflect.Type, schema *Schema) (func(*Session) reflect.Value, error) {
	if f := schema.funcMap.lookup(funcType); f != nil {
		return f, nil
	}
	for _, maker := range []func(reflect.Type, *Schema) (func(*Session) reflect.Value, error){
		selectFunc,
		getOneFunc,
		getManyFunc,
		loadOneFunc,
	} {
		f, err := maker(funcType, schema)
		if err != nil {
			return nil, err
		}
		if f != nil {
			f = schema.funcMap.add(funcType, f)
			return f, nil
		}
	}

	return nil, newError("cannot recognize function: %v", funcType.String())
}

// selectFunc returns a func implementation if the func is a select func.
// input args alternatives:
//   (query string, args ...interface{})
// output args alternatives:
//   ([]*Row, error)
//   (*Row, error)
//   (int, error)
//   (int64, error)
// Returns nil if not a match, returns error if the function looks like a query
// but is not quite conformant.
func selectFunc(funcType reflect.Type, schema *Schema) (func(*Session) reflect.Value, error) {
	if funcType.NumIn() < 1 {
		return nil, nil // not a select function
	}
	if funcType.In(0) != wellKnownTypes.stringType {
		return nil, nil // not a select function
	}
	if funcType.NumIn() == 1 {
		return nil, nil
	}
	const invalidInputsMsg = "expect query function inputs to be like (query string, args ...interface{})"
	if funcType.NumIn() != 2 {
		return nil, newError(invalidInputsMsg)
	}
	if funcType.In(1) != wellKnownTypes.sliceOfInterfaceType {
		return nil, newError(invalidInputsMsg)
	}

	rowTypeName := "Row" // don't know the type yet
	invalidOutputsMsg := fmt.Sprintf("MakeQuery: expect query function outputs to be like ([]*%s, error) or (*%s, error)", rowTypeName, rowTypeName)
	if funcType.NumOut() != 2 {
		return nil, newError(invalidOutputsMsg)
	}
	if funcType.Out(1) != wellKnownTypes.errorType {
		return nil, newError(invalidOutputsMsg)
	}
	rowType := funcType.Out(0)
	switch rowType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return makeSelectIntFunc(funcType), nil
	}
	if rowType.Kind() == reflect.Slice {
		rowType = rowType.Elem()
	}
	if rowType.Kind() == reflect.Ptr {
		rowType = rowType.Elem()
	}
	if rowType.Kind() != reflect.Struct {
		return nil, newError("expected struct type, got %v", rowType.String())
	}
	tbl := schema.TableFor(rowType)
	rowPtrType := reflect.PtrTo(rowType)
	if funcType.Out(0) == rowPtrType {
		return makeSelectRowFunc(funcType, tbl), nil
	}
	sliceOfRowPtrType := reflect.SliceOf(rowPtrType)
	if funcType.Out(0) == sliceOfRowPtrType {
		return makeSelectRowsFunc(funcType, tbl), nil
	}

	return nil, newError(invalidOutputsMsg)
}

func makeSelectIntFunc(funcType reflect.Type) func(*Session) reflect.Value {
	return func(sess *Session) reflect.Value {
		return reflect.MakeFunc(funcType, func(args []reflect.Value) []reflect.Value {
			intType := funcType.Out(0)
			intPtrValue := reflect.New(intType)
			query := args[0].Interface().(string)
			queryArgs := args[1].Interface().([]interface{})
			rows, err := sess.Query(query, queryArgs...)
			if err != nil {
				err = kv.Wrap(err, "cannot query").With(
					"query", query,
					"args", queryArgs,
				)
				return []reflect.Value{
					intPtrValue.Elem(),
					errorValueFor(err),
				}
			}
			defer rows.Close()
			if !rows.Next() {
				return []reflect.Value{
					intPtrValue.Elem(),
					errorValueFor(sql.ErrNoRows),
				}
			}
			if err := rows.Scan(intPtrValue.Interface()); err != nil {
				return []reflect.Value{
					intPtrValue.Elem(),
					errorValueFor(err),
				}
			}
			if err := rows.Err(); err != nil {
				return []reflect.Value{
					intPtrValue.Elem(),
					errorValueFor(err),
				}
			}
			return []reflect.Value{
				intPtrValue.Elem(),
				wellKnownTypes.nilErrorValue,
			}
		})
	}
}

func makeSelectRowsFunc(funcType reflect.Type, tbl *Table) func(*Session) reflect.Value {
	return func(sess *Session) reflect.Value {
		return reflect.MakeFunc(funcType, func(args []reflect.Value) []reflect.Value {
			rowsPtrValue := reflect.New(reflect.SliceOf(reflect.PtrTo(tbl.RowType())))
			query := args[0].Interface().(string)
			queryArgs := args[1].Interface().([]interface{})
			_, err := sess.Select(rowsPtrValue.Interface(), query, queryArgs...)
			if err != nil {
				err = kv.Wrap(err, "cannot query rows").With(
					"rowType", tbl.RowType(),
					"query", query,
					"args", queryArgs,
				)
			}
			rowsValue := rowsPtrValue.Elem()
			errValue := errorValueFor(err)
			return []reflect.Value{
				rowsValue,
				errValue,
			}
		})
	}
}

func makeSelectRowFunc(funcType reflect.Type, tbl *Table) func(*Session) reflect.Value {
	return func(sess *Session) reflect.Value {
		return reflect.MakeFunc(funcType, func(args []reflect.Value) []reflect.Value {
			rowPtrValue := reflect.New(tbl.RowType())
			query := args[0].Interface().(string)
			queryArgs := args[1].Interface().([]interface{})
			n, err := sess.Select(rowPtrValue.Interface(), query, queryArgs...)
			if err != nil {
				err = kv.Wrap(err, "cannot query one row").With(
					"rowType", tbl.RowType(),
					"query", query,
					"args", queryArgs,
				)
				rowPtrValue = reflect.Zero(reflect.PtrTo(tbl.RowType()))
			}
			if n == 0 {
				// no rows returned
				rowPtrValue = reflect.Zero(reflect.PtrTo(tbl.RowType()))
			}

			return []reflect.Value{
				rowPtrValue,
				errorValueFor(err),
			}
		})
	}
}

func getOneFunc(funcType reflect.Type, schema *Schema) (func(*Session) reflect.Value, error) {
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
	rowType := funcType.Out(0)
	if rowType.Kind() != reflect.Ptr {
		return nil, newError("expecting first return arg to be a pointer to struct")
	}
	rowType = rowType.Elem()
	if rowType.Kind() != reflect.Struct {
		return nil, newError("expecting first return arg to be a pointer to struct")
	}
	tbl := schema.TableFor(rowType)
	if len(tbl.PrimaryKey()) == 0 {

	}

	pkCol, err := getPKCol(tbl)
	if err != nil {
		return nil, err
	}

	inType := funcType.In(0)
	if inType != pkCol.info.Field.Type {
		return nil, newError("looks like a get func, but %s has primary key type of %s", tbl.RowType().String(), pkCol.info.Field.Type.String())
	}

	return makeGetOneFunc(funcType, tbl), nil
}

func makeGetOneFunc(funcType reflect.Type, tbl *Table) func(*Session) reflect.Value {
	return func(sess *Session) reflect.Value {
		return reflect.MakeFunc(funcType, func(args []reflect.Value) []reflect.Value {
			rowPtrValue := reflect.New(tbl.RowType())
			query := fmt.Sprintf("select {} from %s where {}", tbl.Name())
			queryArgs := []interface{}{args[0].Interface()}
			n, err := sess.Select(rowPtrValue.Interface(), query, queryArgs...)
			if err != nil {
				err = kv.Wrap(err, "cannot get one row").With(
					"rowType", tbl.RowType(),
					"query", query,
					"args", queryArgs,
				)
				rowPtrValue = reflect.Zero(reflect.PtrTo(tbl.RowType()))
			} else if n == 0 {
				// nothing returned, so zero out the pointer
				rowPtrValue = reflect.Zero(reflect.PtrTo(tbl.RowType()))
			}
			return []reflect.Value{
				rowPtrValue,
				errorValueFor(err),
			}
		})
	}
}

func getManyFunc(funcType reflect.Type, schema *Schema) (func(*Session) reflect.Value, error) {
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
	rowType := funcType.Out(0)
	if rowType.Kind() != reflect.Slice {
		return nil, newError("expecting first return arg to be a slice of pointer to struct")
	}
	rowType = rowType.Elem()
	if rowType.Kind() != reflect.Ptr {
		return nil, newError("expecting first return arg to be a slice of pointer to struct")
	}
	rowType = rowType.Elem()
	if rowType.Kind() != reflect.Struct {
		return nil, newError("expecting first return arg to be a slice of pointer to struct")
	}
	tbl := schema.TableFor(rowType)
	pkCol, err := getPKCol(tbl)
	if err != nil {
		return nil, err
	}

	inType := funcType.In(0)
	if inType != reflect.SliceOf(pkCol.info.Field.Type) {
		return nil, newError("looks like a get func, but %s has primary key type of %s", tbl.RowType().String(), pkCol.info.Field.Type.String())
	}

	return makeGetManyFunc(funcType, tbl), nil
}

func makeGetManyFunc(funcType reflect.Type, tbl *Table) func(*Session) reflect.Value {
	return func(sess *Session) reflect.Value {
		return reflect.MakeFunc(funcType, func(args []reflect.Value) []reflect.Value {
			var err error
			rowsPtrValue := reflect.New(reflect.SliceOf(reflect.PtrTo(tbl.RowType())))
			idsValue := args[0]
			if idsValue.Len() > 0 {
				pkColName := tbl.PrimaryKey()[0].Name()
				query := fmt.Sprintf("select {} from %s where `%s` in (?)", tbl.Name(), pkColName)
				ids := idsValue.Interface()
				_, err = sess.Select(rowsPtrValue.Interface(), query, ids)
				if err != nil {
					err = kv.Wrap(err, "cannot get rows").With(
						"rowType", tbl.RowType(),
						"query", query,
						"args", ids,
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

func loadOneFunc(funcType reflect.Type, schema *Schema) (func(*Session) reflect.Value, error) {
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
	invalidOutputsErr := newError("expect thunk to return a pointer to struct and an error)")
	if thunkType.NumOut() != 2 {
		return nil, invalidOutputsErr
	}
	if thunkType.Out(1) != wellKnownTypes.errorType {
		return nil, invalidOutputsErr
	}
	rowType := thunkType.Out(0)
	if rowType.Kind() != reflect.Ptr {
		return nil, invalidOutputsErr
	}
	rowType = rowType.Elem()
	if rowType.Kind() != reflect.Struct {
		return nil, invalidOutputsErr
	}
	tbl := schema.TableFor(rowType)

	pkCol, err := getPKCol(tbl)
	if err != nil {
		return nil, err
	}

	inType := funcType.In(0)
	if inType != pkCol.info.Field.Type {
		return nil, newError("looks like a load func, but %s has primary key type of %s", tbl.RowType().String(), pkCol.info.Field.Type.String())
	}

	return makeLoadOneFunc(funcType, tbl), nil
}

func makeLoadOneFunc(funcType reflect.Type, tbl *Table) func(*Session) reflect.Value {
	return func(sess *Session) reflect.Value {
		pkCol := tbl.PrimaryKey()[0]
		queryFuncIn := []reflect.Type{reflect.SliceOf(pkCol.fieldType())}
		queryFuncOut := []reflect.Type{reflect.SliceOf(reflect.PtrTo(tbl.RowType())), wellKnownTypes.errorType}
		queryFuncType := reflect.FuncOf(queryFuncIn, queryFuncOut, false)
		queryFuncValue := makeGetManyFunc(queryFuncType, tbl)(sess)

		keyFuncIn := []reflect.Type{reflect.PtrTo(tbl.RowType())}
		keyFuncOut := []reflect.Type{pkCol.fieldType()}
		keyFuncType := reflect.FuncOf(keyFuncIn, keyFuncOut, false)
		keyFuncValue := reflect.MakeFunc(keyFuncType, func(args []reflect.Value) []reflect.Value {
			rowPtrValue := args[0]
			rowValue := rowPtrValue.Elem()
			keyValue := rowValue.FieldByIndex([]int(pkCol.fieldIndex()))
			return []reflect.Value{keyValue}
		})

		loadFuncPtrValue := reflect.New(funcType)
		dataloader.Make(loadFuncPtrValue.Interface(), queryFuncValue.Interface(), keyFuncValue.Interface())
		return loadFuncPtrValue.Elem()
	}
}

func getPKCol(tbl *Table) (*Column, error) {
	pkCols := tbl.PrimaryKey()
	if len(pkCols) > 1 {
		return nil, newError("compositeprimary key not supported: %s", tbl.RowType().String())
	}
	if len(pkCols) == 0 {
		return nil, newError("no primary key defined for %s (table %s)", tbl.RowType().String(), tbl.Name())
	}
	return pkCols[0], nil
}

type rowFuncError string

func newError(format string, args ...interface{}) rowFuncError {
	msg := fmt.Sprintf(format, args...)
	return rowFuncError(msg)
}

func (e rowFuncError) Error() string {
	return string(e)
}
