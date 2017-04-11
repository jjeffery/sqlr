package column

import (
	"reflect"
	"strings"

	"github.com/jjeffery/sqlr/private/scanner"
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

	// FieldTag is the tag of the associated StructField.
	FieldTag reflect.StructTag
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
func NewPath(fieldName string, fieldTag reflect.StructTag) Path {
	var path Path
	return path.Append(fieldName, fieldTag)
}

// Append details of a field to an existing path to create
// a new path. The original path is unchanged.
func (path Path) Append(fieldName string, fieldTag reflect.StructTag) Path {
	clone := path.Clone()
	clone = append(path, struct {
		FieldName string
		FieldTag  reflect.StructTag
	}{
		FieldName: fieldName,
		FieldTag:  fieldTag,
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

// ColumnName returns a column name by applying the naming
// convention to the contents of the path.
func (path Path) ColumnName(nc NamingConvention, key string) string {
	if len(path) == 1 {
		// The path almost always has one element in it,
		// so have a special case that requires less memory
		// allocation.
		return convertField(path[0].FieldName, path[0].FieldTag, nc, key)
	}

	// Less common case where there is more than one item in the path.
	frags := make([]string, len(path))
	for i, f := range path {
		frags[i] = convertField(f.FieldName, f.FieldTag, nc, key)
	}
	return nc.Join(frags)
}

// structTagKeys specifies the list of struct tag keys that are searched
// in order for column information.
var structTagKeys = []string{"sqlr", "sql"}

func convertField(fieldName string, fieldTag reflect.StructTag, nc NamingConvention, key string) string {
	if fieldTag != "" {
		var nameFromTag string  // the name extracted from the tag, which might be empty
		var foundNameInTag bool // was the name extracted from the tag

		// First lookup the key for the naming convention. This key is different
		// because, if it exists and is blank, then it means to stop searching
		// and to use the naming convention rules.
		if key != "" {
			if value, ok := fieldTag.Lookup(key); ok {
				foundNameInTag = true
				nameFromTag = nameFromTagValue(value)
			}
		}

		// If the key for the naming convention was not provided, then
		// look through the standard struct tag keys. Keep looking until
		// one is found that specifies a name.
		if !foundNameInTag {
			for _, key := range structTagKeys {
				if value, ok := fieldTag.Lookup(key); ok {
					nameFromTag = nameFromTagValue(value)
					if nameFromTag != "" {
						foundNameInTag = true
						break
					}
				}
			}
		}

		if foundNameInTag && nameFromTag != "" {
			return nameFromTag
		}
	}

	// The name is not to be found in the struct field tag, so apply the
	// naming convention conversion rules.
	return nc.Convert(fieldName)
}

func nameFromTagValue(tagValue string) string {
	tagValue = strings.TrimSpace(tagValue)
	if tagValue == "" {
		return ""
	}
	scan := newScannerForString(tagValue)
	for scan.Scan() {
		tok, lit := scan.Token(), scan.Text()
		switch tok {
		case scanner.KEYWORD:
			// exit on first keyword, no column specified
			return ""
		case scanner.IDENT:
			// first identifier indicates the column name, and
			// may be quoted
			return scanner.Unquote(lit)
		case scanner.LITERAL:
			if scanner.IsQuoted(lit) {
				// a string literal is accepted as the column name
				return scanner.Unquote(lit)
			}
		case scanner.OP:
			if lit == "-" {
				// indicates should not be a column
				return lit
			}
			return ""
		}
	}
	return ""
}
