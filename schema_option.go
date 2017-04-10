package sqlrow

import "database/sql"

// A SchemaOption provides optional configuration and is supplied when
// creating a new Schema, or cloning a Schema.
type SchemaOption func(schema *Schema)

// ForDB creates an option that sets the dialect for the open DB handle.
func ForDB(db *sql.DB) SchemaOption {
	return func(schema *Schema) {
		schema.dialect = dialectFor(db)
		schema.cache.clear()
	}
}

// WithDialect provides an option that sets the schema's dialect.
func WithDialect(dialect Dialect) SchemaOption {
	return func(schema *Schema) {
		schema.dialect = dialect
		schema.cache.clear()
	}
}

// WithNamingConvention creates and option that sets the schema's naming convention.
func WithNamingConvention(convention NamingConvention) SchemaOption {
	return func(schema *Schema) {
		schema.convention = convention
		schema.cache.clear()
	}
}

// WithField creates an option that maps a Go field name to a
// database column name.
//
// It is more common to override column names in the struct tag of
// the field, but there are some cases where it makes sense to
// declare column name overrides directly with the schema. One situation
// is with fields within embedded structures. For example, with the following
// structures:
//  type UserRow struct {
//      Name string
//      HomeAddress Address
//      WorkAddress Address
//  }
//
//  type Address struct {
//      Street   string
//      Locality string
//      State    string
//  }
//
// If the column name for HomeAddress.Locality is called "home_suburb" for historical
// reasons, then it is not possible to specify a rename in the structure tag
// without also affecting the WorkAddress.Locality field. In this situation it is only
// possible to specify the column name override using the WithField option:
//  schema := NewSchema(
//      WithField("HomeAddress.Locality", "home_suburb"),
//  )
//
func WithField(fieldName string, columnName string) SchemaOption {
	return func(schema *Schema) {
		if schema.fieldMap == nil {
			schema.fieldMap = newFieldMap(schema.fieldMap)
		}
		schema.fieldMap.add(fieldName, columnName)
	}
}
