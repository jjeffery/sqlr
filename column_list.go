package sqlrow

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jjeffery/sqlrow/private/column"
	"github.com/jjeffery/sqlrow/private/scanner"
)

// columnList represents a list of columns for use in an SQL clause.
//
// Each columnList represents a subset of the available columns.
// For example a column list for the WHERE clause in a row update
// statement will only contain the columns for the primary key.
type columnList struct {
	allColumns []*column.Info
	filter     func(col *column.Info) bool
	clause     sqlClause
	alias      string
}

type columnNamer interface {
	ColumnName(col *column.Info) string
}

type columnNamerFunc func(*column.Info) string

func (f columnNamerFunc) ColumnName(col *column.Info) string {
	return f(col)
}

func newColumns(allColumns []*column.Info) columnList {
	return columnList{
		allColumns: allColumns,
		clause:     clauseSelectColumns,
	}
}

func (cols columnList) Parse(clause sqlClause, text string) (columnList, error) {
	cols2 := cols
	cols2.clause = clause
	cols2.filter = clause.defaultFilter()

	// TODO: update filter based on text
	scan := scanner.New(strings.NewReader(text))
	scan.AddKeywords("alias", "all", "pk")
	scan.IgnoreWhiteSpace = true

	for scan.Scan() {
		tok, lit := scan.Token(), scan.Text()

		// TODO: dodgy job to get going quickly
		if tok == scanner.KEYWORD {
			switch strings.ToLower(lit) {
			case "alias":
				if scan.Scan() {
					cols2.alias = scan.Text()
				} else {
					return columnList{}, fmt.Errorf("missing ident after 'alias'")
				}
			case "all":
				cols2.filter = columnFilterAll
			case "pk":
				cols2.filter = columnFilterPK
			}
		}
	}
	if err := scan.Err(); err != nil {
		return columnList{}, err
	}

	return cols2, nil
}

// String returns a string representation of the columns.
// The string returned depends on the SQL clause in which the
// columns appear.
func (cols columnList) String(dialect Dialect, columnNamer columnNamer, counter func() int) string {
	var buf bytes.Buffer

	quotedColumnName := func(col *column.Info) string {
		return dialect.Quote(columnNamer.ColumnName(col))
	}
	placeholder := func() string {
		return dialect.Placeholder(counter())
	}

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
			buf.WriteString(quotedColumnName(col))
		case clauseInsertColumns:
			buf.WriteString(quotedColumnName(col))
		case clauseInsertValues:
			buf.WriteString(placeholder())
		case clauseUpdateSet, clauseUpdateWhere, clauseDeleteWhere, clauseSelectWhere:
			if cols.alias != "" {
				buf.WriteString(cols.alias)
				buf.WriteRune('.')
			}
			buf.WriteString(quotedColumnName(col))
			buf.WriteRune('=')
			buf.WriteString(placeholder())
		}
	}
	return buf.String()
}

func (cols columnList) filtered() []*column.Info {
	v := make([]*column.Info, 0, len(cols.allColumns))
	for _, col := range cols.allColumns {
		if cols.filter == nil || cols.filter(col) {
			v = append(v, col)
		}
	}
	return v
}

func columnFilterAll(col *column.Info) bool {
	return true
}

func columnFilterPK(col *column.Info) bool {
	return col.Tag.PrimaryKey
}

func columnFilterInsertable(col *column.Info) bool {
	return !col.Tag.AutoIncrement
}

func columnFilterUpdateable(col *column.Info) bool {
	return !col.Tag.PrimaryKey && !col.Tag.AutoIncrement
}
