package sqlr

import "testing"

func TestJSONCell(t *testing.T) {
	{
		var row struct {
			V1 int
			V2 string
		}
		nc := newJSONCell("col", &row)
		nc.data = []byte(`{"V1":1,"V2":"2"}`)
		if err := nc.Unmarshal(); err != nil {
			t.Error(err)
		}
		if got, want := row.V1, 1; got != want {
			t.Errorf("want=%v, got=%v", want, got)
		}
		if got, want := row.V2, "2"; got != want {
			t.Errorf("want=%v, got=%v", want, got)
		}
	}
	{
		var row struct {
			V1 int
			V2 string
		}
		nc := newJSONCell("col", &row)
		nc.data = nil
		if err := nc.Unmarshal(); err != nil {
			t.Error(err)
		}
		if got, want := row.V1, 0; got != want {
			t.Errorf("want=%v, got=%v", want, got)
		}
		if got, want := row.V2, ""; got != want {
			t.Errorf("want=%v, got=%v", want, got)
		}
	}
	{
		var row struct {
			V1 int
			V2 string
		}
		nc := newJSONCell("col", &row)
		nc.data = []byte(`{"V1":1,"V2":`)
		err := nc.Unmarshal()
		if err == nil {
			t.Error("expected error, got none")
		}
		if got, want := err.Error(), `cannot unmarshal JSON field "col": unexpected end of JSON input`; got != want {
			t.Errorf("want=%v, got=%v", want, got)
		}
	}
}
