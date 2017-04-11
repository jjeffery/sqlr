package column_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/jjeffery/sqlr/private/column"
)

func TestNewList(t *testing.T) {
	type Common struct {
		ID        int64 `sql:",pk"`
		Version   int64 `sql:",version"`
		UpdatedAt time.Time
	}
	tests := []struct {
		row   interface{}
		infos []*column.Info
	}{
		{
			row: struct {
				ID   int
				Name string
			}{},
			infos: []*column.Info{
				{
					Path:  column.NewPath("ID", ""),
					Index: column.NewIndex(0),
				},
				{
					Path:  column.NewPath("Name", ""),
					Index: column.NewIndex(1),
				},
			},
		},
		{
			row: struct {
				ID   int    `sql:",primary key"`
				Name string `sql:"'primary' key"`
			}{},
			infos: []*column.Info{
				{
					Path:  column.NewPath("ID", `sql:",primary key"`),
					Index: column.NewIndex(0),
					Tag:   column.TagInfo{PrimaryKey: true},
				},
				{
					Path:  column.NewPath("Name", `sql:"'primary' key"`),
					Index: column.NewIndex(1),
				},
			},
		},
		{
			row: struct {
				ID   int    `sql:",primary key identity"`
				Name string `sql:""`
			}{},
			infos: []*column.Info{
				{
					Path:  column.NewPath("ID", `sql:",primary key identity"`),
					Index: column.NewIndex(0),
					Tag: column.TagInfo{
						PrimaryKey:    true,
						AutoIncrement: true,
					},
				},
				{
					Path:  column.NewPath("Name", `sql:""`),
					Index: column.NewIndex(1),
				},
			},
		},
		{
			row: struct {
				ID   int    `sql:"primary key auto increment"`
				Name string `sql:"[primary] key"`
			}{},
			infos: []*column.Info{
				{
					Path:  column.NewPath("ID", `sql:"primary key auto increment"`),
					Index: column.NewIndex(0),
					Tag: column.TagInfo{
						PrimaryKey:    true,
						AutoIncrement: true,
					},
				},
				{
					Path:  column.NewPath("Name", `sql:"[primary] key"`),
					Index: column.NewIndex(1),
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
			infos: []*column.Info{
				{
					Path:  column.NewPath("ID", `sql:",pk autoincr"`),
					Index: column.NewIndex(0),
					Tag: column.TagInfo{
						PrimaryKey:    true,
						AutoIncrement: true,
					},
				},
				{
					Path:  column.NewPath("Name", ""),
					Index: column.NewIndex(1),
				},
				{
					Path:  column.NewPath("Address", "").Append("Street", ""),
					Index: column.NewIndex(2, 0),
				},
				{
					Path:  column.NewPath("Address", "").Append("Suburb", ""),
					Index: column.NewIndex(2, 1),
				},
				{
					Path:  column.NewPath("Address", "").Append("Postcode", ""),
					Index: column.NewIndex(2, 2),
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
			infos: []*column.Info{
				{
					Path:  column.NewPath("ID", `sql:",pk autoincr"`),
					Index: column.NewIndex(0),
					Tag: column.TagInfo{
						PrimaryKey:    true,
						AutoIncrement: true,
					},
				},
				{
					Path: column.NewPath("Address", "").
						Append("Street", "").
						Append("Number", ""),
					Index: column.NewIndex(2, 0, 0),
				},
				{
					Path: column.NewPath("Address", "").
						Append("Street", "").
						Append("Name", ""),
					Index: column.NewIndex(2, 0, 2),
				},
				{
					Path:  column.NewPath("Address", "").Append("Suburb", ""),
					Index: column.NewIndex(2, 1),
				},
				{
					Path:  column.NewPath("Address", "").Append("Postcode", ""),
					Index: column.NewIndex(2, 2),
				},
			},
		},
		{
			row: struct {
				Common
				SomeData string
			}{},
			infos: []*column.Info{
				{
					Path:  column.NewPath("ID", `sql:",pk"`),
					Index: column.NewIndex(0, 0),
					Tag:   column.TagInfo{PrimaryKey: true},
				},
				{
					Path:  column.NewPath("Version", `sql:",version"`),
					Index: column.NewIndex(0, 1),
					Tag:   column.TagInfo{Version: true},
				},
				{
					Path:  column.NewPath("UpdatedAt", ``),
					Index: column.NewIndex(0, 2),
				},
				{
					Path:  column.NewPath("SomeData", ""),
					Index: column.NewIndex(1),
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
			infos: []*column.Info{
				{
					Path:  column.NewPath("Yes", ""),
					Index: column.NewIndex(0),
				},
			},
		},
		{
			row: struct {
				ID        int    `sql:"primary key"`
				Optional1 string `sql:"null"`
			}{},
			infos: []*column.Info{
				{
					Path:  column.NewPath("ID", `sql:"primary key"`),
					Index: column.NewIndex(0),
					Tag: column.TagInfo{
						PrimaryKey: true,
					},
				},
				{
					Path:  column.NewPath("Optional1", `sql:"null"`),
					Index: column.NewIndex(1),
					Tag: column.TagInfo{
						EmptyNull: true,
					},
				},
			},
		},
	}

	for i, tt := range tests {
		infos := column.ListForType(reflect.TypeOf(tt.row))
		compareInfos(t, i, tt.infos, infos)
	}

	// test that lists cache
	list1 := column.ListForType(reflect.TypeOf(Common{}))
	list2 := column.ListForType(reflect.TypeOf(Common{}))
	if !reflect.DeepEqual(list1, list2) {
		t.Errorf("expected list1 and list2 to have identical contents")
	}
}

func compareInfos(t *testing.T, testCase int, expected, actual []*column.Info) {
	if len(expected) != len(actual) {
		t.Errorf("%d: expected len=%d, actual len=%d", testCase, len(expected), len(actual))
		t.FailNow()
	}
	for i, expect := range expected {
		act := actual[i]
		compareInfo(t, testCase, i, expect, act)
	}
}

func compareInfo(t *testing.T, testCase int, index int, info1, info2 *column.Info) {
	if !info1.Path.Equal(info2.Path) ||
		!info1.Index.Equal(info2.Index) ||
		info1.Tag.PrimaryKey != info2.Tag.PrimaryKey ||
		info1.Tag.AutoIncrement != info2.Tag.AutoIncrement ||
		info1.Tag.EmptyNull != info2.Tag.EmptyNull ||
		info1.Tag.Version != info2.Tag.Version {
		t.Errorf("%d/%d: expected: %#v\nactual: %#v\n", testCase, index, *info1, *info2)
		t.FailNow()
	}

}
