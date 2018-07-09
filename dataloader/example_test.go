package dataloader

import (
	"fmt"
)

func ExampleMake() {
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
