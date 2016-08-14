package sqlrow

import (
	"database/sql"
	"testing"
)

func TestInvalidStmts(t *testing.T) {
	type Row struct {
		ID     int64 `sql:"primary key"`
		Name   string
		Number int
	}

	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	var row Row
	var notRow int

	tests := []struct {
		fn   func() (int, error)
		want string
	}{
		{
			fn:   func() (int, error) { return 0, Insert(db, &notRow, "rows") },
			want: `expected arg for "row" to refer to a struct type`,
		},
		{
			fn:   func() (int, error) { return 0, Insert(db, &row, "insert into xyz values({})") },
			want: `cannot expand "insert values" clause because "insert columns" clause is missing`,
		},
		{
			fn:   func() (int, error) { return 0, Insert(db, &row, "insert into xyz({}) values({pk})") },
			want: `columns for "insert values" clause must match the "insert columns" clause`,
		},
		{
			fn:   func() (int, error) { return Update(db, &row, "update {} this is not valid SQL") },
			want: `cannot expand "{}" in "update table" clause`,
		},
		{
			fn:   func() (int, error) { return Update(db, &row, "update rows set {} where {} and number=?") },
			want: `expected arg count=1, actual=0`,
		},
		{
			fn:   func() (int, error) { return Delete(db, &row, "rows") },
			want: `no such table: rows`,
		},
		{
			fn:   func() (int, error) { return Select(db, &row, "select {alias} from rows") },
			want: `cannot expand "alias" in "select columns" clause: missing ident after 'alias'`,
		},
		{
			fn:   func() (int, error) { return Select(db, &row, "select {'col1} from rows") },
			want: `cannot expand "'col1" in "select columns" clause: unrecognised input near "'col1"`,
		},
		{
			fn:   func() (int, error) { return Select(db, &notRow, "select {} from rows") },
			want: `expected arg for "rows" to refer to a struct type`,
		},
	}

	for i, tt := range tests {
		_, err := tt.fn()
		if err != nil {
			if tt.want != err.Error() {
				t.Errorf("%d: want %s, got %v", i, tt.want, err.Error())
			}
			continue
		}
		t.Errorf("%d: want %s, got nil", i, tt.want)
	}
}

func TestInvalidPrepare(t *testing.T) {
	var notRow []int
	_, err := Prepare(notRow, "select {} from rows")
	want := `expected arg for "row" to refer to a struct type`
	if err != nil {
		if want != err.Error() {
			t.Errorf("want %s, got %v", want, err)
		}
	} else {
		t.Errorf("want %s, got nil", want)
	}
}
