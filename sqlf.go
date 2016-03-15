// Package sqlf assists writing SQL statements.
// It is intended for programmers who are comfortable with
// writing SQL, but would like assistance with the tedious
// process of preparing SELECT, INSERT and UPDATE statements
// for tables that have a large number of columns.
package sqlf

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// execCommand handles inserting or updating a single table at a time.
// A future implementation might be able to handle updates involving
// multiple tables, but keeping it simple for now.
type execCommand struct {
	command string
	table   *TableInfo
	inputs  []columnInfo
}

func (cmd execCommand) Command() string {
	return cmd.command
}

func (cmd execCommand) Args(row interface{}) ([]interface{}, error) {
	if cmd.table == nil {
		return nil, errors.New("table not specified")
	}
	var args []interface{}

	rowVal := reflect.ValueOf(row)
	for rowVal.Type().Kind() == reflect.Ptr {
		rowVal = rowVal.Elem()
	}
	if rowVal.Type() != cmd.table.rowType {
		return nil, fmt.Errorf("Args: expected type %s.%s or pointer", cmd.table.rowType.PkgPath(), cmd.table.rowType.Name())
	}

	for _, ci := range cmd.inputs {
		args = append(args, rowVal.Field(ci.fieldIndex).Interface())
	}

	return args, nil
}

func (cmd execCommand) Exec(db Execer, row interface{}) (sql.Result, error) {
	return nil, nil
}

func Insertf(format string, args ...interface{}) ExecCommand {
	cmd := execCommand{
		command: fmt.Sprintf(format, args...),
	}

	for _, arg := range args {
		if tn, ok := arg.(TableName); ok {
			if tn.clause == ClauseInsertInto {
				cmd.table = tn.table
			}
		}
		if cil, ok := arg.(ColumnList); ok {
			if cil.clause.isInput() {
				// input parameters for the INSERT statement
				cmd.inputs = append(cmd.inputs, cil.filtered()...)
			}
		}
	}
	return cmd
}

func Updatef(format string, args ...interface{}) ExecCommand {
	cmd := execCommand{
		command: fmt.Sprintf(format, args...),
	}

	for _, arg := range args {
		if tn, ok := arg.(TableName); ok {
			if tn.clause == ClauseUpdateTable {
				cmd.table = tn.table
			}
		}
		if cil, ok := arg.(ColumnList); ok {
			if cil.clause.isInput() {
				// input parameters for the INSERT statement
				cmd.inputs = append(cmd.inputs, cil.filtered()...)
			}
		}
	}
	return cmd
}

func Selectf(format string, args ...interface{}) QueryCommand {
	// TODO: not implemented
	return nil
}

// TableInfo contains enough information about a database table
// to assist with generating SQL strings.
type TableInfo struct {
	TableName string
	Select    SelectInfo
	Insert    InsertInfo
	Update    UpdateInfo

	rowType reflect.Type
	columns ColumnList
	dialect Dialect
}

// Table creates a TableInfo with the specified table name
// and schema as defined by the struct that is pointed to
// by row.
//
//This function wil panic if row is not a struct
// or a pointer to a struct. The contents of row
// are ignored, only the structure fields and field tags
// are used.
func Table(name string, row interface{}) *TableInfo {
	ti := &TableInfo{TableName: name}

	ti.rowType = reflect.TypeOf(row)
	for ti.rowType.Kind() == reflect.Ptr {
		// derefernce pointer(s)
		ti.rowType = ti.rowType.Elem()
	}
	if ti.rowType.Kind() != reflect.Struct {
		panic("sqlf.Table: expected struct or pointer to struct")
	}

	for i := 0; i < ti.rowType.NumField(); i++ {
		field := ti.rowType.Field(i)

		// For compatibility, use Gorm's tag formats, as they have
		// all the information we need. This means you can interop
		// using Gorm with this package if you like.
		if field.Tag.Get("sql") == "-" {
			// ignore field
			continue
		}
		tagSettings := parseTagSetting(field.Tag)

		ci := columnInfo{
			table:      ti,
			fieldName:  field.Name,
			fieldIndex: i,
		}

		if value, ok := tagSettings["COLUMN"]; ok {
			ci.columnName = value
		} else {
			ci.columnName = ToDBName(ci.fieldName)
		}
		if _, ok := tagSettings["PRIMARY_KEY"]; ok {
			ci.primaryKey = true
		} else if strings.ToLower(ci.fieldName) == "id" {
			ci.primaryKey = true
		}
		if _, ok := tagSettings["AUTO_INCREMENT"]; ok {
			ci.autoIncrement = true
		}
		ti.columns.add(ci)
	}

	ti.Select.TableName = TableName{clause: ClauseSelectColumns, table: ti}
	ti.Select.Columns = ColumnList{clause: ClauseSelectColumns, list: ti.columns.list}.All()
	ti.Insert.TableName = TableName{clause: ClauseInsertInto, table: ti}
	ti.Insert.Columns = ColumnList{clause: ClauseInsertColumns, list: ti.columns.list}.Insertable()
	ti.Insert.Values = ColumnList{clause: ClauseInsertValues, list: ti.columns.list}.Insertable()
	ti.Update.TableName = TableName{clause: ClauseUpdateTable, table: ti}
	ti.Update.SetColumns = ColumnList{clause: ClauseUpdateSet, list: ti.columns.list}.Updateable()
	ti.Update.WhereColumns = ColumnList{clause: ClauseUpdateWhere, list: ti.columns.list}.PrimaryKey()

	return ti
}

