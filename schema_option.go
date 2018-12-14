package sqlr

import "database/sql"

// A SchemaOption provides optional configuration and is supplied when
// creating a new Schema.
type SchemaOption func(schema *Schema) error

// ForDB creates an option that sets the dialect based on the DB handle.
func ForDB(db *sql.DB) SchemaOption {
	return func(schema *Schema) error {
		schema.dialect = dialectFor(db)
		schema.cache.clear()
		return nil
	}
}

// WithDialect provides an option that sets the schema's dialect.
func WithDialect(dialect Dialect) SchemaOption {
	return func(schema *Schema) error {
		schema.dialect = dialect
		schema.cache.clear()
		return nil
	}
}

// WithNamingConvention creates and option that sets the schema's naming convention.
func WithNamingConvention(convention NamingConvention) SchemaOption {
	return func(schema *Schema) error {
		schema.convention = convention
		schema.cache.clear()
		return nil
	}
}

// WithField creates an option that maps a Go field name to a
// database column name.
//
// Deprecated: define field/column name mappings using TableConfig instead.
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
	return func(schema *Schema) error {
		if schema.fieldMap == nil {
			schema.fieldMap = newFieldMap(schema.fieldMap)
		}
		schema.fieldMap.add(fieldName, columnName)
		return nil
	}
}

// WithIdentifier creates an option that performs a global rename
// of an identifier when preparing SQL queries. This option is not
// needed very often: its main purpose is for helping a program
// operate against two different database schemas where table and
// column names follow a different naming convention.
//
// The example shows a situation where a program operates against
// an SQL Server database where a table is named "[User]", but the
// same table is named "users" in the Postgres schema.
//
// Deprecated: Use (TODO) instead. This is a confusing API, and another,
// clearer option will be provided.
func WithIdentifier(identifier string, meaning string) SchemaOption {
	return func(schema *Schema) error {
		if schema.identMap == nil {
			schema.identMap = newIdentMap(schema.identMap)
		}
		schema.identMap.add(meaning, identifier)
		return nil
	}
}

// WithKey creates an option that associates the schema
// with a key in struct field tags. This option is not needed
// very often: its main purpose is for helping a program operate
// against two different database schemas.
func WithKey(key string) SchemaOption {
	return func(schema *Schema) error {
		schema.key = key
		return nil
	}
}

// WithTables creates an option that sets configuration for
// one or more tables and their associated row types.
func WithTables(tables TablesConfig) SchemaOption {
	return func(schema *Schema) error {
		// Can't do much with this now, because we might not have
		// the naming convention finalized yet, so just store it
		// for processing later.
		schema.init.tablesConfig = tables
		return nil
	}
}
