package sqlstmt

import (
	"bytes"
	"strings"

	"github.com/jjeffery/sqlstmt/private/column"
	"github.com/jjeffery/sqlstmt/private/scanner"
)

// columnsT represents a set of columns associated with
// a table for use in a specific SQL clause.
//
// Each columnsT set represents a subset of the columns
// in the table. For example a column list for the WHERE
// clause in a row update statement will only contain the
// columns for the primary key.
type columnsT struct {
	allColumns []*column.Info
	convention Convention
	dialect    Dialect
	counter    func() int
	filter     func(col *column.Info) bool
	clause     sqlClause
	alias      string
}

func newColumns(allColumns []*column.Info, convention Convention, dialect Dialect, counter func() int) columnsT {
	return columnsT{
		allColumns: allColumns,
		convention: convention,
		dialect:    dialect,
		counter:    counter,
		clause:     clauseSelectColumns,
	}
}

func (cols columnsT) Parse(clause sqlClause, text string) (columnsT, error) {
	cols2 := cols
	cols2.clause = clause
	cols2.filter = clause.defaultFilter()

	// TODO: update filter based on text
	scan := scanner.New(strings.NewReader(text))
	scan.AddKeywords("alias", "all", "pk")

	for {
		tok, lit := scan.Scan()
		if tok == scanner.EOF {
			break
		}

		// TODO: dodgy job to get going quickly
		if tok == scanner.KEYWORD {
			switch strings.ToLower(lit) {
			case "alias":
				tok2, lit2 := scan.Scan()
				for tok2 == scanner.WS {
					tok2, lit2 = scan.Scan()
				}
				cols2.alias = lit2
			case "all":
				cols2.filter = columnFilterAll
			case "pk":
				cols2.filter = columnFilterPK
			}
		}
	}

	return cols2, nil
}

// Alias returns a columns collection with the specified alias.
func (cols columnsT) Alias(alias string) columnsT {
	cols2 := cols
	cols2.alias = alias
	return cols2
}

// PK returns a columns collection that contains
// the primary key column or columns.
func (cols columnsT) PK() columnsT {
	cols2 := cols
	cols2.filter = func(col *column.Info) bool {
		return col.PrimaryKey
	}
	return cols2
}

// PKV returns a columns collection that contains the
// primary key column or columns, plus the optimistic
// locking version column, if it exists.
func (cols columnsT) PKV() columnsT {
	cols2 := cols
	cols2.filter = func(col *column.Info) bool {
		return col.PrimaryKey || col.Version
	}
	return cols2
}

// String returns a string representation of the columns.
// The string returned depends on the SQL clause in which the
// columns appear.
func (cols columnsT) String() string {
	var buf bytes.Buffer
	for i, col := range cols.filtered() {
		if i > 0 {
			if cols.clause.matchAny(
				clauseUpdateWhere,
				clauseDeleteWhere,
				clauseSelectWhere) {
				buf.WriteString(" and ")
			} else {
				buf.WriteRune(',')
			}
		}
		switch cols.clause {
		case clauseSelectColumns, clauseSelectOrderBy:
			if cols.alias != "" {
				buf.WriteString(cols.alias)
				buf.WriteRune('.')
			}
			buf.WriteString(cols.columnName(col))
		case clauseInsertColumns:
			buf.WriteString(cols.columnName(col))
		case clauseInsertValues:
			buf.WriteString(cols.dialect.Placeholder(cols.counter()))
		case clauseUpdateSet, clauseUpdateWhere, clauseDeleteWhere, clauseSelectWhere:
			buf.WriteString(cols.columnName(col))
			buf.WriteRune('=')
			buf.WriteString(cols.dialect.Placeholder(cols.counter()))
		}
	}
	return buf.String()
}

func (cols columnsT) columnName(info *column.Info) string {
	var path = info.Path
	var columnName string

	for _, field := range path {
		name := field.ColumnName
		if name == "" {
			name = cols.convention.ColumnName(field.FieldName)
		}
		if columnName == "" {
			columnName = name
		} else {
			columnName = cols.convention.Join(columnName, name)
		}
	}
	return cols.dialect.Quote(columnName)
}

func (cols columnsT) filtered() []*column.Info {
	v := make([]*column.Info, 0, len(cols.allColumns))
	for _, col := range cols.allColumns {
		if cols.filter == nil || cols.filter(col) {
			v = append(v, col)
		}
	}
	return v
}

func (cols columnsT) updateable() columnsT {
	cols2 := cols
	cols2.filter = columnFilterUpdateable
	return cols2
}

func (cols columnsT) insertable() columnsT {
	cols2 := cols
	cols2.filter = columnFilterInsertable
	return cols2
}

func columnFilterAll(col *column.Info) bool {
	return true
}

func columnFilterPK(col *column.Info) bool {
	return col.PrimaryKey
}

func columnFilterInsertable(col *column.Info) bool {
	return !col.AutoIncrement
}

func columnFilterUpdateable(col *column.Info) bool {
	return !col.PrimaryKey && !col.AutoIncrement
}
