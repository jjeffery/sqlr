package sqlr

import (
	"fmt"
	"reflect"

	"github.com/jjeffery/sqlr/private/column"
)

// Table represents the information known about a database table
// and its associated row struct type.
type Table struct {
	schema    *Schema
	rowType   reflect.Type
	tableName string
	cols      []*Column
	pk        []*Column
	nk        []*Column
}

// TableFor returns the table information associated with
// row, which should be an instance of a struct type
// or a pointer to a struct type.
// If row does not refer to a struct type then a panic results.
func (s *Schema) TableFor(row interface{}) *Table {
	var rowType reflect.Type
	if t, ok := row.(reflect.Type); ok {
		rowType = t
	} else {
		rowType = reflect.TypeOf(row)
	}
	for rowType.Kind() != reflect.Struct {
		switch rowType.Kind() {
		case reflect.Ptr, reflect.Slice, reflect.Array, reflect.Chan, reflect.Map:
			rowType = rowType.Elem()
			break
		default:
			err := fmt.Errorf("exected rowType to be a struct, found %v", rowType.String())
			panic(err)
		}
	}
	// TODO(jpj): lookup cache, schema based.
	return newTable(s, rowType)
}

func newTable(schema *Schema, rowType reflect.Type) *Table {
	columnNamer := schema.columnNamer()
	tableNamer := schema.tableNamer()

	tbl := &Table{
		schema:    schema,
		rowType:   rowType,
		tableName: tableNamer(rowType),
	}

	for _, colInfo := range column.ListForType(rowType) {
		if colInfo.Tag.Ignore {
			continue
		}
		col := &Column{
			columnName: columnNamer.ColumnName(colInfo),
		}

		tbl.cols = append(tbl.cols, col)

		if colInfo.Tag.PrimaryKey {
			tbl.pk = append(tbl.pk, col)
		}
		if colInfo.Tag.NaturalKey {
			tbl.nk = append(tbl.nk, col)
		}
	}

	return tbl
}

// Name returns the name of the table.
func (tbl *Table) Name() string {
	return tbl.tableName
}

// RowType returns the row type, which is always a struct.
func (tbl *Table) RowType() reflect.Type {
	return tbl.rowType
}

// PrimaryKey returns the column or columns that form the
// primary key for the table. Returns nil if no primary key
// has been defined.
func (tbl *Table) PrimaryKey() []*Column {
	return columnSlice(tbl.pk)
}

// NaturalKey returns the natural key columns for the table.
// Returns nil if no natural key columns have been defined.
// Natural key columns are columns that are useful for identifying
// a row. They are used in error messages only. (And we might remove
// them to make the API simpler to start with).
func (tbl *Table) NaturalKey() []*Column {
	return columnSlice(tbl.nk)
}

// Columns returns all columns defined for the table.
func (tbl *Table) Columns() []*Column {
	return columnSlice(tbl.cols)
}

// Column represents a table column.
type Column struct {
	columnName string
	info       *column.Info
}

// Name returns the name of the database column.
func (col *Column) Name() string {
	return col.columnName
}

// PrimaryKey returns true if this column is the primary key,
// or forms part of the primary key.
func (col *Column) PrimaryKey() bool {
	return col.info.Tag.PrimaryKey
}

// AutoIncrement returns true if this column is an auto-increment
// column.
func (col *Column) AutoIncrement() bool {
	return col.info.Tag.AutoIncrement
}

// Version returns true if this  column is an optimistic locking version column.
func (col *Column) Version() bool {
	return col.info.Tag.Version
}

// EmptyIsNull returns true if the empty value for the associated field type
// should be stored as NULL in the database, and if the NULL value in the
// database should be stored in the associated field as the empty (or zero)
// value.
//
// This is commonly set for string values and time.Time values. It is common
// for an empty string value or an empty time.Time value to be represented
// as a database NULL.
func (col *Column) EmptyIsNull() bool {
	return col.info.Tag.EmptyNull
}

// StoreAsJSON returns true if column's value is unmarshaled from JSON into
// the associated struct field, and if the struct field is marshaled into
// JSON to be stored in the database column.
func (col *Column) StoreAsJSON() bool {
	return col.info.Tag.JSON
}

func columnSlice(src []*Column) []*Column {
	dest := make([]*Column, len(src), len(src))
	copy(dest, src)
	return dest
}
