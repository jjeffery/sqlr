package sqlr

import (
	"github.com/jjeffery/sqlr/private/naming"
)

// The NamingConvention interface provides methods that are used to
// infer a database column name from its associated Go struct field,
// and a database table name from the name of its associated row type.
type NamingConvention interface {
	// Convert converts a Go struct field name according to the naming convention.
	Convert(fieldName string) string

	// Join joins two or more converted names to form a column name.
	// Used for naming columns based on fields within embedded
	// structures.
	Join(names []string) string

	// TableName converts the name of the row type into a table name.
	TableName(typeName string) string
}

// Pre-defined naming conventions. If a naming convention is not specified
// for a schema, it defaults to snake_case.
var (
	SnakeCase NamingConvention // eg "FieldName" -> "field_name"
	SameCase  NamingConvention // eg "FieldName" -> "FieldName"
	LowerCase NamingConvention // eg "FieldName" -> "fieldname"
)

var (
	// defaultNamingConvention is used for a schema if no naming
	// convention has been specified
	defaultNamingConvention = naming.SnakeCase
)

func init() {
	SnakeCase = naming.SnakeCase
	SameCase = naming.SameCase
	LowerCase = naming.LowerCase
}

/**

// PluralizeTableNames returns a naming convention based on nc,
// that pluralizes table names.
func PluralizeTableNames(nc NamingConvention) NamingConvention {
	panic("not implemented yet")
}

**/
