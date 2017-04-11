package sqlrow

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jjeffery/sqlrow/private/column"
	"github.com/jjeffery/sqlrow/private/scanner"
)

// The columnNamer interface is used for naming columns.
type columnNamer interface {
	ColumnName(col *column.Info) string
}

// columnNamerFunc converts a function into a columnNamer.
type columnNamerFunc func(*column.Info) string

func (f columnNamerFunc) ColumnName(col *column.Info) string {
	return f(col)
}

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

func newColumns(allColumns []*column.Info) columnList {
	return columnList{
		allColumns: allColumns,
		clause:     clauseSelectColumns,
	}
}

// Parse parses the text inside the curly braces to obtain more information
// about how to render the column list. It is not very sophisticated at the moment,
// currently the only recognised values are:
//  "alias n" => use alias "n" for each column in the list
//  "pk"      => primary key columns only
//  "all"     => all columns
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

// filtered returns the columns after the filter has been applied
func (cols columnList) filtered() []*column.Info {
	v := make([]*column.Info, 0, len(cols.allColumns))
	for _, col := range cols.allColumns {
		if cols.filter == nil || cols.filter(col) {
			v = append(v, col)
		}
	}
	return v
}

// columnFilter is the filter for all columns
func columnFilterAll(col *column.Info) bool {
	return true
}

// columnFilterPK is the filter for primary key columns only
func columnFilterPK(col *column.Info) bool {
	return col.Tag.PrimaryKey
}

// columnFilterInsertable is the filter for all columns except the autoincrement
// column (if it exists)
func columnFilterInsertable(col *column.Info) bool {
	return !col.Tag.AutoIncrement
}

// columnFitlerUpdateable is the filter for all columns not part of the primary key,
// and not autoincrement
func columnFilterUpdateable(col *column.Info) bool {
	return !col.Tag.PrimaryKey && !col.Tag.AutoIncrement
}