// WithDialect sets the SQL dialect for this table.
func (ti *TableInfo) WithDialect(dialect Dialect) *TableInfo {
	ti.dialect = dialect
	return ti
}

// Dialect returns the SQL dialect to use with this table.
func (ti *TableInfo) Dialect() Dialect {
	if ti.dialect != nil {
		return ti.dialect
	}
	return defaultDialect()
}

// SelectInfo contains information about a table that will
// be formatted for a SELECT clause.
type SelectInfo struct {
	TableName TableName
	Columns   ColumnList
}

type UpdateInfo struct {
	TableName    TableName
	SetColumns   ColumnList
	WhereColumns ColumnList
}

type InsertInfo struct {
	TableName TableName
	Columns   ColumnList
	Values    ColumnList
}

// ColumnInfo contains enough information about a database column
// to assist with generating SQL strings.
type columnInfo struct {
	fieldName     string
	columnName    string
	table         *TableInfo
	primaryKey    bool
	autoIncrement bool
	version       bool
	fieldIndex    int

	// varies with each command
	columnAlias string
	tableAlias  string
	placeholder int
}

// Clause represents a specific SQL clause. Column lists
// and table names are represented differently depending on
// which SQL clause they appear in.
type Clause int

// All of the different clauses of an SQL statement where columns
// and table names can be found.
const (
	ClauseSelectColumns Clause = iota
	ClauseSelectFrom
	ClauseInsertInto
	ClauseInsertColumns
	ClauseInsertValues
	ClauseUpdateTable
	ClauseUpdateSet
	ClauseUpdateWhere
)

func (c Clause) isInput() bool {
	return c == ClauseInsertValues ||
		c == ClauseUpdateSet ||
		c == ClauseUpdateWhere
}

type TableName struct {
	table  *TableInfo
	clause Clause

	// varies with each command
	alias string
}

func (tn TableName) clone() TableName {
	return TableName{
		table:  tn.table,
		clause: tn.clause,
		alias:  tn.alias,
	}
}

// Clause identifies the SQL clause that this table name applies to.
func (tn TableName) Clause() Clause {
	return tn.clause
}

func (tn TableName) String() string {
	dialect := tn.table.Dialect()
	switch tn.clause {
	case ClauseSelectFrom:
		if tn.alias != "" {
			return fmt.Sprintf("%s as %s",
				dialect.Quote(tn.table.TableName),
				tn.alias,
			)
		}
		return dialect.Quote(tn.table.TableName)
	case ClauseInsertInto, ClauseUpdateTable:
		return dialect.Quote(tn.table.TableName)
	}
	panic(fmt.Sprintf("invalid clause for table name: %d", tn.clause))
}

// ColumnList represents a list of columns associated
// with a table for use in a specific SQL clause.
type ColumnList struct {
	list   []columnInfo
	filter func(ci columnInfo) bool
	clause Clause
}

func (cil ColumnList) clone() ColumnList {
	cil2 := ColumnList{
		list:   make([]columnInfo, len(cil.list)),
		filter: cil.filter,
		clause: cil.clause,
	}
	copy(cil2.list, cil.list)
	return cil2
}

func (cil *ColumnList) add(ci columnInfo) {
	cil.list = append(cil.list, ci)
}

func (cil ColumnList) filtered() []columnInfo {
	if cil.filter == nil {
		return cil.list
	}
	var list []columnInfo
	for _, ci := range cil.list {
		if cil.filter(ci) {
			list = append(list, ci)
		}
	}
	return list
}

