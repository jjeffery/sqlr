package sqlr

import "testing"

func TestTableFor(t *testing.T) {
	type Row struct {
		ID      string `sql:"pk"`
		Version int64  `sql:"version"`
		Name    string `sql:"natural key"`
	}

	schema := NewSchema()
	tbl := schema.TableFor(&Row{})

	if got, want := tbl.Name(), "row"; got != want {
		t.Errorf("got=%v, want=%v", got, want)
	}
	if got, want := len(tbl.Columns()), 3; got != want {
		t.Errorf("got=%v, want=%v", got, want)
	}
	if got, want := len(tbl.PrimaryKey()), 1; got != want {
		t.Errorf("got=%v, want=%v", got, want)
	}
	if got, want := len(tbl.NaturalKey()), 1; got != want {
		t.Errorf("got=%v, want=%v", got, want)
	}
}
