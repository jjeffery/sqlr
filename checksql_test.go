package sqlr

import "testing"

func TestCheckSQL(t *testing.T) {
	tests := []struct {
		in      string
		out     string
		errText string
	}{
		{
			in:  "insert into table_name",
			out: "insert into table_name({}) values({})",
		},
		{
			in:  "Insert Into [TableName]",
			out: "insert into [TableName]({}) values({})",
		},
		{
			in:  `UPDATE "Table Name"`,
			out: `update "Table Name" set {} where {}`,
		},
		{
			in:  `select tblname`,
			out: `select {} from tblname where {}`,
		},
		{
			in:  "  select     from\ttblname ",
			out: `select {} from tblname where {}`,
		},
		{
			in:      "delete from my_table",
			errText: `will not delete all rows in table my_table: use database/sql if you want to do this`,
		},
		{
			in:      "delete  [my table]",
			errText: `will not delete all rows in table [my table]: use database/sql if you want to do this`,
		},
	}

	for i, tt := range tests {
		sql, err := checkSQL(tt.in)
		if got, want := sql, tt.out; got != want {
			t.Errorf("%d: want=%q, got=%q", i, want, got)
		}
		var errText string
		if err != nil {
			errText = err.Error()
		}
		if got, want := errText, tt.errText; got != want {
			t.Errorf("%d: want=%q, got=%q", i, want, got)
		}
	}
}
