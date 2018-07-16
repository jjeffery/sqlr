package dataloader

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestLoader(t *testing.T) {
	type Row struct {
		ID   int
		Name string
	}

	type RowThunk func() (*Row, error)

	queryFunc := func(ids []int) ([]*Row, error) {
		var result []*Row

		for _, id := range ids {
			result = append(result, &Row{
				ID:   id,
				Name: fmt.Sprintf("ID %d", id),
			})
		}

		return result, nil
	}

	keyFunc := func(t *Row) int {
		return t.ID
	}

	var loader func(id int) RowThunk

	Make(&loader, queryFunc, keyFunc)

	var thunks []RowThunk

	for _, id := range []int{18, 55, 82} {
		thunks = append(thunks, loader(id))
	}

	for i, want := range []string{"ID 18", "ID 55", "ID 82"} {
		got, err := thunks[i]()
		if err != nil {
			t.Errorf("got err=%v, want err=nil", err)
			continue
		}
		if got.Name != want {
			t.Errorf("got %v, want %v", got, want)
			continue
		}
	}
}

func TestLoaderMultiples(t *testing.T) {
	type Row struct {
		ID      int
		Name    string
		OtherID int
	}

	type RowsThunk func() ([]*Row, error)

	queryFunc := func(otherIDs []int) ([]*Row, error) {
		var result []*Row

		for _, otherID := range otherIDs {
			for id := otherID * 10; id < otherID*10+3; id++ {
				result = append(result, &Row{
					ID:      id,
					Name:    fmt.Sprintf("ID %d", id),
					OtherID: otherID,
				})
			}
		}

		return result, nil
	}

	keyFunc := func(t *Row) int {
		return t.OtherID
	}

	var loader func(otherID int) RowsThunk

	Make(&loader, queryFunc, keyFunc)

	var thunks []RowsThunk
	otherIDs := []int{23, 31, 47}

	for _, otherID := range otherIDs {
		thunks = append(thunks, loader(otherID))
	}

	for i, otherID := range otherIDs {
		got, err := thunks[i]()
		if err != nil {
			t.Errorf("want no error, got=%v", err)
			continue
		}
		for j := 0; j < 3; j++ {
			want := fmt.Sprintf("ID %d", otherID*10+j)
			if got[j].Name != want {
				t.Errorf("%d-%d: want=%v, got=%v", i, j, want, got)
			}
		}
	}
}

func TestKnownTypes(t *testing.T) {
	if got, want := knownTypes.nilErrorValue.IsNil(), true; got != want {
		t.Errorf("got=%v, want=%v", got, want)
	}
	if got := knownTypes.nilErrorValue.Interface(); got != nil {
		t.Errorf("got=%v, want=%v", got, nil)
	}
	if errType := reflect.TypeOf(errors.New("test")); !errType.Implements(knownTypes.errorType) {
		t.Errorf("type %v does not implement %v", errType, knownTypes.errorType)
	}
}
