package field_test

import (
	"testing"
	"time"

	"github.com/jjeffery/sqlf/private/colname"
	"github.com/jjeffery/sqlf/private/field"
)

func TestNewList(t *testing.T) {
	type Common struct {
		ID        int64     `sql:",pk"`
		Version   int64     `sql:",version"`
		UpdatedAt time.Time `sql:"updated_at"`
	}
	tests := []struct {
		row        interface{}
		convention colname.Convention
		infos      []*field.Info
	}{
		{
			row: struct {
				ID   int
				Name string
			}{},
			convention: colname.Snake,
			infos: []*field.Info{
				&field.Info{
					ColumnName: "id",
					Path:       "ID",
					Index:      field.NewIndex(0),
				},
				&field.Info{
					ColumnName: "name",
					Path:       "Name",
					Index:      field.NewIndex(1),
				},
			},
		},
		{
			row: struct {
				ID   int `sql:",primary key"`
				Name string
			}{},
			convention: colname.Snake,
			infos: []*field.Info{
				&field.Info{
					ColumnName: "id",
					Path:       "ID",
					Index:      field.NewIndex(0),
					PrimaryKey: true,
				},
				&field.Info{
					ColumnName: "name",
					Path:       "Name",
					Index:      field.NewIndex(1),
				},
			},
		},
		{
			row: struct {
				ID   int `sql:",pk autoincr"`
				Name string
			}{},
			convention: colname.Snake,
			infos: []*field.Info{
				&field.Info{
					ColumnName:    "id",
					Path:          "ID",
					Index:         field.NewIndex(0),
					PrimaryKey:    true,
					AutoIncrement: true,
				},
				&field.Info{
					ColumnName: "name",
					Path:       "Name",
					Index:      field.NewIndex(1),
				},
			},
		},
		{
			row: struct {
				ID      int `sql:",pk autoincr"`
				Name    string
				Address struct {
					Street   string
					Suburb   string
					Postcode string
				}
			}{},
			convention: colname.Same,
			infos: []*field.Info{
				&field.Info{
					ColumnName:    "ID",
					Path:          "ID",
					Index:         field.NewIndex(0),
					PrimaryKey:    true,
					AutoIncrement: true,
				},
				&field.Info{
					ColumnName: "Name",
					Path:       "Name",
					Index:      field.NewIndex(1),
				},
				&field.Info{
					ColumnName: "AddressStreet",
					Path:       "Address.Street",
					Index:      field.NewIndex(2, 0),
				},
				&field.Info{
					ColumnName: "AddressSuburb",
					Path:       "Address.Suburb",
					Index:      field.NewIndex(2, 1),
				},
				&field.Info{
					ColumnName: "AddressPostcode",
					Path:       "Address.Postcode",
					Index:      field.NewIndex(2, 2),
				},
			},
		},
		{
			row: struct {
				ID       int    `sql:",pk autoincr"`
				IgnoreMe string `sql:"-"`
				Address  struct {
					Street *struct {
						Number      int
						IgnoreMeToo int `sql:"-"`
						Name        string
					}
					Suburb   string
					Postcode string
				}
			}{},
			convention: colname.Snake,
			infos: []*field.Info{
				&field.Info{
					ColumnName:    "id",
					Path:          "ID",
					Index:         field.NewIndex(0),
					PrimaryKey:    true,
					AutoIncrement: true,
				},
				&field.Info{
					ColumnName: "address_street_number",
					Path:       "Address.Street.Number",
					Index:      field.NewIndex(2, 0, 0),
				},
				&field.Info{
					ColumnName: "address_street_name",
					Path:       "Address.Street.Name",
					Index:      field.NewIndex(2, 0, 2),
				},
				&field.Info{
					ColumnName: "address_suburb",
					Path:       "Address.Suburb",
					Index:      field.NewIndex(2, 1),
				},
				&field.Info{
					ColumnName: "address_postcode",
					Path:       "Address.Postcode",
					Index:      field.NewIndex(2, 2),
				},
			},
		},
		{
			row: struct {
				Common
				SomeData string
			}{},
			convention: colname.Snake,
			infos: []*field.Info{
				&field.Info{
					ColumnName: "id",
					Path:       "ID",
					Index:      field.NewIndex(0, 0),
					PrimaryKey: true,
				},
				&field.Info{
					ColumnName: "version",
					Path:       "Version",
					Index:      field.NewIndex(0, 1),
					Version:    true,
				},
				&field.Info{
					ColumnName: "updated_at",
					Path:       "UpdatedAt",
					Index:      field.NewIndex(0, 2),
				},
				&field.Info{
					ColumnName: "some_data",
					Path:       "SomeData",
					Index:      field.NewIndex(1),
				},
			},
		},
		{
			row: struct {
				Yes string
				No1 []string
				No2 chan string
				No3 map[string]string
				No4 [2]string
				No5 func(string) string
				no6 string
			}{},
			convention: colname.Snake,
			infos: []*field.Info{
				&field.Info{
					ColumnName: "yes",
					Path:       "Yes",
					Index:      field.NewIndex(0),
				},
			},
		},
	}

	for _, tt := range tests {
		infos := field.NewList(tt.row, tt.convention)
		compareInfos(t, tt.infos, infos)
	}
}

func compareInfos(t *testing.T, expected, actual []*field.Info) {
	if len(expected) != len(actual) {
		t.Errorf("expected len=%d, actual len=%d", len(expected), len(actual))
		t.FailNow()
	}
	for i, expect := range expected {
		act := actual[i]
		compareInfo(t, expect, act)
	}
}

func compareInfo(t *testing.T, info1, info2 *field.Info) {
	if info1.ColumnName != info2.ColumnName ||
		info1.Path != info2.Path ||
		!info1.Index.Equal(info2.Index) ||
		info1.PrimaryKey != info2.PrimaryKey ||
		info1.AutoIncrement != info2.AutoIncrement ||
		info1.Version != info2.Version {
		t.Errorf("expected: %#v\nactual: %#v\n", *info1, *info2)
		t.FailNow()
	}

}
