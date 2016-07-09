package scanner

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
		{ // identifiers
			sql: "a _aa BB xyz_abc-d",
			tokens: []tokenLit{
				{IDENT, "a"},
				{WS, " "},
				{IDENT, "_aa"},
				{WS, " "},
				{IDENT, "BB"},
				{WS, " "},
				{IDENT, "xyz_abc"},
				{OP, "-"},
				{IDENT, "d"},
				{EOF, ""},
			},
		},
		{ // delimited identifiers
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
		{ // unfinished delimited identifier
			sql: "[table_]]name]]",
			tokens: []tokenLit{
				{ILLEGAL, "[table_]]name]]"},
				{EOF, ""},
			},
		},
		{ // placeholders
			sql: "? ?123 $4 $ ?",
			tokens: []tokenLit{
				{PLACEHOLDER, "?"},
				{WS, " "},
				{PLACEHOLDER, "?123"},
				{WS, " "},
				{PLACEHOLDER, "$4"},
				{WS, " "},
				{OP, "$"},
				{WS, " "},
				{PLACEHOLDER, "?"},
				{EOF, ""},
			},
		},
		{ // comments
			sql: "select -- this is a comment\n5-2-- another comment",
			tokens: []tokenLit{
				{IDENT, "select"},
				{WS, " "},
				{COMMENT, "-- this is a comment\n"},
				{LITERAL, "5"},
				{OP, "-"},
				{LITERAL, "2"},
				{COMMENT, "-- another comment"},
				{EOF, ""},
			},
		},
		{ // literals
			sql: "'literal ''string''',x'1010',X'1010',n'abc',N'abc',xy,X,nm,N,",
			tokens: []tokenLit{
				{LITERAL, "'literal ''string'''"},
				{OP, ","},
				{LITERAL, "x'1010'"},
				{OP, ","},
				{LITERAL, "X'1010'"},
				{OP, ","},
				{LITERAL, "n'abc'"},
				{OP, ","},
				{LITERAL, "N'abc'"},
				{OP, ","},
				{IDENT, "xy"},
				{OP, ","},
				{IDENT, "X"},
				{OP, ","},
				{IDENT, "nm"},
				{OP, ","},
				{IDENT, "N"},
				{OP, ","},
				{EOF, ""},
			},
		},
		{ // illegal quoted literal
			sql: "'missing quote",
			tokens: []tokenLit{
				{ILLEGAL, "'missing quote"},
				{EOF, ""},
			},
		},
		{ // numbers
			sql: "123,123.456,.123,5",
			tokens: []tokenLit{
				{LITERAL, "123"},
				{OP, ","},
				{LITERAL, "123.456"},
				{OP, ","},
				{LITERAL, ".123"},
				{OP, ","},
				{LITERAL, "5"},
				{EOF, ""},
			},
		},
		{ // not-equals, gt, lt operators
			sql: "<<>>",
			tokens: []tokenLit{
				{OP, "<"},
				{OP, "<>"},
				{OP, ">"},
				{EOF, ""},
			},
		},
		{ // illegal token
			sql: "\x03",
			tokens: []tokenLit{
				{ILLEGAL, "\x03"},
				{EOF, ""},
			},
		},
		{ // white space
			sql: " a  b\r\nc\td \v\t\r\n  e\n\n",
			tokens: []tokenLit{
				{WS, " "},
				{IDENT, "a"},
				{WS, "  "},
				{IDENT, "b"},
				{WS, "\r\n"},
				{IDENT, "c"},
				{WS, "\t"},
				{IDENT, "d"},
				{WS, " \v\t\r\n  "},
				{IDENT, "e"},
				{WS, "\n\n"},
				{EOF, ""},
			},
		},
		// placehoder
		{
			sql: "select * from [tbl] t where t.id = ? and t.version = ?",
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
				{PLACEHOLDER, "?"},
				{WS, " "},
				{IDENT, "and"},
				{WS, " "},
				{IDENT, "t"},
				{OP, "."},
				{IDENT, "version"},
				{WS, " "},
				{OP, "="},
				{WS, " "},
				{PLACEHOLDER, "?"},
				{EOF, ""},
			},
		},
		{
			sql: "id=$1 and version=$2",
			tokens: []tokenLit{
				{IDENT, "id"},
				{OP, "="},
				{PLACEHOLDER, "$1"},
				{WS, " "},
				{IDENT, "and"},
				{WS, " "},
				{IDENT, "version"},
				{OP, "="},
				{PLACEHOLDER, "$2"},
				{EOF, ""},
			},
		},
		{
			sql: "$1",
			tokens: []tokenLit{
				{PLACEHOLDER, "$1"},
				{EOF, ""},
			},
		},
	}

	for i, tc := range testCases {
		scanner := New(strings.NewReader(tc.sql))
		for j, expected := range tc.tokens {
			tok, lit := scanner.Scan()
			if tok != expected.token || lit != expected.lit {
				t.Errorf("%d,%d: expected (%v,%s), got (%v,%s)", i, j, expected.token, expected.lit, tok, lit)
			}
		}
	}
}
