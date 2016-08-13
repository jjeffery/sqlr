package sqlrow

import "github.com/jjeffery/sqlrow/private/colname"

// Convention provides naming convention methods for
// inferring database column names from Go struct field names.
type Convention interface {
	// The name of the convention. This name is used as
	// a key for caching, so if If two conventions have
	// the same name, then they should be identical.
	Name() string

	// ColumnName returns the name of a database column based
	// on the name of a Go struct field.
	ColumnName(fieldName string) string

	// Join joins a prefix with a name to form a column name.
	// Used for naming columns based on fields within embedded
	// structures. The column name will be based on the name of
	// the Go struct field and its enclosing embedded struct fields.
	Join(prefix, name string) string
}

// ConventionSnake is the default, 'snake_case' naming convention
var ConventionSnake Convention

// ConventionSame is a naming convention where the column name
// is identical to the Go struct field name.
var ConventionSame Convention

func init() {
	ConventionSnake = colname.Snake
	ConventionSame = colname.Same
}
