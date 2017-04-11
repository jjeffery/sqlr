package sqlrow

import "testing"

func TestCheckSQL(t *testing.T) {
	tests := []struct {
		in  string
		out string
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
	}

	for i, tt := range tests {
		if got, want := checkSQL(tt.in), tt.out; got != want {
			t.Errorf("%d: want=%q, got=%q", i, want, got)
		}
	}
}
