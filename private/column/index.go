package column

import (
	"reflect"
)

// Index is used to efficiently find the value for a database column
// in the associated field within a structure.
// In most cases an index is a single integer, which
// represents the index of the relevant field in the structure. In the
// case of fields in embedded structs, a field index consists of more than
// one integer.
type Index []int

// NewIndex returns an index with the specified values.
func NewIndex(vals ...int) Index {
	return Index(vals)
}

// Append a number to an existing index to create
// a new index. The original index ix is unchanged.
//
// If ix is nil, then Append returns an index
// with a single index value.
func (ix Index) Append(index int) Index {
	clone := ix.Clone()
	return append(clone, index)
}

// Clone creates a deep copy of ix.
func (ix Index) Clone() Index {
	// Because the main purpose of cloning is to append
	// another index, create the cloned field index to be
	// the same length, but with capacity for an additional index.
	clone := make(Index, len(ix), len(ix)+1)
	copy(clone, ix)
	return clone
}

// Equal returns true if ix is equal to v.
func (ix Index) Equal(v Index) bool {
	if len(ix) != len(v) {
		return false
	}
	for i := range ix {
		if ix[i] != v[i] {
			return false
		}
	}
	return true
}

// ValueRW returns the value of the field from the structure v.
// If any referenced field in v contains a nil pointer, then an
// empty value is created.
func (ix Index) ValueRW(v reflect.Value) reflect.Value {
	for _, i := range ix {
		v = reflect.Indirect(v).Field(i)
		// Create empty value for nil pointers, maps and slices.
		if v.Kind() == reflect.Ptr && v.IsNil() {
			a := reflect.New(v.Type().Elem())
			v.Set(a)
		} else if v.Kind() == reflect.Map && v.IsNil() {
			v.Set(reflect.MakeMap(v.Type()))
		} else if v.Kind() == reflect.Slice && v.IsNil() {
			v.Set(reflect.MakeSlice(v.Type(), 0, 0))
		}
	}
	return v
}

// ValueRO returns a value from the structure v without
// checking for nil pointers.
func (ix Index) ValueRO(v reflect.Value) reflect.Value {
	for _, i := range ix {
		v = reflect.Indirect(v).Field(i)
	}
	return v
}
