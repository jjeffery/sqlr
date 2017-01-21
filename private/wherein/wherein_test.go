package wherein

import (
	"reflect"
	"testing"
)

func TestFlatten(t *testing.T) {
	tests := []struct {
		sql      string
		args     []interface{}
		wantSQL  string
		wantArgs []interface{}
		wantErr  string
	}{
		// error conditions
		{
			sql:     "select * from tbl where id in ($0)",
			args:    []interface{}{[]int{100}},
			wantErr: "invalid placeholder $0",
		},
		{
			sql:     "select * from tbl where id in ($9999999)",
			args:    []interface{}{[]int{100}},
			wantErr: "not enough arguments for placeholder $9999999",
		},
		{
			sql:     "select * from tbl where id in (?) and name = ?",
			args:    []interface{}{[]int{100}},
			wantErr: "not enough arguments for placeholders",
		},
		{
			sql:     "select * from tbl where id in (?) and name = ?2",
			args:    []interface{}{[]int{100}, "zoe"},
			wantErr: "mix of positional and numbered placeholders",
		},

		{ // no placeholders in SQL
			sql:      "select * from tbl where id is not null",
			args:     []interface{}{[]int{100}},
			wantSQL:  "select * from tbl where id is not null",
			wantArgs: []interface{}{100},
		},
		{
			sql:      "select * from tbl where id = ?",
			args:     []interface{}{100},
			wantSQL:  "select * from tbl where id = ?",
			wantArgs: []interface{}{100},
		},
		{
			sql:      "select * from tbl where id in (?)",
			args:     []interface{}{[]int{1, 2, 3}},
			wantSQL:  "select * from tbl where id in (?,?,?)",
			wantArgs: []interface{}{1, 2, 3},
		},
		{
			sql:      "select * from tbl where id in ($1)",
			args:     []interface{}{[]int{1, 2, 3}},
			wantSQL:  "select * from tbl where id in ($1,$2,$3)",
			wantArgs: []interface{}{1, 2, 3},
		},
		{
			sql:      "select * from tbl where id in ($2) and name=$1",
			args:     []interface{}{"claire", []int{1, 2, 3}},
			wantSQL:  "select * from tbl where id in ($2,$3,$4) and name=$1",
			wantArgs: []interface{}{"claire", 1, 2, 3},
		},
	}

	for i, tt := range tests {
		gotSQL, gotArgs, gotErr := Expand(tt.sql, tt.args)
		if gotErr != nil {
			if got, want := gotErr.Error(), tt.wantErr; got != want {
				t.Errorf("%d: got=%q want=%q", i, got, want)
			}
			continue
		} else if tt.wantErr != "" {
			t.Errorf("%d: got=noerror want=%q", i, tt.wantErr)
			continue
		}

		if got, want := gotSQL, tt.wantSQL; got != want {
			t.Errorf("%d: got=%q want=%q", i, got, want)
		}

		if got, want := gotArgs, tt.wantArgs; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: got=%v want=%v", i, got, want)
		}
		t.Logf("sql: %s", gotSQL)
		t.Logf("args: %+v", gotArgs)
	}
}
