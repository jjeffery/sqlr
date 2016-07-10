package sqlf

import (
	"errors"
	"reflect"

	"github.com/jjeffery/sqlf/private/column"
)

// TableInfo contains schema information about a database table that can
// be used to build an SQL query.
type TableInfo struct {
	Select struct {
		TableName TableName // Table name for use in SELECT FROM clause
		Columns   Columns   // All columns for SELECT clause
		OrderBy   Columns   // Primary key columns for ORDER BY clause
	}
	Insert struct {
		TableName TableName // Table name for INSERT INTO clause.
		Columns   Columns   // All columns except auto increment for INSERT INTO (...)
		Values    Columns   // Placeholders to match Columns for INSERT INTO (...) VALUES (...)
	}
	Update struct {
		TableName    TableName // Table name for use in UPDATE statement
		SetColumns   Columns   // Non-primary key columns and associated placeholders for UPDATE SET clause
		WhereColumns Columns   // Primary key columns and assocated placeholders for UPDATE WHERE clause
	}
	Delete struct {
		TableName    TableName // Table name for use in DELETE statements.
		WhereColumns Columns   // Primary key columns and assocated placeholders for DELETE WHERE clause
	}

	dialect Dialect
	rowType reflect.Type
	name    string
	alias   string
	fields  []*column.Info
}

func newTable(s *Schema, name string, row interface{}) *TableInfo {
	rowType := reflect.TypeOf(row)
	for rowType.Kind() == reflect.Ptr {
		rowType = rowType.Elem()
	}
	if rowType.Kind() != reflect.Struct {
		panic(errors.New("row must be a struct"))
	}
	table := &TableInfo{
		dialect: s.dialect(),
		rowType: rowType,
		name:    name,
		fields:  column.NewList(row, s.convention()),
	}
	table.Select.TableName = newTableName(table, clauseSelectFrom)
	table.Select.Columns = newColumns(table, clauseSelectColumns)
	table.Select.OrderBy = newColumns(table, clauseSelectOrderBy).PK()

	table.Insert.TableName = newTableName(table, clauseInsertInto)
	table.Insert.Columns = newColumns(table, clauseInsertColumns).insertable()
	table.Insert.Values = newColumns(table, clauseInsertValues).insertable()

	table.Update.TableName = newTableName(table, clauseUpdateTable)
	table.Update.SetColumns = newColumns(table, clauseUpdateSet).updateable()
	table.Update.WhereColumns = newColumns(table, clauseUpdateWhere).PKV()

	table.Delete.TableName = newTableName(table, clauseDeleteTable)
	table.Delete.WhereColumns = newColumns(table, clauseDeleteWhere).PKV()

	return table
}

// Table returns a table info using the default schema,
// the structure of whose rows are represented by
// the structure of row.
func Table(name string, row interface{}) *TableInfo {
	return DefaultSchema.Table(name, row)
}
