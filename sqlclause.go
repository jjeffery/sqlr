package sqlr

import (
	"fmt"
	"strings"
)

// sqlClause represents a specific SQL clause. Column lists
// and table names are represented differently depending on
// which SQL clause they appear in.
type sqlClause int

// All of the different clauses of an SQL statement where columns
// and table names can be found.
const (
	clauseNone sqlClause = iota
	clauseSelectColumns
	clauseSelectFrom
	clauseSelectWhere
	clauseSelectOrderBy
	clauseInsertColumns
	clauseInsertValues
	clauseUpdateTable
	clauseUpdateSet
	clauseUpdateWhere
	clauseDeleteFrom
	clauseDeleteWhere
)

// queryType deduces the type of query based on the SQL clause.
func (c sqlClause) queryType() queryType {
	switch c {
	case clauseSelectColumns, clauseSelectFrom, clauseSelectWhere, clauseSelectOrderBy:
		return querySelect
	case clauseInsertColumns, clauseInsertValues:
		return queryInsert
	case clauseUpdateTable, clauseUpdateSet, clauseUpdateWhere:
		return queryUpdate
	case clauseDeleteFrom, clauseDeleteWhere:
		return queryDelete
	}
	return queryUnknown
}

func (c sqlClause) String() string {
	switch c {
	case clauseNone:
		return "(no clause)"
	case clauseSelectColumns:
		return "select columns"
	case clauseSelectFrom:
		return "select from"
	case clauseSelectWhere:
		return "select where"
	case clauseSelectOrderBy:
		return "select order by"
	case clauseInsertColumns:
		return "insert columns"
	case clauseInsertValues:
		return "insert values"
	case clauseUpdateTable:
		return "update table"
	case clauseUpdateSet:
		return "update set"
	case clauseUpdateWhere:
		return "update where"
	case clauseDeleteFrom:
		return "delete from"
	case clauseDeleteWhere:
		return "delete where"
	}
	return fmt.Sprintf("Unknown %d", c)
}

// isInput identifies whether the SQL clause contains placeholders
// for variable input.
func (c sqlClause) isInput() bool {
	return c.matchAny(
		clauseInsertValues,
		clauseUpdateSet,
		clauseSelectWhere,
		clauseUpdateWhere,
		clauseDeleteWhere)
}

func (c sqlClause) isOutput() bool {
	return c == clauseSelectColumns
}

func (c sqlClause) acceptsColumns() bool {
	return c.isInput() ||
		c.isOutput() ||
		c.matchAny(clauseSelectOrderBy,
			clauseInsertColumns)
}

func (c sqlClause) matchAny(clauses ...sqlClause) bool {
	for _, clause := range clauses {
		if c == clause {
			return true
		}
	}
	return false
}

func (c sqlClause) defaultFilter() func(col *Column) bool {
	switch c {
	case clauseSelectWhere, clauseSelectOrderBy, clauseUpdateWhere, clauseDeleteWhere:
		return columnFilterPK
	case clauseInsertColumns, clauseInsertValues:
		return columnFilterInsertable
	case clauseUpdateSet:
		return columnFilterUpdateable
	}
	return columnFilterAll
}

// nextClause operates an extremely simple state transition
// keeping track of which part of an SQL clause we are in.
func (c sqlClause) nextClause(keyword string) sqlClause {
	keyword = strings.ToLower(keyword)

	switch keyword {
	case "delete":
		return clauseDeleteFrom
	case "from":
		switch c {
		case clauseSelectColumns:
			return clauseSelectFrom
		}
	case "insert", "into":
		return clauseInsertColumns
	case "order":
		switch c {
		case clauseSelectFrom, clauseSelectColumns, clauseSelectWhere:
			return clauseSelectOrderBy
		}
	case "select":
		return clauseSelectColumns
	case "set":
		switch c {
		case clauseUpdateTable:
			return clauseUpdateSet
		}
	case "update":
		return clauseUpdateTable
	case "values":
		switch c {
		case clauseInsertColumns:
			return clauseInsertValues
		}
	case "where":
		switch c {
		case clauseSelectFrom, clauseSelectColumns:
			return clauseSelectWhere
		case clauseDeleteFrom:
			return clauseDeleteWhere
		case clauseUpdateSet, clauseUpdateTable:
			return clauseUpdateWhere
		}
	}

	return c
}

type queryType int

const (
	queryUnknown queryType = iota
	queryInsert
	queryUpdate
	queryDelete
	querySelect
)
