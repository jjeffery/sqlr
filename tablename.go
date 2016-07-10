package sqlf

import (
	"fmt"
)

// TableName represents the name of a table for formatting
// in an SQL statement. The format will depend on where the
// table appears in the SQL statement. For example, in a SELECT FROM
// clause, the table may include an alias, but in an INSERT INTO statement
// the table will not have an alias. (TODO: INSERT x INTO x ... FROM x, y, etc)
type TableName struct {
	table  *TableInfo
	clause sqlClause
}

func newTableName(table *TableInfo, clause sqlClause) TableName {
	return TableName{
		table:  table,
		clause: clause,
	}
}

// String prints the table name in the appropriate
// form for the part of the SQL statement that this TableName
// applies to. Because TableName implements the Stringer
// interface, it can be formatted using "%s" in fmt.Sprintf.
func (tn TableName) String() string {
	dialect := tn.table.dialect
	switch tn.clause {
	case clauseSelectFrom:
		if tn.table.alias != "" {
			return fmt.Sprintf("%s as %s",
				dialect.Quote(tn.table.name),
				dialect.Quote(tn.table.alias),
			)
		}
		return dialect.Quote(tn.table.name)
	case clauseInsertInto, clauseUpdateTable, clauseDeleteTable:
		return dialect.Quote(tn.table.name)
	}
	panic(fmt.Sprintf("invalid clause for table name: %d", tn.clause))
}
