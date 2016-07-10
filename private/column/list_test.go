package column_test

import (
	"testing"
	"time"

	"github.com/jjeffery/sqlf/private/colname"
	"github.com/jjeffery/sqlf/private/column"
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
		infos      []*column.Info
	}{
		{
			row: struct {
				ID   int
				Name string
			}{},
			convention: colname.Snake,
			infos: []*column.Info{
				{
					ColumnName: "id",
					Path:       "ID",
					Index:      column.NewIndex(0),
				},
				{
					ColumnName: "name",
					Path:       "Name",
					Index:      column.NewIndex(1),
				},
			},
		},
		{
			row: struct {
				ID   int `sql:",primary key"`
				Name string
			}{},
			convention: colname.Snake,
			infos: []*column.Info{
				{
					ColumnName: "id",
					Path:       "ID",
					Index:      column.NewIndex(0),
					PrimaryKey: true,
				},
				{
					ColumnName: "name",
					Path:       "Name",
					Index:      column.NewIndex(1),
				},
			},
		},
		{
			row: struct {
				ID   int `sql:",pk autoincr"`
				Name string
			}{},
			convention: colname.Snake,
			infos: []*column.Info{
				{
					ColumnName:    "id",
					Path:          "ID",
					Index:         column.NewIndex(0),
					PrimaryKey:    true,
					AutoIncrement: true,
				},
				{
					ColumnName: "name",
					Path:       "Name",
					Index:      column.NewIndex(1),
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
			infos: []*column.Info{
				{
					ColumnName:    "ID",
					Path:          "ID",
					Index:         column.NewIndex(0),
					PrimaryKey:    true,
					AutoIncrement: true,
				},
				{
					ColumnName: "Name",
					Path:       "Name",
					Index:      column.NewIndex(1),
				},
				{
					ColumnName: "AddressStreet",
					Path:       "Address.Street",
					Index:      column.NewIndex(2, 0),
				},
				{
					ColumnName: "AddressSuburb",
					Path:       "Address.Suburb",
					Index:      column.NewIndex(2, 1),
				},
				{
					ColumnName: "AddressPostcode",
					Path:       "Address.Postcode",
					Index:      column.NewIndex(2, 2),
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
			infos: []*column.Info{
				{
					ColumnName:    "id",
					Path:          "ID",
					Index:         column.NewIndex(0),
					PrimaryKey:    true,
					AutoIncrement: true,
				},
				{
					ColumnName: "address_street_number",
					Path:       "Address.Street.Number",
					Index:      column.NewIndex(2, 0, 0),
				},
				{
					ColumnName: "address_street_name",
					Path:       "Address.Street.Name",
					Index:      column.NewIndex(2, 0, 2),
				},
				{
					ColumnName: "address_suburb",
					Path:       "Address.Suburb",
					Index:      column.NewIndex(2, 1),
				},
				{
					ColumnName: "address_postcode",
					Path:       "Address.Postcode",
					Index:      column.NewIndex(2, 2),
				},
			},
		},
		{
			row: struct {
				Common
				SomeData string
			}{},
			convention: colname.Snake,
			infos: []*column.Info{
				{
					ColumnName: "id",
					Path:       "ID",
					Index:      column.NewIndex(0, 0),
					PrimaryKey: true,
				},
				{
					ColumnName: "version",
					Path:       "Version",
					Index:      column.NewIndex(0, 1),
					Version:    true,
				},
				{
					ColumnName: "updated_at",
					Path:       "UpdatedAt",
					Index:      column.NewIndex(0, 2),
				},
				{
					ColumnName: "some_data",
					Path:       "SomeData",
					Index:      column.NewIndex(1),
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
			infos: []*column.Info{
				{
					ColumnName: "yes",
					Path:       "Yes",
					Index:      column.NewIndex(0),
				},
			},
		},
	}

	for _, tt := range tests {
		infos := column.NewList(tt.row, tt.convention)
		compareInfos(t, tt.infos, infos)
	}
}

func compareInfos(t *testing.T, expected, actual []*column.Info) {
	if len(expected) != len(actual) {
		t.Errorf("expected len=%d, actual len=%d", len(expected), len(actual))
		t.FailNow()
	}
	for i, expect := range expected {
		act := actual[i]
		compareInfo(t, expect, act)
	}
}

func compareInfo(t *testing.T, info1, info2 *column.Info) {
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
