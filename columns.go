package sqlf

import (
	"bytes"

	"github.com/jjeffery/sqlf/private/column"
)

// Columns represents a set of columns associated with
// a table for use in a specific SQL clause.
//
// Each Columns set represents a subset of the columns
// in the table. For example a column list for the WHERE
// clause in a row update statement will only contain the
// columns for the primary key.
type Columns struct {
	table  *TableInfo
	filter func(col *column.Info) bool
	clause sqlClause
	alias  string
}

func newColumns(table *TableInfo, clause sqlClause) Columns {
	return Columns{
		table:  table,
		clause: clause,
	}
}

// Alias returns a columns collection with the specified alias.
func (cols Columns) Alias(alias string) Columns {
	cols2 := cols
	cols2.alias = alias
	return cols2
}

// PK returns a columns collection that contains
// the primary key column or columns.
func (cols Columns) PK() Columns {
	cols2 := cols
	cols2.filter = func(col *column.Info) bool {
		return col.PrimaryKey
	}
	return cols2
}

// PKV returns a columns collection that contains the
// primary key column or columns, plus the optimistic
// locking version column, if it exists.
func (cols Columns) PKV() Columns {
	cols2 := cols
	cols2.filter = func(col *column.Info) bool {
		return col.PrimaryKey || col.Version
	}
	return cols2
}

// String returns a string representation of the columns.
// The string returned depends on the SQL clause in which the
// columns appear.
func (cols Columns) String() string {
	var buf bytes.Buffer
	dialect := cols.table.dialect
	for i, col := range cols.filtered() {
		if i > 0 {
			if cols.clause == clauseUpdateWhere {
				buf.WriteString(" and ")
			} else {
				buf.WriteRune(',')
			}
		}
		switch cols.clause {
		case clauseSelectColumns, clauseSelectOrderBy:
			if cols.alias != "" {
				buf.WriteString(dialect.Quote(cols.alias))
				buf.WriteRune('.')
			}
			buf.WriteString(cols.columnName(col))
		case clauseInsertColumns:
			buf.WriteString(cols.columnName(col))
		case clauseInsertValues:
			buf.WriteString("?")
		case clauseUpdateSet, clauseUpdateWhere, clauseDeleteWhere:
			buf.WriteString(cols.columnName(col))
			buf.WriteString("=?")
		}
	}
	return buf.String()
}

func (cols Columns) columnName(info *column.Info) string {
	var path = info.Path
	var columnName string

	for _, field := range path {
		name := field.ColumnName
		if name == "" {
			name = cols.table.convention.ColumnName(field.FieldName)
		}
		if columnName == "" {
			columnName = name
		} else {
			columnName = cols.table.convention.Join(columnName, name)
		}
	}
	return cols.table.dialect.Quote(columnName)
}

func (cols Columns) filtered() []*column.Info {
	vec := make([]*column.Info, 0, len(cols.table.fields))
	for _, col := range cols.table.fields {
		if cols.filter == nil || cols.filter(col) {
			vec = append(vec, col)
		}
	}
	return vec
}

func (cols Columns) updateable() Columns {
	cols2 := cols
	cols2.filter = func(col *column.Info) bool {
		return !col.AutoIncrement && !col.PrimaryKey
	}
	return cols2
}

func (cols Columns) insertable() Columns {
	cols2 := cols
	cols2.filter = func(col *column.Info) bool {
		return !col.AutoIncrement
	}
	return cols2
}
