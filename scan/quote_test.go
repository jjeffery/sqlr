package scan

import (
	"testing"
)

func TestQuote(t *testing.T) {
	testCases := []struct {
		ident    string
		isQuoted bool
		unQuoted string
		quoted   []string
	}{
		{
			ident:    `"identifier"`,
			isQuoted: true,
			unQuoted: "identifier",
			quoted: []string{
				`"identifier"`,
				"`identifier`",
				"[identifier]",
				"'identifier'",
			},
		},
		{
			ident:    `"id""1"`,
			isQuoted: true,
			unQuoted: `id"1`,
			quoted: []string{
				`"id""1"`,
				"`id\"1`",
				`[id"1]`,
				`'id"1'`,
			},
		},
		{
			ident:    `"id""2"""`,
			isQuoted: true,
			unQuoted: `id"2"`,
			quoted: []string{
				`"id""2"""`,
				"`id\"2\"`",
				`[id"2"]`,
				`'id"2"'`,
			},
		},
		{
			ident:    "`table ``name```",
			isQuoted: true,
			unQuoted: "table `name`",
			quoted: []string{
                "\"table `name`\"",
				"`table ``name```",
				"[table `name`]",
				"'table `name`'",
			},
		},
        {
          ident: "some_identifier",
          isQuoted: false,
          unQuoted: "some_identifier",
			quoted: []string{
                `"some_identifier"`,
				"`some_identifier`",
				"[some_identifier]",
				"'some_identifier'",
			},
        },
	}

	for i, tc := range testCases {
		isQuoted := IsQuoted(tc.ident)
		if isQuoted != tc.isQuoted {
			t.Errorf("%d: isQuoted: expected=%v, actual=%v", i, tc.isQuoted, isQuoted)
		}
		unQuoted := Unquote(tc.ident)
		if unQuoted != tc.unQuoted {
			t.Errorf("%d: unQuoted: expected=%s, actual=%s", i, tc.unQuoted, unQuoted)
			continue
		}
		for _, q := range tc.quoted {
			start := q[:1]
			end := q[len(q)-1:]
			quoted := Quote(tc.ident, start, end)
			if quoted != q {
				t.Errorf("%d: quoted: expected=%s, actual=%s", i, q, quoted)
			}
		}
	}
}
