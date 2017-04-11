package sqlrow

import (
	"testing"

	"github.com/jjeffery/sqlrow/private/column"
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
			).Clone(
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
		rowType, err := inferRowType(tt.row)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}
		cols := column.ListForType(rowType)
		columnNamer := tt.schema.columnNamer()
		for j, col := range cols {
			if got, want := columnNamer.ColumnName(col), tt.columnNames[j]; got != want {
				t.Errorf("%d: %d: want=%q, got=%q", i, j, want, got)
			}
		}

	}
}
