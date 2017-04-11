package sqlrow

import "github.com/jjeffery/sqlrow/private/column"

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

type schemaHelper struct {
	schema *Schema
}

func (s schemaHelper) Dialect() Dialect {
	if s.schema.dialect != nil {
		return s.schema.dialect
	}
	return DefaultDialect
}

func (s schemaHelper) NamingConvention() NamingConvention {
	if s.schema.convention != nil {
		return s.schema.convention
	}
	return SnakeCase
}

func (s schemaHelper) ColumnName(col *column.Info) string {
	if s.schema.fieldMap != nil {
		if columnName := s.schema.fieldMap.lookup(col.FieldNames); columnName != "" {
			return columnName
		}
	}
	return col.Path.ColumnName(s.NamingConvention(), s.schema.key)
}

func (s *Schema) columnNamer() columnNamer {
	return columnNamerFunc(func(col *column.Info) string {
		if s.fieldMap != nil {
			if columnName := s.fieldMap.lookup(col.FieldNames); columnName != "" {
				return columnName
			}
		}
		convention := s.convention
		if convention == nil {
			convention = SnakeCase
		}
		return col.Path.ColumnName(convention, s.key)
	})
}

func (s *Schema) getDialect() Dialect {
	if s.dialect != nil {
		return s.dialect
	}
	return DefaultDialect
}

// Prepare creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the returned
// statement.
func (s *Schema) Prepare(row interface{}, sql string) (*Stmt, error) {
	rowType, err := inferRowType(row)
	if err != nil {
		return nil, err
	}

	stmt, _ := s.cache.lookup(rowType, sql)
	if stmt == nil {
		stmt, err = newStmt(s.getDialect(), s.columnNamer(), rowType, sql)
		if err != nil {
			return nil, err
		}
		stmt = s.cache.set(rowType, sql, stmt)
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
