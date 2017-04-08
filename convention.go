package sqlrow

import "github.com/jjeffery/sqlrow/private/naming"

// Convention provides naming convention methods for
// inferring database column names from Go struct field names.
type Convention interface {
	// Key is a short name describing the naming convention.
	// If the key is not empty it is used to locate a value in
	// a struct field's tag specific to the naming convention.
	Key() string

	// Convert converts a Go struct field name according to the naming convention.
	Convert(fieldName string) string

	// Join joins two or more converted names to form a column name.
	// Used for naming columns based on fields within embedded
	// structures.
	Join(names []string) string
}

// ConventionSnake is the default, 'snake_case' naming convention. Its key is "snake".
var ConventionSnake Convention

// ConventionSame is a naming convention where the column name
// is identical to the Go struct field name. Its key is "same".
var ConventionSame Convention

// ConventionLower is a naming convention where the column name
// is the Go struct field name converted to lower case. This naming
// convention is useful for some PostgreSQL databases. Its key is "lower".
var ConventionLower Convention

func init() {
	ConventionSnake = naming.Snake
	ConventionSame = naming.Same
	ConventionLower = naming.Lower
}
