package sqlr

/*
SchemaConfig was a possible replacement for the SchemaOptions. The thinking
was to replace the function-as-option pattern with a SchemaConfig, on the
basis that it would be more obvious. After a bit of experimentation, v0.6
will stick with the SchemaOptions, and just add WithTables to apply a
TablesConfig struct.

Keeping this code commented out, because it just might get used in v0.7 if
use in v0.6 suggests it is a better way to go.

// SchemaConfig contains optional configuration information about a schema.
type SchemaConfig struct {
	// NamingConvention is the naming convention used to convert
	// field names to column names, and type names to database
	// table names.
	//
	// If not specifed, then the default naming convention (snake_case)
	// is used.
	NamingConvention NamingConvention

	// Dialect is the SQL dialect used for generating SQL queries.
	// If the program only uses one SQL database driver, then there
	// is no need to specify this value. It is only required if
	// the program includes more than one SQL database driver.
	Dialect Dialect

	// StructTagKey optionally specifies the key used in struct tags
	// for specifying database schema information. If not specified,
	// the key used is "sql".
	StructTagKey string

	// Tables optionally specifies configuration for individual
	// database tables. Tables for which the default configuration
	// applies do not need to be included in this list.
	Tables TablesConfig
}
*/

// TablesConfig is a map of table configurations, keyed by the row type
// that represents the table.
type TablesConfig map[interface{}]TableConfig

// TableConfig contains configuration for an individual
// database table. It is not necessary to specify configuration
// for a table if the default settings apply.
type TableConfig struct {
	// TableName optionally specifies the name of the database
	// table associated with the row type.
	TableName string

	// Columns is an optional list of column configurations.
	// Only columns with non-default configuration need to
	// be included in this list.
	Columns ColumnsConfig
}

// ColumnsConfig is a map of individual column configurations, keyed
// by the field path to the field.
//
// The field path is the the field name when the field is a simple,
// scalar type. When the row type contains embedded struct
// fields, then the field path is all of the field names required
// to navigate to the field separated by a period.
//
//  type Row struct {
//      Name string         // field path = "Name"
//      Address struct {
//          Street   string // field path = "Address.Street"
//          Locality string	// field path = "Address.Locality"
//      }
//  }
type ColumnsConfig map[string]ColumnConfig

// ColumnConfig contains configuration for an individual
// database column.
//
// Column configuration is typically specified in the struct tag
// of the relevant struct field, so it is not usually necessary
// to specify column configuration using this structure.
type ColumnConfig struct {
	// ColumnName optionally specifies the database column
	// associated with the field.
	ColumnName string

	// Ignore optionally indicates that there is no database
	// column associated with this field.
	Ignore bool

	// PrimaryKey optionally indicates that this field forms
	// part of the primary key. When more than one field forms
	// part of the primary key, the order of columns is determined
	// by the order in the corresponding row struct.
	PrimaryKey bool

	// AutoIncrement optionally indicates that the column associated with
	// this field is an auto-incrementing column (aka an identity column).
	AutoIncrement bool

	// Version optionally indicates that the column associated with this
	// field is used for optimistic locking. The value of the field is incremented
	// every time the row is updated.
	Version bool

	// EmptyNull optionally indicates that an empty value in the field
	// should be interpreted as a NULL value in the database column, and a
	// NULL value in the database should be converted to an empty value.
	//
	// This setting is most applicable to strings and timestamps: it is
	// common to convert an empty string ("") and a zero time (time.Time{})
	// to a NULL value in the database.
	EmptyNull bool

	// JSON optionally indicates that the value of the field should be
	// marshaled into JSON before storing in the database, and that the
	// value in the database should be unmarshaled from JSON before
	// storing in the field.
	JSON bool

	// NaturalKey optionally indicates that the column forms part of a
	// natural key for the row. When a column forms part of a natural
	// key, then the value in that field is included in any error message
	// generated for row-level error conditions. This can be helpful
	// for diagnostics and debugging.
	NaturalKey bool

	// OverrideStructTag optionally specifies that the configuration
	// in this struct should override all configuration present in
	// the field's struct tag. This would only be used in unusual
	// circumstances.
	OverrideStructTag bool
}
