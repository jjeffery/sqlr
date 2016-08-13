package column

import (
	"strings"
)

// A Path contains information about all the StructFields traversed
// to obtain the value for a column.
//
// The significance of the path is that it is used to construct
// the column name, either by the column name(s) specified
// in the struct tags, or by applying a naming convention to the field
// name(s).
type Path []struct {
	// FieldName is the name of the associated StructField.
	FieldName string

	// ColumnName is the associated column name, extracted
	// from the StructTag. If no column name has been specified,
	// this field is blank.
	ColumnName string
}

// Clone creates a deep copy of path.
func (path Path) Clone() Path {
	// Because the main purpose of cloning is to append
	// another item, create the clone to be the same length,
	// but with capacity for an additional item.
	clone := make(Path, len(path), len(path)+1)
	copy(clone, path)
	return path
}

// NewPath returns a new path with a single field.
func NewPath(fieldName, columnName string) Path {
	var path Path
	return path.Append(fieldName, columnName)
}

// Append details of a field to an existing path to create
// a new path. The original path is unchanged.
func (path Path) Append(fieldName, columnName string) Path {
	clone := path.Clone()
	clone = append(path, struct {
		FieldName  string
		ColumnName string
	}{
		FieldName:  fieldName,
		ColumnName: columnName,
	})
	return clone
}

// Equal returns true if path and other are equal.
func (path Path) Equal(other Path) bool {
	if len(path) != len(other) {
		return false
	}
	for i, f := range path {
		if f != other[i] {
			return false
		}
	}
	return true
}

func (path Path) String() string {
	if len(path) == 0 {
		return ""
	}
	if len(path) == 1 {
		return path[0].FieldName
	}
	var fieldNames []string
	for _, item := range path {
		fieldNames = append(fieldNames, item.FieldName)
	}
	return strings.Join(fieldNames, ".")
}