// Clause returns the SQL clause that this column list applies to.
func (cil ColumnList) Clause() Clause {
	return cil.clause
}

// String prints the columns in the list in the appropriate
// form for the part of the SQL statement that this column
// list applies to. Because ColumnList implements the Stringer
// interface, it can be formatted using "%s" in fmt.Sprintf.
func (cil ColumnList) String() string {
	var buf bytes.Buffer
	for i, ci := range cil.filtered() {
		if i > 0 {
			if cil.clause == ClauseUpdateWhere {
				buf.WriteString(" and ")
			} else {
				buf.WriteRune(',')
			}
		}
		switch cil.clause {
		case ClauseSelectColumns:
			if ci.tableAlias != "" {
				buf.WriteString(ci.tableAlias)
				buf.WriteRune('.')
			}
			buf.WriteString(ci.table.Dialect().Quote(ci.columnName))
			if ci.columnAlias != "" {
				buf.WriteString(" as ")
				buf.WriteString(ci.columnAlias)
			}
		case ClauseInsertColumns:
			buf.WriteString(ci.table.Dialect().Quote(ci.columnName))
		case ClauseInsertValues:
			buf.WriteString(ci.table.Dialect().Placeholder(ci.placeholder))
		case ClauseUpdateSet, ClauseUpdateWhere:
			buf.WriteString(ci.table.Dialect().Quote(ci.columnName))
			buf.WriteRune('=')
			buf.WriteString(ci.table.Dialect().Placeholder(ci.placeholder))
		}
	}
	return buf.String()
}

// All returns a list of all of the columns in the associated table.
func (cil ColumnList) All() ColumnList {
	return ColumnList{clause: cil.clause, list: cil.list}
}

// Filter returns a column list of all columns in the
// table for which the filter function f returns true.
func (cil ColumnList) Filter(f func(ci columnInfo) bool) ColumnList {
	return ColumnList{
		clause: cil.clause,
		list:   cil.list,
		filter: f,
	}
}

// PrimaryKey returns a column list of all primary key columns in the
// associated table.
func (cil ColumnList) PrimaryKey() ColumnList {
	return cil.Filter(func(ci columnInfo) bool {
		return ci.primaryKey
	})
}

// Insertable returns a column list of all columns in the associated
// table that can be inserted. This list includes all columns except
// an auto-increment column, if the table has one.
func (cil ColumnList) Insertable() ColumnList {
	return cil.Filter(func(ci columnInfo) bool {
		return !ci.autoIncrement
	})
}

// Updateable returns a column list of all columns that can be
// updated in the associated table. This list excludes any
// primary key columns and any auto-increment column.
func (cil ColumnList) Updateable() ColumnList {
	return cil.Filter(func(ci columnInfo) bool {
		return !ci.primaryKey && !ci.autoIncrement
	})
}

// Include returns a column list of all columns whose name
// is in the list of names. For consistency, use the name
// of the Go struct field. This function will, however, match
// the name of the DB table column as well.
func (cil ColumnList) Include(names ...string) ColumnList {
	return cil.Filter(func(ci columnInfo) bool {
		for _, name := range names {
			if name == ci.fieldName || name == ci.columnName {
				return true
			}
		}
		return false
	})
}

type Placeholder struct {
	table    *TableInfo
	position int
}

func (p Placeholder) String() string {
	return p.table.Dialect().Placeholder(p.position)
}

// Execer is an interface for the Exec method. This interface
// is implemented by the *sql.DB and *sql.Tx types in the
// standard library database/sql package.
type Execer interface {
	Exec(command string, args ...interface{}) (sql.Result, error)
}

// ExecCommand is the return value when creating an SQL command
// that does not return rows (ie INSERT, UPDATE, DELETE). It contains
// all the information required to execute the command against the database.
type ExecCommand interface {
	Command() string
	Args(row interface{}) ([]interface{}, error)
	Exec(db Execer, row interface{}) (sql.Result, error)
}

// Queryer is an interface for the Query method. This interface
// is implemented by the *sql.DB and *sql.Tx types in the standard
// library database/sql package.
type Queryer interface {
	Query(command string, args ...interface{}) (*sql.Rows, error)
}

// QueryCommand is the return value when creating an SQL command
// that return rows (ie SELECT). It contains all the information
// required to execute the command against the database.
type QueryCommand interface {
	Command() string
	Args(args ...interface{}) ([]interface{}, error)
	Query(db Queryer, args ...interface{}) (*sql.Rows, error)
	Scan(rows *sql.Rows, dest ...interface{}) error
}
