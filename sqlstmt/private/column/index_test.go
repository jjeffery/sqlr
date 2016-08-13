package column

import (
	"reflect"
	"testing"
)

func TestIndexEqual(t *testing.T) {
	tests := []struct {
		ix1   Index
		ix2   Index
		equal bool
	}{
		{
			ix1:   nil,
			ix2:   NewIndex(),
			equal: true,
		},
		{
			ix1:   nil,
			ix2:   nil,
			equal: true,
		},
		{
			ix1:   NewIndex(),
			ix2:   NewIndex(),
			equal: true,
		},
		{
			ix1:   NewIndex(0),
			ix2:   NewIndex(0),
			equal: true,
		},
		{
			ix1:   NewIndex(0, 1),
			ix2:   NewIndex(0, 1),
			equal: true,
		},
		{
			ix1:   nil,
			ix2:   NewIndex(0, 1),
			equal: false,
		},
		{
			ix1:   NewIndex(1, 0),
			ix2:   NewIndex(0, 1),
			equal: false,
		},
	}

	for _, tt := range tests {
		equal := tt.ix1.Equal(tt.ix2)
		if tt.equal != equal {
			t.Errorf("expected %v, actual %v", tt.equal, equal)
		}
		equal = tt.ix2.Equal(tt.ix1)
		if tt.equal != equal {
			t.Errorf("expected %v, actual %v", tt.equal, equal)
		}
	}
}

func TestValueRO(t *testing.T) {
	tests := []struct {
		row   interface{}
		index Index
		value interface{}
	}{
		{
			row: struct {
				A int
			}{A: 4},
			index: NewIndex(0),
			value: 4,
		},
		{
			row: struct {
				A int
				B struct {
					C string
				}
			}{B: struct{ C string }{C: "xyz"}},
			index: NewIndex(1, 0),
			value: "xyz",
		},
	}

	for _, tt := range tests {
		rv := reflect.ValueOf(tt.row)
		actualValue := tt.index.ValueRO(rv)
		expectedValue := reflect.ValueOf(tt.value)
		if reflect.DeepEqual(expectedValue, actualValue) {
			t.Errorf("expected=%#v, actual=%#v", expectedValue, actualValue)
		}
	}
}

func TestValueRW(t *testing.T) {
	tests := []struct {
		row   interface{}
		index Index
		value interface{}
	}{
		{
			row: &struct {
				A *int
			}{},
			index: NewIndex(0),
			value: 0,
		},
		{
			row: &struct {
				A int
				B *struct {
					C *string
				}
			}{},
			index: NewIndex(1, 0),
			value: "",
		},
		{
			row: &struct {
				A []int
			}{},
			index: NewIndex(0),
			value: []int{},
		},
		{
			row: &struct {
				A map[string]int
			}{},
			index: NewIndex(0),
			value: map[string]int{},
		},
	}

	for _, tt := range tests {
		rv := reflect.ValueOf(tt.row)
		actualValue := tt.index.ValueRW(rv)
		expectedValue := reflect.ValueOf(tt.value)
		if reflect.DeepEqual(expectedValue, actualValue) {
			t.Errorf("expected=%#v, actual=%#v", expectedValue, actualValue)
		}
	}
}
