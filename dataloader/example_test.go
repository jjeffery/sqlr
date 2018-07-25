package dataloader

import (
	"fmt"
)

func ExampleMake_1() {
	// This example shows how to create a data loader that will return a thunk
	// for a row given its primary key.
	type Row struct {
		ID   int64
		Name string
	}

	type RowThunk func() (*Row, error)

	doQuery := func(ids []int64) ([]*Row, error) {
		// a real-world example would reference a database
		// here we just create rows to keep things simple
		fmt.Printf("calling query with ids = %v\n", ids)
		var rows []*Row
		for _, id := range ids {
			rows = append(rows, &Row{
				ID:   id,
				Name: fmt.Sprintf("Row #%d", id),
			})
		}
		return rows, nil
	}

	getKey := func(row *Row) int64 {
		return row.ID
	}

	var loader func(id int64) RowThunk

	Make(&loader, doQuery, getKey)

	thunk1 := loader(6)
	thunk2 := loader(32)

	for i, thunk := range []RowThunk{thunk1, thunk2} {
		v, err := thunk()
		fmt.Printf("thunk%d returns (%+v, %v)\n", i+1, v, err)
	}

	// Output:
	// calling query with ids = [6 32]
	// thunk1 returns (&{ID:6 Name:Row #6}, <nil>)
	// thunk2 returns (&{ID:32 Name:Row #32}, <nil>)
}

func ExampleMake_2() {
	// This example shows how to create a data loader whose thunk returns
	// all the Gadgets for a given Widget.

	type WidgetID int64
	type GadgetID int64

	// Widget is a widget. Many gadgets belong to a widget
	type Widget struct {
		ID   WidgetID
		Name string
	}

	// Gadget is a gadget. A gadget belongs to a widget
	type Gadget struct {
		ID       GadgetID
		Name     string
		WidgetID WidgetID
	}

	type GadgetsThunk func() ([]*Gadget, error)

	// queryFunc performs the database query
	queryFunc := func(ids []WidgetID) ([]*Gadget, error) {
		var result []*Gadget
		query := `
			select {}
			from gadgets
			where widget_id in (?)
			order by id
		`

		// performQuery is a hypothetical function that does the
		// work of running the database query and storing the results
		// in result. Package sqlr would be handy for implementing
		// performQuery.
		if err := performQuery(&result, query, ids); err != nil {
			return nil, err
		}
		return result, nil
	}

	// keyFunc is used to work out the WidgetID for each Gadget row.
	keyFunc := func(gadget *Gadget) WidgetID {
		return gadget.WidgetID
	}

	// now we can make the loader function
	var loader func(widgetID WidgetID) GadgetsThunk
	Make(&loader, queryFunc, keyFunc)

	// call the loader function a few times
	thunk1 := loader(6)
	thunk2 := loader(32)

	// call each thunk, which will result in only one query being sent
	// to the database server
	for i, thunk := range []GadgetsThunk{thunk1, thunk2} {
		v, err := thunk()
		fmt.Printf("thunk%d returns (%+v, %v)\n", i+1, v, err)
	}
}

func performQuery(rows interface{}, query string, args ...interface{}) error {
	return nil
}
