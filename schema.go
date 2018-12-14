package sqlr

import (
	"github.com/jjeffery/sqlr/private/column"
)

// Schema contains information known about the database schema and is used
// when generating SQL statements.
//
// Information stored in the schema includes the SQL dialect,
// and the naming convention used to convert Go struct field names
// into database column names, and Go type names into database
// table names.
//
// Although the zero value schema can be used and represents a database schema
// with default values, it is also common to use the NewSchema function to
// create a schema with options.
//
// A schema maintains an internal cache, which is used to store details of
// frequently called SQL commands for improved performance. All methods are
// safe to call concurrently from different goroutines.
type Schema struct {
	dialect    Dialect
	convention NamingConvention
	cache      stmtCache
	funcMap    funcMap
	fieldMap   *fieldMap
	identMap   *identMap
	tableMap   tableMap
	key        string

	init *schemaInit // only used during initialization
}

// schemaInit contains info that is only used during initialization
// of the schema. It contains data that has been collected from the
// schema options that needs to be processed once all options have
// been collected.
type schemaInit struct {
	tablesConfig TablesConfig
}

// NewSchema creates a schema with options.
//
// If the schema has any inconsistencies, then this function will panic.
// If there is an expectation that the options contain invalid data, call
// NewSchemaE instead. Errors due to inconsistencies
// are only possible if the WithTables option is specified.
func NewSchema(opts ...SchemaOption) *Schema {
	schema, err := NewSchemaE(opts...)
	if err != nil {
		panic(err)
	}
	return schema
}

// NewSchemaE creates a new schema with options. If the schema
// contains any inconsistencies, then an error is returned.
//
// Because a schema is usually created during program initialization,
// it is more common for a program to call NewSchema, which will
// panic if there are any inconsistencies in the schema configuration.
// Errors due to inconsistencies are only possible if the WithTables
// option is specified.
func NewSchemaE(opts ...SchemaOption) (*Schema, error) {
	schema := &Schema{
		init: &schemaInit{},
	}

	for _, opt := range opts {
		if opt != nil {
			if err := opt(schema); err != nil {
				return nil, err
			}
		}
	}

	// configure any tables specified at initialization
	if schema.init.tablesConfig != nil {
		for row, cfg := range schema.init.tablesConfig {
			rowType, err := getRowType(row)
			if err != nil {
				return nil, err
			}
			tbl, err := newTableWithConfig(schema, rowType, &cfg)
			if err != nil {
				return nil, err
			}
			schema.tableMap.add(rowType, tbl)
		}
	}

	// remove stuff only needed during initialization
	schema.init = nil

	return schema, nil
}

// TableFor returns the table information associated with
// row, which should be an instance of a struct type
// or a pointer to a struct type.
// If row does not refer to a struct type then a panic results.
func (s *Schema) TableFor(row interface{}) *Table {
	rowType, err := getRowType(row)
	if err != nil {
		panic(err)
	}
	tbl := s.tableMap.lookup(rowType)
	if tbl != nil {
		return tbl
	}
	// If we get here, then the table/row type mapping was
	// not supplied when the schema was created. Create the
	// table info for the row and add to the table map. Note
	// that the *Table returned from tableMap.add might be different
	// to tbl created by this function if another goroutine
	// has beaten us to creating an entry in the tableMap.
	tbl = newTable(s, rowType, nil)
	return s.tableMap.add(rowType, tbl)
}

// columnNamerFunc converts a function into a columnNamer.
// It implements the columnNamer interface used in the column package.
// TODO(jpj): this will be removed when fieldMap is removed, and the
// functionality will be moved from the column package into this package.
type columnNamerFunc func(*column.Info) string

func (f columnNamerFunc) ColumnName(col *column.Info) string {
	return f(col)
}

// columnNamer returns an object that implements the columnNamer interface
// for the schema. The column namer returns the column name based on the
// list of field name/column name mappings for the schema, and the naming
// convention.
func (s *Schema) columnNamer() columnNamerFunc {
	return columnNamerFunc(func(col *column.Info) string {
		if s.fieldMap != nil {
			if columnName, ok := s.fieldMap.lookup(col.FieldNames); ok {
				// If the field map returns an empty string, this means to
				// fallback to the naming convention. This provides a mechanism
				// to override any naming from a previous schema.
				if columnName != "" {
					return columnName
				}
			}
		}
		convention := s.convention
		if convention == nil {
			convention = defaultNamingConvention
		}
		return col.Path.ColumnName(convention, s.key)
	})
}

// renameIdent implements the identRenamer interface.
func (s *Schema) renameIdent(ident string) (string, bool) {
	if s.identMap == nil {
		return "", false
	}
	return s.identMap.lookup(ident)
}

// getDialect returns the dialect for the schema. The aim is to make
// an empty Schema usable, so this method is necessary to ensure that
// a non-nil dialect is always available.
func (s *Schema) getDialect() Dialect {
	if s.dialect != nil {
		return s.dialect
	}
	return DefaultDialect()
}

// Prepare creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the returned
// statement.
//
// Statements are low-level, and most programs do not need to use them
// directly. This method may be removed in a future version of the API.
func (s *Schema) Prepare(row interface{}, query string) (*Stmt, error) {
	// for queries that do not involve a row, just use an empty struct
	if row == nil {
		row = &struct{}{}
	}

	// determine row type to use for statement
	rowType, err := getRowType(row)
	if err != nil {
		return nil, err
	}

	// convert common shorthand SQL notations
	if query, err = checkSQL(query); err != nil {
		return nil, err
	}

	// attempt to get statement from the schema's statement cache
	stmt, ok := s.cache.lookup(rowType, query)
	if !ok {
		// build statement from scratch
		tbl := s.TableFor(rowType)
		stmt, err = newStmt(s, tbl, query)
		if err != nil {
			return nil, err
		}
		// add to schema's statement cache, returning the statement in the
		// cache -- this is just in case another goroutine has beaten us to it
		stmt = s.cache.set(rowType, query, stmt)
	}
	return stmt, nil
}

// Key returns the key associated with the schema, which is specififed using
// the WithKey schema option.
func (s *Schema) Key() string {
	return s.key
}
