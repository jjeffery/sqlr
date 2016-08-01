package sqlstmt

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/jjeffery/sqlstmt/private/column"
	"github.com/jjeffery/sqlstmt/private/scanner"
)

var errNotImplemented = errors.New("not implemented")

// InsertRowStmt inserts a single row. It is safe for concurrent
// access by multiple goroutines.
type InsertRowStmt struct {
	commonStmt
	autoIncrColumn *column.Info
}

// NewInsertRowStmt returns a new InsertRowStmt for the given
// row and SQL.
// It is safe for concurrent access by multiple goroutines.
// The dialect and naming conventions are inferred
// from DefaultSchema.
func NewInsertRowStmt(row interface{}, sql string) *InsertRowStmt {
	return newInsertRowStmt(nil, row, sql)
}

func newInsertRowStmt(schema *Schema, row interface{}, sql string) *InsertRowStmt {
	stmt := &InsertRowStmt{}
	stmt.onceFunc = func() {
		// Initialize on first run. This way the calling program can define
		// statements in init() functions without having to be concerned with
		// the order that schemas/dialects/conventions are defined.
		sql = checkSQL(sql, insertFormat)
		stmt.err = stmt.prepareExecRow(schema, row, sql)

		for _, col := range stmt.columns {
			if col.AutoIncrement {
				stmt.autoIncrColumn = col
				break
			}
		}

		if stmt.autoIncrColumn != nil {
			// Some DBs allow the auto-increment column to be specified.
			// Work out if this statement is doing this.
			for _, col := range stmt.inputs {
				if col == stmt.autoIncrColumn {
					// this statement is setting the auto-increment column explicitly
					stmt.autoIncrColumn = nil
					break
				}
			}
		}
	}

	return stmt
}

// Exec executes the insert statement using the row as arguments.
func (stmt *InsertRowStmt) Exec(db DB, row interface{}) error {
	stmt.initOnce()

	// field for setting the auto-increment value
	var field reflect.Value
	if stmt.autoIncrColumn != nil {
		rowVal := reflect.ValueOf(row)
		field = stmt.autoIncrColumn.Index.ValueRW(rowVal)
		if !field.CanSet() {
			return fmt.Errorf("cannot set auto-increment value for type %s", rowVal.Type().Name())
		}
	}

	result, err := stmt.doExec(db, row)
	if err != nil {
		return err
	}

	if field.IsValid() {
		n, err := result.LastInsertId()
		if err != nil {
			return err
		}
		// TODO: could catch a panic here if the type is not int8, 1nt16, int32, int64
		field.SetInt(n)
	}
	return nil
}

// ExecRowStmt updates or deletes a single row. It is safe for concurrent
// access by multiple goroutines.
type ExecRowStmt struct {
	commonStmt
}

// NewUpdateRowStmt returns a new ExecRowStmt for updating a single row.
// It is safe for concurrent access by multiple goroutines.
// The dialect and naming conventions are obtained from DefaultSchema.
func NewUpdateRowStmt(row interface{}, sql string) *ExecRowStmt {
	return newUpdateRowStmt(nil, row, sql)
}

func newUpdateRowStmt(schema *Schema, row interface{}, sql string) *ExecRowStmt {
	stmt := &ExecRowStmt{}
	stmt.onceFunc = func() {
		// Initialize on first run. This way the calling program can define
		// statements in init() functions without having to be concerned with
		// the order that schemas/dialects/conventions are defined.
		sql = checkSQL(sql, updateFormat)
		stmt.err = stmt.prepareExecRow(schema, row, sql)
	}
	return stmt
}

// NewDeleteRowStmt returns a new ExecRowStmt for deleting a single
// row.
// It is safe for concurrent access by multiple goroutines.
// The dialect and naming conventions are obtained from DefaultSchema.
func NewDeleteRowStmt(row interface{}, sql string) *ExecRowStmt {
	return newDeleteRowStmt(nil, row, sql)
}

