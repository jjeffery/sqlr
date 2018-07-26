package wherein

import (
	"reflect"
	"testing"
)

func TestFlatten(t *testing.T) {
	type intType int
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
			sql:      "select * from tbl where id in (?)",
			args:     []interface{}{[]intType{1, 2, 3}},
			wantSQL:  "select * from tbl where id in (?,?,?)",
			wantArgs: []interface{}{intType(1), intType(2), intType(3)},
		},
		{
			sql:      "select * from tbl where id in ($1)",
			args:     []interface{}{[]int{1, 2, 3}},
			wantSQL:  "select * from tbl where id in ($1,$2,$3)",
			wantArgs: []interface{}{1, 2, 3},
		},
		{
			sql:      "select * from tbl where id in ($2) and name=$1",
			args:     []interface{}{[]byte("claire"), []int{1, 2, 3}},
			wantSQL:  "select * from tbl where id in ($2,$3,$4) and name=$1",
			wantArgs: []interface{}{[]byte("claire"), 1, 2, 3},
		},
		{
			sql:      "select * from tbl where id in ($2) and name in ($1)",
			args:     []interface{}{[]string{"zoe", "michaela", "nick", "claire"}, []int{1, 2, 3}},
			wantSQL:  "select * from tbl where id in ($5,$6,$7) and name in ($1,$2,$3,$4)",
			wantArgs: []interface{}{"zoe", "michaela", "nick", "claire", 1, 2, 3},
		},
		{
			sql:      "select * from tbl where age > $1 and id in ($3) and name in ($2)",
			args:     []interface{}{16, []string{"zoe", "michaela", "nick", "claire"}, []int{1, 2, 3}},
			wantSQL:  "select * from tbl where age > $1 and id in ($6,$7,$8) and name in ($2,$3,$4,$5)",
			wantArgs: []interface{}{16, "zoe", "michaela", "nick", "claire", 1, 2, 3},
		},
		{
			sql:      "select * from tbl where age > ? and id in (?) and name in (?)",
			args:     []interface{}{16, []string{"zoe", "michaela", "nick", "claire"}, []int{1, 2, 3}},
			wantSQL:  "select * from tbl where age > ? and id in (?,?,?,?) and name in (?,?,?)",
			wantArgs: []interface{}{16, "zoe", "michaela", "nick", "claire", 1, 2, 3},
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
