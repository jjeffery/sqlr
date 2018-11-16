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
	operatorList := func(ops ...string) []tokenLit {
		var list []tokenLit
		list = append(list, tokenLit{WS, " "})
		for _, op := range ops {
			list = append(list, tokenLit{OP, op})
			list = append(list, tokenLit{WS, " "})
		}
		return list
	}
	testCases := []struct {
		sql                    string
		tokens                 []tokenLit
		ignoreWhiteSpaceTokens []tokenLit
		errText                string
	}{
		{
			sql: `select * from tblname where some_column @@ to_tsquery('some & text')`,
			tokens: []tokenLit{
				{KEYWORD, "select"},
				{WS, " "},
				{OP, "*"},
				{WS, " "},
				{KEYWORD, "from"},
				{WS, " "},
				{IDENT, "tblname"},
				{WS, " "},
				{IDENT, "where"},
				{WS, " "},
				{IDENT, "some_column"},
				{WS, " "},
				{OP, "@@"},
				{WS, " "},
				{IDENT, "to_tsquery"},
				{OP, "("},
				{LITERAL, "'some & text'"},
				{OP, ")"},
				{EOF, ""},
			},
		},
		{
			sql: "select * from [from] t where t.id = 'one'",
			tokens: []tokenLit{
				{KEYWORD, "select"},
				{WS, " "},
				{OP, "*"},
				{WS, " "},
				{KEYWORD, "from"},
				{WS, " "},
				{IDENT, "[from]"},
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
			ignoreWhiteSpaceTokens: []tokenLit{
				{KEYWORD, "select"},
				{OP, "*"},
				{KEYWORD, "from"},
				{IDENT, "[from]"},
				{IDENT, "t"},
				{IDENT, "where"},
				{IDENT, "t"},
				{OP, "."},
				{IDENT, "id"},
				{OP, "="},
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
			errText: `unrecognised input near "[table_]]name]]"`,
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
				{PLACEHOLDER, "$"},
				{WS, " "},
				{PLACEHOLDER, "?"},
				{EOF, ""},
			},
		},
		{ // comments
			sql: "select -- this is a comment\n5-2-- another comment",
			tokens: []tokenLit{
				{KEYWORD, "select"},
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
			errText: `unrecognised input near "'missing quote"`,
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
		{ // illegal token
			sql: "\x03",
			tokens: []tokenLit{
				{ILLEGAL, "\x03"},
				{EOF, ""},
			},
			errText: `unrecognised input near "\x03"`,
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
		// placeholder
		{
			sql: "select * from [tbl] t where t.id = ? and t.version = ?",
			tokens: []tokenLit{
				{KEYWORD, "select"},
				{WS, " "},
				{OP, "*"},
				{WS, " "},
				{KEYWORD, "from"},
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
		// placeholder
		{
			sql: "select {whatever} from [tbl]",
			tokens: []tokenLit{
				{KEYWORD, "select"},
				{WS, " "},
				{IDENT, "{whatever}"},
				{WS, " "},
				{KEYWORD, "from"},
				{WS, " "},
				{IDENT, "[tbl]"},
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
		{ // postgres comparison operators
			sql:    ` < > <= >= = <> != `,
			tokens: operatorList("<", ">", "<=", ">=", "=", "<>", "!="),
		},
		{ // postgres mathematical operators
			sql:    ` + - * / % ^ |/ ||/ ! !! @ & | # ~ << >> `,
			tokens: operatorList("+", "-", "*", "/", "%", "^", "|/", "||/", "!", "!!", "@", "&", "|", "#", "~", "<<", ">>"),
		},
		{ // postgres bit string
			sql: `B'10001' || B'011'`,
			tokens: []tokenLit{
				{LITERAL, "B'10001'"},
				{WS, " "},
				{OP, "||"},
				{WS, " "},
				{LITERAL, "B'011'"},
			},
		},
		{ // postgres geometric operators
			sql: ` @-@ @@ ## <-> && &< &> <<| |>> &<| |&> <^ >^ ?- ?# ?| ?-| ?|| @> <@ ~= `,
			tokens: operatorList(
				"@-@", "@@", "##", "<->", "&&", "&<", "&>", "<<|", "|>>", "&<|", "|&>",
				"<^", ">^", "?-", "?#", "?|", "?-|", "?||", "@>", "<@", "~=",
			),
		},
		{ // postgres network comparison
			sql:    ` <<= >>= ~ `,
			tokens: operatorList("<<=", ">>=", "~"),
		},
		{ // postgres text search
			sql:    ` @@ @@@ || && !! @> <@ `,
			tokens: operatorList("@@", "@@@", "||", "&&", "!!", "@>", "<@"),
		},
		{ // postgres json operators
			sql:    ` -> ->> #> #>> `,
			tokens: operatorList("->", "->>", "#>", "#>>"),
		},
		{ // range operators (that have not already been tested elsewhere
			sql:    ` -|- `,
			tokens: operatorList("-|-"),
		},
	}

	check := func(scan *Scanner, tokens []tokenLit, sql string, errText string) {
		if len(tokens) == 0 {
			return
		}
		for i, expected := range tokens {
			if !scan.Scan() {
				if scan.Token() != EOF && scan.Token() != ILLEGAL {
					t.Errorf("%d: premature end of input: tok=%v, lit=%q, sql=%q",
						i, scan.Token(), scan.Text(), sql)
				}
				continue
			}
			tok, lit := scan.Token(), scan.Text()
			if tok != expected.token || lit != expected.lit {
				t.Errorf("%d: %q, expected (%v,%s), got (%v,%s)",
					i, sql, expected.token, expected.lit, tok, lit)
			}
		}
		if errText == "" {
			if scan.Err() != nil {
				t.Errorf("expected no error, actual=%v", scan.Err())
			}
		} else {
			if scan.Err() == nil {
				t.Errorf("expected error %q, actual=nil", errText)
			} else if scan.Err().Error() != errText {
				t.Errorf("expected error %q, actual=%v", errText, scan.Err())
			}
		}
	}

	for _, tc := range testCases {
		scanner := New(strings.NewReader(tc.sql))
		scanner.AddKeywords("select", "from")
		check(scanner, tc.tokens, tc.sql, tc.errText)
		scanner = New(strings.NewReader(tc.sql))
		scanner.AddKeywords("select", "from")
		scanner.IgnoreWhiteSpace = true
		check(scanner, tc.ignoreWhiteSpaceTokens, tc.sql, tc.errText)
	}
}
