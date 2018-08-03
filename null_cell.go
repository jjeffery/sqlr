package sqlr

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"
)

var (
	timeType = reflect.TypeOf(time.Time{})
	timeZero = reflect.Zero(reflect.TypeOf(time.Time{}))
)

// newNullCell returns a scannable value for fields that are configured
// such that a SQL NULL value means to store an empty value for the type.
// These fields should have a backing field type of int, uint, bool, float, string or time.Time.
func newNullCell(colname string, cellValue reflect.Value, cellPtr interface{}) interface{} {
	if scanner, ok := cellPtr.(sql.Scanner); ok {
		return &nullScannerCell{colname: colname, cellValue: cellValue, scanner: scanner}
	}
	switch cellValue.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return &nullIntCell{colname: colname, cellValue: cellValue}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return &nullUintCell{colname: colname, cellValue: cellValue}
	case reflect.Float32, reflect.Float64:
		return &nullFloatCell{colname: colname, cellValue: cellValue}
	case reflect.Bool:
		return &nullBoolCell{colname: colname, cellValue: cellValue}
	case reflect.String:
		return &nullStringCell{colname: colname, cellValue: cellValue}
	case reflect.Struct:
		if cellValue.Type() == timeType {
			return &nullTimeCell{colname: colname, cellValue: cellValue}
		}
		return cellPtr
	default:
		// other valid types include pointer and slice, which
		// can handle a null value without resorting to reflection
		return cellPtr
	}
}

type nullScannerCell struct {
	colname   string
	cellValue reflect.Value
	scanner   sql.Scanner
}

func (nc *nullScannerCell) Scan(v interface{}) (err error) {
	defer func() {
		// handle panic if Set fails
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot scan column %q: %v", nc.colname, r)
		}
	}()

	// attempt to scan, because the Scan implementation may handle
	// nil values correctly
	err = nc.scanner.Scan(v)
	if err != nil && v == nil {
		// scan failed for nil value, so set the zero value
		nc.cellValue.Set(reflect.Zero(nc.cellValue.Type()))
		err = nil
	}
	return err
}

type nullIntCell struct {
	colname   string
	cellValue reflect.Value
	bits      int
}

func (nc *nullIntCell) Scan(v interface{}) (err error) {
	defer func() {
		// handle panic if SetFloat overflows
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot scan column %q: %v", nc.colname, r)
		}
	}()
	var nullable sql.NullInt64
	if err = nullable.Scan(v); err != nil {
		return fmt.Errorf("cannot scan column %q: %v", nc.colname, err)
	}
	if nullable.Valid {
		nc.cellValue.SetInt(nullable.Int64)
	} else {
		nc.cellValue.SetInt(0)
	}
	return nil
}

type nullUintCell struct {
	colname   string
	cellValue reflect.Value
}

func (nc *nullUintCell) Scan(v interface{}) (err error) {
	defer func() {
		// handle panic if SetFloat overflows
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot scan column %q: %v", nc.colname, r)
		}
	}()
	var nullable sql.NullInt64
	if err = nullable.Scan(v); err != nil {
		return fmt.Errorf("cannot scan column %q: %v", nc.colname, err)
	}
	if nullable.Valid {
		nc.cellValue.SetUint(uint64(nullable.Int64))
	} else {
		nc.cellValue.SetUint(0)
	}
	return nil
}

type nullFloatCell struct {
	colname   string
	cellValue reflect.Value
	bits      int
}

func (nc *nullFloatCell) Scan(v interface{}) (err error) {
	defer func() {
		// handle panic if SetFloat overflows
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot scan column %q: %v", nc.colname, r)
		}
	}()
	var nullable sql.NullFloat64
	if err := nullable.Scan(v); err != nil {
		return fmt.Errorf("cannot scan column %q: %v", nc.colname, err)
	}
	if nullable.Valid {
		nc.cellValue.SetFloat(nullable.Float64)
	} else {
		nc.cellValue.SetFloat(0.0)
	}
	return nil
}

type nullBoolCell struct {
	colname   string
	cellValue reflect.Value
}

func (nc *nullBoolCell) Scan(v interface{}) error {
	var nullable sql.NullBool
	if err := nullable.Scan(v); err != nil {
		return fmt.Errorf("cannot scan column %q: %v", nc.colname, err)
	}
	if nullable.Valid {
		nc.cellValue.SetBool(nullable.Bool)
	} else {
		nc.cellValue.SetBool(false)
	}
	return nil
}

type nullStringCell struct {
	colname   string
	cellValue reflect.Value
}

func (nc *nullStringCell) Scan(v interface{}) error {
	var nullable sql.NullString
	if err := nullable.Scan(v); err != nil {
		return fmt.Errorf("cannot scan column %q: %v", nc.colname, err)
	}
	if nullable.Valid {
		nc.cellValue.SetString(nullable.String)
	} else {
		nc.cellValue.SetString("")
	}
	return nil
}

type nullTimeCell struct {
	colname   string
	cellValue reflect.Value
}

func (nc *nullTimeCell) Scan(v interface{}) error {
	if v == nil {
		nc.cellValue.Set(timeZero)
		return nil
	}
	switch v.(type) {
	case time.Time:
		nc.cellValue.Set(reflect.ValueOf(v))
		return nil
	}

	return fmt.Errorf("cannot scan column %q: type %q is not compatible with time.Time", nc.colname, reflect.TypeOf(v))
}
