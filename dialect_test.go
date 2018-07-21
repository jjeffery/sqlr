package sqlr

import (
	"testing"
)

func TestDialect(t *testing.T) {
	tests := []struct {
		dialect     Dialect
		quoted      string
		placeholder string
	}{
		{
			dialect:     MySQL,
			quoted:      "`quoted`",
			placeholder: "?",
		},
		{
			dialect:     Postgres,
			quoted:      `"quoted"`,
			placeholder: "$1",
		},
	}

	for _, tt := range tests {
		dialect := tt.dialect
		quoted := dialect.Quote("quoted")
		placeholder := dialect.Placeholder(1)
		if quoted != tt.quoted {
			t.Errorf("expected=%q, actual=%q", tt.quoted, quoted)
		}
		if placeholder != tt.placeholder {
			t.Errorf("expected=%q, actual=%q", tt.placeholder, placeholder)
		}
	}
}

func TestDialectFor(t *testing.T) {
	if got, want := dialectFor(nil), DefaultDialect(); got != want {
		t.Errorf("want=%v, got=%v", want, got)
	}
}

func TestIsPostgres(t *testing.T) {
	tests := []struct {
		d          Dialect
		isPostgres bool
	}{
		{
			d:          Postgres,
			isPostgres: true,
		},
		{
			d: MySQL,
		},
		{
			d: SQLite,
		},
		{
			d: MSSQL,
		},
	}

	for i, tt := range tests {
		if got, want := isPostgres(tt.d), tt.isPostgres; got != want {
			t.Logf("%d: got=%v, want=%v", i, got, want)
		}
	}
}
