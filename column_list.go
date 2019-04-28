package sqlr

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jjeffery/sqlr/private/scanner"
)

// columnList represents a list of columns for use in an SQL clause.
//
// Each columnList represents a subset of the available columns.
// For example a column list for the WHERE clause in a row update
// statement will only contain the columns for the primary key.
type columnList struct {
	allColumns []*Column
	filter     func(col *Column) bool
	clause     sqlClause
	alias      string
	exclude    map[string]struct{}
}

func newColumns(allColumns []*Column) columnList {
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
	scan.AddKeywords("alias", "all", "pk", "exclude")
	scan.IgnoreWhiteSpace = true

	if scan.Scan() {
		needScan := false
		for {
			if needScan {
				needScan = false
				if !scan.Scan() {
					break
				}
			}
			tok, lit := scan.Token(), scan.Text()
			if tok == scanner.EOF {
				break
			}

			// TODO: dodgy job to get going quickly
			if tok == scanner.KEYWORD {
				switch strings.ToLower(lit) {
				case "alias":
					if scan.Scan() {
						cols2.alias = scan.Text()
						needScan = true
					} else {
						return columnList{}, fmt.Errorf("missing ident after 'alias'")
					}
				case "all":
					cols2.filter = columnFilterAll
					needScan = true
				case "pk":
					cols2.filter = columnFilterPK
					needScan = true
				case "exclude":
					if scan.Scan() {
						if cols2.exclude == nil {
							cols2.exclude = make(map[string]struct{})
						}
						cols2.exclude[scan.Text()] = struct{}{}
						for {
							if !scan.Scan() {
								break
							}
							if scan.Text() != "," {
								break
							}
							if !scan.Scan() {
								break
							}
							cols2.exclude[scan.Text()] = struct{}{}
						}
					} else {
						return columnList{}, fmt.Errorf("missing column after 'exclude'")
					}
				}
			} else {
				needScan = true
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
func (cols columnList) String(dialect Dialect, counter func() int) string {
	var buf bytes.Buffer

	quotedColumnName := func(col *Column) string {
		return dialect.Quote(col.Name())
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
				buf.WriteString(", ")
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
			buf.WriteString(" = ")
			buf.WriteString(placeholder())
		}
	}
	return buf.String()
}

// filtered returns the columns after the filter has been applied
func (cols columnList) filtered() []*Column {
	v := make([]*Column, 0, len(cols.allColumns))
	for _, col := range cols.allColumns {
		if cols.filter == nil || cols.filter(col) {
			if _, ok := cols.exclude[col.columnName]; ok {
				continue
			}
			v = append(v, col)
		}
	}
	return v
}

// columnFilter is the filter for all columns
func columnFilterAll(col *Column) bool {
	return true
}

// columnFilterPK is the filter for primary key columns only
func columnFilterPK(col *Column) bool {
	return col.PrimaryKey()
}

// columnFilterInsertable is the filter for all columns except the autoincrement
// column (if it exists)
func columnFilterInsertable(col *Column) bool {
	return !col.AutoIncrement()
}

// columnFitlerUpdateable is the filter for all columns not part of the primary key,
// and not autoincrement
func columnFilterUpdateable(col *Column) bool {
	return !col.PrimaryKey() && !col.AutoIncrement()
}
