package sqlr

import (
	"fmt"
	"testing"
)

func TestSqlClauseString(t *testing.T) {
	tests := []struct {
		clause sqlClause
		text   string
	}{
		{
			clause: clauseSelectColumns,
			text:   "select columns",
		},
		{
			clause: clauseSelectFrom,
			text:   "select from",
		},
		{
			clause: clauseSelectWhere,
			text:   "select where",
		},
		{
			clause: clauseSelectOrderBy,
			text:   "select order by",
		},
		{
			clause: clauseInsertColumns,
			text:   "insert columns",
		},
		{
			clause: clauseInsertValues,
			text:   "insert values",
		},
		{
			clause: clauseUpdateTable,
			text:   "update table",
		},
		{
			clause: clauseUpdateSet,
			text:   "update set",
		},
		{
			clause: clauseUpdateWhere,
			text:   "update where",
		},
		{
			clause: clauseDeleteFrom,
			text:   "delete from",
		},
		{
			clause: clauseDeleteWhere,
			text:   "delete where",
		},
		{
			clause: sqlClause(999),
			text:   "Unknown 999",
		},
	}

	for _, tt := range tests {
		if text := tt.clause.String(); text != tt.text {
			t.Errorf("expected=%q, actual=%q", tt.text, text)
		}
		if text := fmt.Sprintf("%s", tt.clause); text != tt.text {
			t.Errorf("expected=%q, actual=%q", tt.text, text)
		}
	}
}
