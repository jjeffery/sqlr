package sqlr

import (
	"fmt"
	"reflect"

	"github.com/jjeffery/sqlr/private/column"
)

// Table represents the information known about a database table.
// The information known about a table is derived from the row
// type (which must be a struct), and any configuration that was
// provided via the SchemaConfig when the schema was created.
type Table struct {
	schema    *Schema
	rowType   reflect.Type
	tableName string
	cols      []*Column
	pk        []*Column
	nk        []*Column
}

// getRowType converts a row instance into a row type.
// Returns an error if row does not refer to a struct type.
func getRowType(row interface{}) (reflect.Type, error) {
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
			err := fmt.Errorf("expected row type to be a struct, found %v", rowType.String())
			return nil, err
		}
	}
	return rowType, nil
}

// newTable returns a new Table value for the row type. If cfg is non-nil,
// then it must have already been checked for any inconsistencies.
func newTable(schema *Schema, rowType reflect.Type, cfg *TableConfig) *Table {
	columnNamer := schema.columnNamer()

	var tableName string
	if cfg != nil && cfg.TableName != "" {
		tableName = cfg.TableName
	} else if rowTypeName := rowType.Name(); rowTypeName != "" {
		convention := schema.convention
		if convention == nil {
			convention = defaultNamingConvention
		}
		tableName = convention.TableName(rowTypeName)
	} else {
		// TODO(jpj): this will happen for anonymous types.
		// Need a mechanism to specify table name using struct tags.
		// eg
		//  TableName sqlr.Name `sql:"table_name_here"`
		//
		// or
		//  ID int64 `sql:"primary key" table:"table_name_here"`
		tableName = "__unknown_table_name__"
	}

	tbl := &Table{
		schema:    schema,
		rowType:   rowType,
		tableName: tableName,
	}

	for _, colInfo := range column.ListForType(rowType) {
		if colInfo.Tag.Ignore {
			continue
		}
		var colConfig ColumnConfig
		var hasColConfig bool
		if cfg != nil {
			colConfig, hasColConfig = cfg.Columns[colInfo.FieldNames]
		}
		if colConfig.Ignore {
			continue
		}

		col := &Column{
			columnName:    columnNamer.ColumnName(colInfo),
			info:          colInfo,
			primaryKey:    colInfo.Tag.PrimaryKey,
			autoIncrement: colInfo.Tag.AutoIncrement,
			emptyNull:     colInfo.Tag.EmptyNull,
			json:          colInfo.Tag.JSON,
			naturalKey:    colInfo.Tag.NaturalKey,
		}

		if hasColConfig {
			if colConfig.ColumnName != "" {
				col.columnName = colConfig.ColumnName
			}
			if colConfig.OverrideStructTag {
				col.primaryKey = colConfig.PrimaryKey
				col.autoIncrement = colConfig.AutoIncrement
				col.emptyNull = colConfig.EmptyNull
				col.json = colConfig.JSON
				col.naturalKey = colConfig.NaturalKey
			} else {
				col.primaryKey = col.primaryKey || colConfig.PrimaryKey
				col.autoIncrement = col.autoIncrement || colConfig.AutoIncrement
				col.emptyNull = col.emptyNull || colConfig.EmptyNull
				col.json = col.json || colConfig.JSON
				col.naturalKey = col.naturalKey || colConfig.NaturalKey
			}
		}

		tbl.cols = append(tbl.cols, col)

		if col.primaryKey {
			tbl.pk = append(tbl.pk, col)
		}
		if col.naturalKey {
			tbl.nk = append(tbl.nk, col)
		}
	}

	return tbl
}

func newTableWithConfig(schema *Schema, rowType reflect.Type, config *TableConfig) (*Table, error) {
	// check that all of the field names in the config match field names in the row type
	if len(config.Columns) > 0 {
		fieldPaths := make(map[string]bool)
		for _, colInfo := range column.ListForType(rowType) {
			fieldPaths[colInfo.FieldNames] = true
		}

		for fieldPath := range config.Columns {
			if !fieldPaths[fieldPath] {
				return nil, fmt.Errorf("field %s not found in type %s", fieldPath, rowType)
			}
		}
	}

	tbl := newTable(schema, rowType, config)

	var versionCols []string
	var autoIncrementCols []string
	for _, col := range tbl.Columns() {
		if col.AutoIncrement() {
			autoIncrementCols = append(autoIncrementCols, col.Name())
		}
		if col.Version() {
			versionCols = append(versionCols, col.Name())
		}
	}

	if len(versionCols) > 1 {
		return nil, fmt.Errorf("%s: multiple version columns not permitted (%v)", rowType, versionCols)
	}
	if len(autoIncrementCols) > 1 {
		return nil, fmt.Errorf("%s: multiple autoincrement columns not permitted (%v)", rowType, versionCols)
	}

	return tbl, nil
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

func (tbl *Table) singular() string {
	return tbl.rowType.Name()
}

func (tbl *Table) plural() string {
	return tbl.singular() + "s"
}

// Column represents a table column.
type Column struct {
	columnName    string
	primaryKey    bool
	autoIncrement bool
	version       bool
	json          bool
	naturalKey    bool
	emptyNull     bool

	info *column.Info
}

// Name returns the name of the database column.
func (col *Column) Name() string {
	return col.columnName
}

// fieldType returns the type of the field associated with the column.
func (col *Column) fieldType() reflect.Type {
	return col.info.Field.Type
}

// fieldIndex returns the index sequence for Type.FieldByIndex
func (col *Column) fieldIndex() []int {
	return []int(col.info.Index)
}

// PrimaryKey returns true if this column is the primary key,
// or forms part of the primary key.
func (col *Column) PrimaryKey() bool {
	return col.primaryKey
}

// AutoIncrement returns true if this column is an auto-increment
// column.
func (col *Column) AutoIncrement() bool {
	return col.autoIncrement
}

// Version returns true if this  column is an optimistic locking version column.
func (col *Column) Version() bool {
	return col.version
}

// EmptyNull returns true if the empty value for the associated field type
// should be stored as NULL in the database, and if the NULL value in the
// database should be stored in the associated field as the empty (or zero)
// value.
//
// This is commonly set for string values and time.Time values. It is common
// for an empty string value or an empty time.Time value to be represented
// as a database NULL.
func (col *Column) EmptyNull() bool {
	return col.emptyNull
}

// JSON returns true if column's value is unmarshaled from JSON into
// the associated struct field, and if the struct field is marshaled into
// JSON to be stored in the database column.
func (col *Column) JSON() bool {
	return col.json
}

func columnSlice(src []*Column) []*Column {
	dest := make([]*Column, len(src), len(src))
	copy(dest, src)
	return dest
}
