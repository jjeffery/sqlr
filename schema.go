package sqlr

import (
	"context"
	"errors"
	"reflect"

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
// with default values, it is also common to use the MustCreateSchema function to
// create a schema from a SchemaConfig struct.
//
// A schema maintains an internal cache, which is used to store details of
// frequently called SQL commands for improved performance.
type Schema struct {
	dialect    Dialect
	convention NamingConvention
	cache      stmtCache
	fieldMap   *fieldMap
	identMap   *identMap
	key        string
}

// NewSchema creates a schema with options.
//
// Deprecated: Use MustCreateSchema (or CreateSchema) instead.
func NewSchema(opts ...SchemaOption) *Schema {
	schema := &Schema{}
	for _, opt := range opts {
		if opt != nil {
			opt(schema)
		}
	}
	return schema
}

// MustCreateSchema creates a new schema based on the schema config. If the
// schema config contains any inconsistencies, then this function will panic.
func MustCreateSchema(config SchemaConfig) *Schema {
	schema, err := CreateSchema(config)
	if err != nil {
		panic(err)
	}
	return schema
}

// CreateSchema creates a new schema based on the schema config. If the
// schema config contains any inconsistencies, then an error is returned.
//
// It is more common for a program to call MustCreateSchema, which will
// panic if there are any inconsistencies in the schema configuration.
func CreateSchema(config SchemaConfig) (*Schema, error) {
	return nil, errors.New("not implemented yet")
}

// columnNamer returns an object that implements the columnNamer interface
// for the schema. The column namer returns the column name based on the
// list of field name/column name mappings for the schema, and the naming
// convention.
func (s *Schema) columnNamer() columnNamer {
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

func (s *Schema) tableNamer() func(rowType reflect.Type) string {
	// TODO(jpj): this is pretty lame, does not lookup exceptions
	// or struct tag data. Just a basic implementation that mirrors
	// columnNamer, that will probably be refactored later.
	return func(rowType reflect.Type) string {
		convention := s.convention
		if convention == nil {
			convention = defaultNamingConvention
		}
		return convention.TableName(rowType.Name())
	}
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

// Clone creates a copy of the schema, with options applied.
//
// Deprecated: This method will be removed. If two similar
// schemas are required, copy a SchemaConfig and make necessary
// changes before calling CreateSchema/MustCreateSchema.
func (s *Schema) Clone(opts ...SchemaOption) *Schema {
	clone := &Schema{
		dialect:    s.dialect,
		convention: s.convention,
		fieldMap:   newFieldMap(s.fieldMap),
		identMap:   newIdentMap(s.identMap),
		key:        s.key,
	}
	for _, opt := range opts {
		opt(clone)
	}
	return clone
}

// Prepare creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the returned
// statement.
//
// Deprecated: Use Session object and run queries directly using Session.Select,
// Session.Exec, Session.Insert, Session.Update or Session.Upsert.
func (s *Schema) Prepare(row interface{}, query string) (*Stmt, error) {
	// determine row type to use for statement
	rowType, err := inferRowType(row)
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
		stmt, err = newStmt(s.getDialect(), s.columnNamer(), s, rowType, query)
		if err != nil {
			return nil, err
		}
		// add to schema's statement cache, returning the statement in the
		// cache -- this is just in case another goroutine has beaten us to it
		stmt = s.cache.set(rowType, query, stmt)
	}
	return stmt, nil
}

// Select executes a SELECT query and stores the result in rows.
//
// Deprecated: Use Session.Select instead.
func (s *Schema) Select(db Querier, rows interface{}, sql string, args ...interface{}) (int, error) {
	stmt, err := s.Prepare(rows, sql)
	if err != nil {
		return 0, err
	}
	return stmt.Select(context.TODO(), db, rows, args...)
}

// Exec executes the query with the given row and optional arguments.
// It returns the number of rows affected by the statement.
//
// Deprecated: use Session.Exec instead.
//
// If the statement is an INSERT statement and the row has an auto-increment field,
// then the row is updated with the value of the auto-increment column, as long as
// the SQL driver supports this functionality.
func (s *Schema) Exec(db Querier, row interface{}, sql string, args ...interface{}) (int, error) {
	stmt, err := s.Prepare(row, sql)
	if err != nil {
		return 0, err
	}
	return stmt.Exec(context.TODO(), db, row, args...)
}

// Key returns the key associated with the schema.
func (s *Schema) Key() string {
	return s.key
}
