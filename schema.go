package sqlr

import "github.com/jjeffery/sqlr/private/column"

// Schema contains information about the database that is used
// when generating SQL statements.
//
// Information stored in the schema includes the SQL dialect,
// and the naming convention used to convert Go struct field names
// into database column names.
//
// Although the zero value schema can be used and represents a database schema
// with default values, it is more common to use the NewSchema function to
// create a schema with one or more options.
//
// A schema maintains an internal cache, which is used to store details of
// frequently called SQL commands for improved performance.
//
// A schema can be inexpensively cloned to provide a deep copy.
// This can occasionally be useful to define a common schema for a database,
// and then create copies to handle naming rules that are specific to a particular
// table, or a particular group of tables.
type Schema struct {
	dialect    Dialect
	convention NamingConvention
	cache      stmtCache
	fieldMap   *fieldMap
	key        string
}

// NewSchema creates a schema with options.
func NewSchema(opts ...SchemaOption) *Schema {
	schema := &Schema{}
	for _, opt := range opts {
		if opt != nil {
			opt(schema)
		}
	}
	return schema
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

// getDialect returns the dialect for the schema. The aim is to make
// and empty Schema usable, so this method is necessary to ensure that
// a non-nil dialect is always available.
func (s *Schema) getDialect() Dialect {
	if s.dialect != nil {
		return s.dialect
	}
	return DefaultDialect
}

// Clone creates a copy of the schema, with options applied.
func (s *Schema) Clone(opts ...SchemaOption) *Schema {
	clone := &Schema{
		dialect:    s.dialect,
		convention: s.convention,
		fieldMap:   newFieldMap(s.fieldMap),
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
		stmt, err = newStmt(s.getDialect(), s.columnNamer(), rowType, query)
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
// The argument passed to rows can be one of the following:
//  A pointer to an array of structs; or
//  a pointer to an array of struct pointers; or
//  a pointer to a struct.
// When rows is a pointer to an array it is populated with
// one item for each row returned by the SELECT query.
//
// When rows is a pointer to a struct, it is populated with
// the first row returned from the query. This is a good
// option when the query will only return one row.
//
// Select returns the number of rows returned by the SELECT
// query.
func (s *Schema) Select(db DB, rows interface{}, sql string, args ...interface{}) (int, error) {
	stmt, err := s.Prepare(rows, sql)
	if err != nil {
		return 0, err
	}
	return stmt.Select(db, rows, args...)
}

// Exec executes the query with the given row and optional arguments.
// It returns the number of rows affected by the statement.
//
// If the statement is an INSERT statement and the row has an auto-increment field,
// then the row is updated with the value of the auto-increment column, as long as
// the SQL driver supports this functionality.
func (s *Schema) Exec(db DB, row interface{}, sql string, args ...interface{}) (int, error) {
	stmt, err := s.Prepare(row, sql)
	if err != nil {
		return 0, err
	}
	return stmt.Exec(db, row, args...)
}