func newDeleteRowStmt(schema *Schema, row interface{}, sql string) *ExecRowStmt {
	stmt := &ExecRowStmt{}
	stmt.onceFunc = func() {
		// Initialize on first run. This way the calling program can define
		// statements in init() functions without having to be concerned with
		// the order that schemas/dialects/conventions are defined.
		sql = checkSQL(sql, deleteFormat)
		stmt.err = stmt.prepareExecRow(schema, row, sql)
	}
	return stmt
}

// Exec executes the statement using the row as arguments. Returns the
// number of rows affected.
func (stmt *ExecRowStmt) Exec(db DB, row interface{}) (int, error) {
	stmt.initOnce()
	result, err := stmt.doExec(db, row)
	if err != nil {
		return 0, err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(n), nil
}

// GetRowStmt executes a query that returns a single row.
// It is safe for concurrent access by multiple goroutines.
type GetRowStmt struct {
	commonStmt
}

// NewGetRowStmt returns a new GetRowStmt for the given
// row and SQL.
// It is safe for concurrent access by multiple goroutines.
// The dialect and naming conventions are obtained
// from DefaultSchema.
func NewGetRowStmt(row interface{}, sql string) *GetRowStmt {
	return newGetRowStmt(nil, row, sql)
}

func newGetRowStmt(schema *Schema, row interface{}, sql string) *GetRowStmt {
	stmt := &GetRowStmt{}
	stmt.onceFunc = func() {
		sql = checkSQL(sql, getFormat)
		stmt.err = stmt.prepareCommon(schema, row, sql)
	}
	return stmt
}

// Get a single row into dest based on the fields populated in dest.
func (stmt *GetRowStmt) Get(db DB, dest interface{}) (int, error) {
	stmt.initOnce()
	if stmt.err != nil {
		return 0, stmt.err
	}
	errorPtrType := func() error {
		expectedTypeName := stmt.expectedTypeName()
		return fmt.Errorf("expected dest to be *%s", expectedTypeName)
	}

	destValue := reflect.ValueOf(dest)

	if destValue.Kind() != reflect.Ptr {
		return 0, errorPtrType()
	}
	if destValue.IsNil() {
		return 0, errors.New("nil pointer passed")
	}

	rowValue := reflect.Indirect(destValue)
	rowType := rowValue.Type()
	if rowType != stmt.rowType {
		return 0, errorPtrType()
	}

	args, err := stmt.getArgs(dest)
	if err != nil {
		return 0, err
	}

	if stmt.Logger != nil {
		msg := fmt.Sprintf("query=%q, args=%v\n", stmt.query, args)
		stmt.Logger.Print(msg)
	}
	rows, err := db.Query(stmt.query, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	scanValues := make([]interface{}, len(stmt.outputs))

	if !rows.Next() {
		// no rows returned
		return 0, nil
	}

	for i, col := range stmt.outputs {
		cellValue := col.Index.ValueRW(rowValue)
		scanValues[i] = cellValue.Addr().Interface()
	}
	err = rows.Scan(scanValues...)
	if err != nil {
		return 0, err
	}

	return 1, nil
}

// SelectStmt executes a query that returns multiple rows.
// It is safe for concurrent access by multiple goroutines.
type SelectStmt struct {
	commonStmt
}

// NewSelectStmt returns a new SelectStmt for the given
// row and SQL. The dialect and naming conventions are obtained
// from DefaultSchema.
func NewSelectStmt(row interface{}, sql string) *SelectStmt {
	return newSelectStmt(nil, row, sql)
}

func newSelectStmt(schema *Schema, row interface{}, sql string) *SelectStmt {
	stmt := &SelectStmt{}
	stmt.onceFunc = func() {
		sql = checkSQL(sql, selectFormat)
		stmt.err = stmt.prepareCommon(schema, row, sql)
		if stmt.err == nil && len(stmt.inputs) > 0 {
			stmt.err = errors.New("unexpected inputs in query")
		}
	}
	return stmt
}

// Select executes the statement's query and returns the resulting
// rows in the slice pointed to by dest. The args are for any
// placeholder parameters in the query.
func (stmt *SelectStmt) Select(db DB, dest interface{}, args ...interface{}) error {
	stmt.initOnce()
	if stmt.err != nil {
		return stmt.err
	}
	destValue := reflect.ValueOf(dest)

	if destValue.Kind() != reflect.Ptr {
		return stmt.errorSliceType()
	}
	if destValue.IsNil() {
		return errors.New("Select: nil pointer passed as dest")
	}

	sliceValue := reflect.Indirect(destValue)
	sliceType := sliceValue.Type()
	if sliceType.Kind() != reflect.Slice {
		return stmt.errorSliceType()
	}

	rowType := sliceType.Elem()
	isPtr := rowType.Kind() == reflect.Ptr
	if isPtr {
		rowType = rowType.Elem()
	}
	if rowType != stmt.rowType {
		return stmt.errorSliceType()
	}

	rows, err := db.Query(stmt.query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	scanValues := make([]interface{}, len(stmt.columns))

	for rows.Next() {
		rowValuePtr := reflect.New(rowType)
		rowValue := reflect.Indirect(rowValuePtr)
		for i, col := range stmt.columns {
			cellValue := col.Index.ValueRW(rowValue)
			scanValues[i] = cellValue.Addr().Interface()
		}
		err = rows.Scan(scanValues...)
		if err != nil {
			return err
		}
		if isPtr {
			sliceValue.Set(reflect.Append(sliceValue, rowValuePtr))
		} else {
			sliceValue.Set(reflect.Append(sliceValue, rowValue))
		}
	}

	return rows.Err()
}

func (stmt *SelectStmt) errorSliceType() error {
	expectedType := stmt.expectedTypeName()
	return fmt.Errorf("Expected dest to be pointer to []%s or []*%s", expectedType, expectedType)
}

type commonStmt struct {
	// Logger is used for diagnostic logging.
	Logger Logger

	rowType    reflect.Type
	query      string
	dialect    Dialect
	convention Convention
	columns    []*column.Info
	inputs     []*column.Info
	outputs    []*column.Info
	err        error
	once       sync.Once
	onceFunc   func()
}

// String prints the SQL query associated with the statement.
func (stmt *commonStmt) String() string {
	stmt.initOnce()
	return stmt.query
}

// initOnce initializes the statement on first run. This way the calling program can
// define statements in init() functions without having to be concerned with
// the order that schemas/dialects/conventions are defined.
func (stmt *commonStmt) initOnce() {
	stmt.once.Do(func() {
		stmt.onceFunc()
		stmt.onceFunc = nil
	})
}

func (stmt *commonStmt) prepareCommon(schema *Schema, row interface{}, sql string) error {
	if schema == nil {
		schema = Default
		if schema == nil {
			schema = &Schema{}
		}
	}
	stmt.rowType = reflect.TypeOf(row)
	if stmt.rowType.Kind() == reflect.Ptr {
		stmt.rowType = stmt.rowType.Elem()
	}
	stmt.columns = column.ListForType(stmt.rowType)
	stmt.convention = schema.convention()
	stmt.dialect = schema.dialect()
	stmt.Logger = schema.Logger
	if err := stmt.scanSQL(sql); err != nil {
		return err
	}
	if stmt.Logger != nil {
		msg := fmt.Sprintf("prepared=%q", stmt.query)
		stmt.Logger.Print(msg)
	}
	return nil
}

func (stmt *commonStmt) prepareExecRow(schema *Schema, row interface{}, sql string) error {
	if err := stmt.prepareCommon(schema, row, sql); err != nil {
		return err
	}
	if len(stmt.outputs) > 0 {
		return errors.New("unexpected query columns in exec statement")
	}
	return nil
}

func (stmt *commonStmt) addColumns(cols columnsT) {
	if cols.clause.isInput() {
		for _, col := range cols.filtered() {
			stmt.inputs = append(stmt.inputs, col)
		}
	} else if cols.clause.isOutput() {
		for _, col := range cols.filtered() {
			stmt.outputs = append(stmt.outputs, col)
		}
	}
}

// counter is used for keeping track of placeholders
type counter int

func (c *counter) Next() int {
	*c++
	return int(*c)
}

func (stmt *commonStmt) scanSQL(query string) error {
	query = strings.TrimSpace(query)
	scan := scanner.New(strings.NewReader(query))
	var counter counter
	columns := newColumns(stmt.columns, stmt.convention, stmt.dialect, counter.Next)
	var insertColumns *columnsT
	var clause sqlClause
	var buf bytes.Buffer
	for {
		tok, lit := scan.Scan()
		if tok == scanner.EOF {
			break
		}
		switch tok {
		case scanner.WS:
			buf.WriteRune(' ')
		case scanner.COMMENT:
			// strip comment
		case scanner.LITERAL, scanner.OP:
			buf.WriteString(lit)
		case scanner.PLACEHOLDER:
			buf.WriteString(stmt.dialect.Placeholder(counter.Next()))
		case scanner.IDENT:
			if lit[0] == '{' {
				if !clause.acceptsColumns() {
					// invalid place to insert columns
					return fmt.Errorf("cannot expand %q in %q clause", lit, clause)
				}
				lit = strings.TrimSpace(scanner.Unquote(lit))
				if clause == clauseInsertValues {
					if lit != "" {
						return fmt.Errorf("columns for %q clause must match the %q clause",
							clause, clauseInsertColumns)
					}
					if insertColumns == nil {
						return fmt.Errorf("cannot expand %q clause because %q clause is missing",
							clause, clauseInsertColumns)
					}

					// change the clause but keep the filter and generate string
					cols := *insertColumns
					cols.clause = clause
					buf.WriteString(cols.String())
					stmt.addColumns(cols)
				} else {
					cols, err := columns.Parse(clause, lit)
					if err != nil {
						return fmt.Errorf("cannot expand %q in %q clause: %v", lit, clause, err)
					}
					buf.WriteString(cols.String())
					stmt.addColumns(cols)
					if clause == clauseInsertColumns {
						insertColumns = &cols
					}
				}
			} else if scanner.IsQuoted(lit) {
				lit = scanner.Unquote(lit)
				buf.WriteString(stmt.dialect.Quote(lit))
			} else {
				buf.WriteString(lit)

				// an unquoted identifer might be an SQL keyword
				clause = clause.nextClause(lit)
			}
		}
	}
	stmt.query = buf.String()
	return nil
}

func (stmt commonStmt) doExec(db DB, row interface{}) (sql.Result, error) {
	if stmt.err != nil {
		return nil, stmt.err
	}
	args, err := stmt.getArgs(row)
	if err != nil {
		return nil, err
	}
	if stmt.Logger != nil {
		msg := fmt.Sprintf("query=%q, args=%v\n", stmt.query, args)
		stmt.Logger.Print(msg)
	}
	return db.Exec(stmt.query, args...)
}

func (stmt commonStmt) getArgs(row interface{}) ([]interface{}, error) {
	if stmt.err != nil {
		return nil, stmt.err
	}

	var args []interface{}

	rowVal := reflect.ValueOf(row)
	for rowVal.Type().Kind() == reflect.Ptr {
		rowVal = rowVal.Elem()
	}
	if rowVal.Type() != stmt.rowType {
		expectedType := stmt.expectedTypeName()
		return nil, fmt.Errorf("expected type %s or *(%s)", expectedType, expectedType)
	}

	for _, input := range stmt.inputs {
		args = append(args, input.Index.ValueRO(rowVal).Interface())
	}

	return args, nil
}

func (stmt commonStmt) expectedTypeName() string {
	return fmt.Sprintf("%s.%s", stmt.rowType.PkgPath(), stmt.rowType.Name())
}
