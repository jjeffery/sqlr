package column

// NamingConvention provides naming convention methods for
// inferring database column names from Go struct field names.
type NamingConvention interface {
	// Convert accepts the name of a Go struct field, and returns
	// the equivalent name for the field according to the naming convention.
	Convert(fieldName string) string

	// Join joins two or more string fragments to form a column name.
	// This method is used for naming columns based on fields within embedded
	// structures. The column name will be based on the name of
	// the Go struct field and its enclosing embedded struct fields.
	Join(frags []string) string
}
