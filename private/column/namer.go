package column

// NamingConvention provides naming convention methods for
// inferring database column names from Go struct field names.
type NamingConvention interface {
	// A short key used to identify the naming convention.
	// If not blank, this vaue is used as a key when looking
	// up the struct tag.
	Key() string

	// Convert accepts the name of a Go struct field, and returns
	// the equivalent name for the field according to the naming convention.
	Convert(fieldName string) string

	// Join joins two or more string fragments to form a column name.
	// This method is used for naming columns based on fields within embedded
	// structures. The column name will be based on the name of
	// the Go struct field and its enclosing embedded struct fields.
	Join(frags []string) string
}

// Namer knows how to name a column using a naming convention.
type Namer struct {
	nc NamingConvention
}

// NewNamer creates a namer for a naming convention.
func NewNamer(nc NamingConvention) *Namer {
	return &Namer{
		nc: nc,
	}
}

// ColumnName returns the column name.
func (n *Namer) ColumnName(info *Info) string {
	return info.Path.ColumnName(n.nc)
}
