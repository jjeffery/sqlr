package scan

import (
	"strings"
	"testing"
)

func TestScan(t *testing.T) {
	type tokenLit struct {
		token Token
		lit   string
	}
	testCases := []struct {
		sql    string
		tokens []tokenLit
	}{
		{
			sql: "select * from [tbl] t where t.id = 'one'",
			tokens: []tokenLit{
				{IDENT, "select"},
				{WS, " "},
				{OP, "*"},
				{WS, " "},
				{IDENT, "from"},
				{WS, " "},
				{IDENT, "[tbl]"},
				{WS, " "},
				{IDENT, "t"},
				{WS, " "},
				{IDENT, "where"},
				{WS, " "},
				{IDENT, "t"},
				{OP, "."},
				{IDENT, "id"},
				{WS, " "},
				{OP, "="},
				{WS, " "},
				{LITERAL, "'one'"},
				{EOF, ""},
			},
		},
		{
			sql: "[table_]]name]]] `column_[]_n``ame`, \"another \"\"name\"\"\"",
			tokens: []tokenLit{
				{IDENT, "[table_]]name]]]"},
				{WS, " "},
				{IDENT, "`column_[]_n``ame`"},
				{OP, ","},
				{WS, " "},
				{IDENT, "\"another \"\"name\"\"\""},
				{EOF, ""},
			},
		},
	}

	for i, tc := range testCases {
		scanner := NewScanner(strings.NewReader(tc.sql))
		for j, expected := range tc.tokens {
			tok, lit := scanner.Scan()
			if tok != expected.token || lit != expected.lit {
				t.Errorf("%d,%d: expected (%v,%s), got (%v,%s)", i, j, expected.token, expected.lit, tok, lit)
			}
		}
	}
}
