package sqlr

import (
	"testing"
)

func TestWithNamingConvention(t *testing.T) {
	tests := []struct {
		schema      *Schema
		row         interface{}
		columnNames []string
	}{
		{
			schema: NewSchema(
				WithNamingConvention(SnakeCase),
				WithField("ID", "RowID"),
			),
			row: struct {
				ID       int `sql:"row_id"`
				FullName string
			}{},
			columnNames: []string{
				"RowID",
				"full_name",
			},
		},
		{
			schema: NewSchema(
				WithNamingConvention(SnakeCase),
			),
			row: struct {
				ID       int `sql:"row_id"`
				FullName string
			}{},
			columnNames: []string{
				"row_id",
				"full_name",
			},
		},
		{
			schema: NewSchema(
				WithNamingConvention(SameCase),
			),
			row: struct {
				ID       int    `sql:"Id"`
				FullName string `sql:"Full_Name"`
			}{},
			columnNames: []string{
				"Id",
				"Full_Name",
			},
		},
		{
			schema: NewSchema(
				WithNamingConvention(SameCase),
			),
			row: struct {
				ID            int    `sql:"Id"`
				FullName      string `sql:"Full_Name"`
				SomethingElse float32
			}{},
			columnNames: []string{
				"Id",
				"Full_Name",
				"SomethingElse",
			},
		},
		{
			schema: NewSchema(
				WithNamingConvention(LowerCase),
				WithField("Home.Locality", "suburb"),
				WithField("ID", "rowid"),
			),
			row: struct {
				ID   int
				Home struct {
					Street   string
					Locality string
				}
			}{},
			columnNames: []string{
				"rowid",
				"homestreet",
				"suburb",
			},
		},
	}

	for i, tt := range tests {
		rowType, err := getRowType(tt.row)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}
		tbl := tt.schema.TableFor(rowType)
		for j, col := range tbl.Columns() {
			if got, want := col.Name(), tt.columnNames[j]; got != want {
				t.Errorf("%d: %d: want=%q, got=%q", i, j, want, got)
			}
		}
	}
}

func TestWithKey(t *testing.T) {
	s := NewSchema()
	if got, want := s.Key(), ""; got != want {
		t.Errorf("got=%q want=%q", got, want)
	}
	s = NewSchema(WithKey("xxx"))
	if got, want := s.Key(), "xxx"; got != want {
		t.Errorf("got=%q want=%q", got, want)
	}
}
