package sqlrow

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jjeffery/sqlrow/private/column"
	"github.com/jjeffery/sqlrow/private/scanner"
)

// Stmt is a prepared statement. A Stmt is safe for concurrent use by multiple goroutines.
type Stmt struct {
	rowType        reflect.Type
	queryType      queryType
	query          string
	dialect        Dialect
	convention     Convention
	argCount       int
	columns        []*column.Info
	inputs         []inputT
	outputs        []*column.Info
	autoIncrColumn *column.Info
}

func inferRowType(row interface{}, argName string) (reflect.Type, error) {
	rowType := reflect.TypeOf(row)
	if rowType.Kind() == reflect.Ptr {
		rowType = rowType.Elem()
	}
	if rowType.Kind() == reflect.Slice {
		rowType = rowType.Elem()
		if rowType.Kind() == reflect.Ptr {
			rowType = rowType.Elem()
		}
	}
	if rowType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected arg for %q to refer to a struct type", argName)
	}
	return rowType, nil
}

func newStmt(dialect Dialect, convention Convention, rowType reflect.Type, sql string) (*Stmt, error) {
	stmt := &Stmt{}
	stmt.dialect = dialect
	stmt.convention = convention
	stmt.rowType = rowType
	if stmt.rowType.Kind() != reflect.Struct {
		// should never happen, see inferRowType; could turn this into a panic
		return nil, errors.New("not a struct")
	}
	stmt.columns = column.ListForType(stmt.rowType)
	if err := stmt.scanSQL(sql); err != nil {
		return nil, err
	}

	if stmt.queryType == queryInsert {
		for _, col := range stmt.columns {
			if col.AutoIncrement {
				stmt.autoIncrColumn = col
				// TODO: return an error if col is not an integer type
				break
			}
		}

		if stmt.autoIncrColumn != nil {
			// Some DBs allow the auto-increment column to be specified.
			// Work out if this statement is doing this.
			for _, col := range stmt.inputs {
				if col.col == stmt.autoIncrColumn {
					// this statement is setting the auto-increment column explicitly
					stmt.autoIncrColumn = nil
					break
				}
			}
		}
	}

	return stmt, nil
}

// String prints the SQL query associated with the statement.
func (stmt *Stmt) String() string {
	return stmt.query
}

// Exec executes the prepared statement with the given row and optional arguments.
// It returns the number of rows affected by the statement.
//
// If the statement is an INSERT statement and the row has an auto-increment field,
// then the row is updated with the value of the auto-increment column as long as
// the SQL driver supports this functionality.
func (stmt *Stmt) Exec(db DB, row interface{}, args ...interface{}) (int, error) {
	if stmt.queryType == querySelect {
		return 0, errors.New("attempt to call Exec on select statement")
	}

	// field for setting the auto-increment value
	var field reflect.Value
	if stmt.autoIncrColumn != nil {
		rowVal := reflect.ValueOf(row)
		field = stmt.autoIncrColumn.Index.ValueRW(rowVal)
		if !field.CanSet() {
			return 0, fmt.Errorf("cannot set auto-increment value for type %s", rowVal.Type().Name())
		}
	}

	args, err := stmt.getArgs(row, args)
	if err != nil {
		return 0, err
	}
	result, err := db.Exec(stmt.query, args...)
	if err != nil {
		return 0, err
	}

	if field.IsValid() {
		n, err := result.LastInsertId()
		if err != nil {
			// The statement was successful but getting last insert ID failed.
			// Return error with the expectation that the calling program will
			// roll back the transaction.
			return 0, err
		}
		// TODO: could catch a panic here if the type is not int8, 1nt16, int32, int64
		// but it would be better to check when statement is prepared
		field.SetInt(n)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// The statement was successful but getting the row count failed.
		// Return error with the expectation that the calling program will
		// roll back the transaction.
		return 0, err
	}

	// assuming that rows affected fits in an int
	return int(rowsAffected), nil
}

// Select executes the prepared query statement with the given arguments and
// returns the query results in rows. If rows is a pointer to a slice of structs
// then one item is added to the slice for each row returned by the query. If row
// is a pointer to a struct then that struct is filled with the result of the first
// row returned by the query. In both cases Select returns the number of rows returned
// by the query.
func (stmt *Stmt) Select(db DB, rows interface{}, args ...interface{}) (int, error) {
	if rows == nil {
		return 0, errors.New("nil pointer")
	}
	destValue := reflect.ValueOf(rows)

	errorPtrType := func() error {
		expectedTypeName := stmt.expectedTypeName()
		return fmt.Errorf("expected rows to be *[]%s, *[]*%s, or *%s",
			expectedTypeName, expectedTypeName, expectedTypeName)
	}

	if destValue.Kind() != reflect.Ptr {
		return 0, errorPtrType()
	}
	if destValue.IsNil() {
		return 0, errors.New("nil pointer")
	}

	destValue = reflect.Indirect(destValue)
	destType := destValue.Type()
	if destType == stmt.rowType {
		// pointer to row struct, so only fetch one row
		return stmt.selectOne(db, rows, destValue, args)
	}

	// if not a pointer to a struct, should be a pointer to a
	// slice of structs or a pointer to a slice of struct pointers
	if destType.Kind() != reflect.Slice {
		return 0, errorPtrType()
	}
	sliceValue := destValue

	rowType := destType.Elem()
	isPtr := rowType.Kind() == reflect.Ptr
	if isPtr {
		rowType = rowType.Elem()
	}
	if rowType != stmt.rowType {
		return 0, errorPtrType()
	}

	sqlRows, err := db.Query(stmt.query, args...)
	if err != nil {
		return 0, err
	}
	defer sqlRows.Close()

	var rowCount = 0
	scanValues := make([]interface{}, len(stmt.columns))

	for sqlRows.Next() {
		rowCount++
		rowValuePtr := reflect.New(rowType)
		rowValue := reflect.Indirect(rowValuePtr)
		var jsonCells []*jsonCell
		for i, col := range stmt.outputs {
			cellValue := col.Index.ValueRW(rowValue).Addr().Interface()
			if col.JSON {
				jc := newJSONCell(col.Field.Name, cellValue)
				jsonCells = append(jsonCells, jc)
				scanValues[i] = jc.ScanValue()
			} else {
				scanValues[i] = cellValue
			}
		}
		err = sqlRows.Scan(scanValues...)
		if err != nil {
			return 0, err
		}
		for _, jc := range jsonCells {
			if err := jc.Unmarshal(); err != nil {
				return rowCount, err
			}
		}
		if isPtr {
			sliceValue.Set(reflect.Append(sliceValue, rowValuePtr))
		} else {
			sliceValue.Set(reflect.Append(sliceValue, rowValue))
		}
	}

	return rowCount, sqlRows.Err()
}

func (stmt *Stmt) selectOne(db DB, dest interface{}, rowValue reflect.Value, args []interface{}) (int, error) {
	rows, err := db.Query(stmt.query, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	scanValues := make([]interface{}, len(stmt.outputs))
	var jsonCells []*jsonCell

	if !rows.Next() {
		// no rows returned
		return 0, nil
	}

	// at least one row returned
	rowCount := 1

	for i, col := range stmt.outputs {
		cellValue := col.Index.ValueRW(rowValue).Addr().Interface()
		if col.JSON {
			jc := newJSONCell(col.Field.Name, cellValue)
			jsonCells = append(jsonCells, jc)
			scanValues[i] = jc.ScanValue()
		} else {
			scanValues[i] = cellValue
		}
	}
	err = rows.Scan(scanValues...)
	if err != nil {
		return 0, err
	}
	for _, jc := range jsonCells {
		if err := jc.Unmarshal(); err != nil {
			return rowCount, err
		}
	}

	// count any additional rows
	for rows.Next() {
		rowCount++
	}

	return rowCount, nil
}

func (stmt *Stmt) addColumns(cols columnsT) {
	if cols.clause.isInput() {
		for _, col := range cols.filtered() {
			stmt.inputs = append(stmt.inputs, inputT{col: col})
		}
	} else if cols.clause.isOutput() {
		for _, col := range cols.filtered() {
			stmt.outputs = append(stmt.outputs, col)
		}
	}
}

func (stmt *Stmt) scanSQL(query string) error {
	query = strings.TrimSpace(query)
	scan := scanner.New(strings.NewReader(query))
	var counter counterT
	columns := newColumns(stmt.columns, stmt.convention, stmt.dialect, counter.Next)
	var insertColumns *columnsT
	var clause sqlClause
	var buf bytes.Buffer
	for scan.Scan() {
		tok, lit := scan.Token(), scan.Text()
		switch tok {
		case scanner.WS:
			buf.WriteRune(' ')
		case scanner.COMMENT:
			// strip comment
		case scanner.LITERAL, scanner.OP:
			buf.WriteString(lit)
		case scanner.PLACEHOLDER:
			buf.WriteString(stmt.dialect.Placeholder(counter.Next()))
			stmt.inputs = append(stmt.inputs, inputT{argIndex: stmt.argCount})
			stmt.argCount++
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

				// An unquoted identifer might be an SQL keyword.
				// Attempt to infer the SQL clause and query type.
				clause = clause.nextClause(lit)
				if stmt.queryType == queryUnknown {
					stmt.queryType = clause.queryType()
				}
			}
		}
	}
	stmt.query = strings.TrimSpace(buf.String())
	return nil
}

// getArgs returns an array of args to send to the SQL query, based
// on the contents of the row and the args passed in (renamed here to argv).
// When getting args for a SELECT query, row will be nil and the argv array
// has to supply everything.
func (stmt *Stmt) getArgs(row interface{}, argv []interface{}) ([]interface{}, error) {
	if len(argv) != stmt.argCount {
		return nil, fmt.Errorf("expected arg count=%d, actual=%d", stmt.argCount, len(argv))
	}
	var args []interface{}

	rowVal := reflect.ValueOf(row)
	for rowVal.Type().Kind() == reflect.Ptr {
		rowVal = rowVal.Elem()
	}
	if rowVal.Type() != stmt.rowType {
		// should never happen, calling functions have already checked
		expectedType := stmt.expectedTypeName()
		return nil, fmt.Errorf("expected type %s or *(%s)", expectedType, expectedType)
	}

	for _, input := range stmt.inputs {
		if input.col != nil {
			if input.col.JSON {
				// marshal field contents into JSON and pass as a byte array
				valueRO := input.col.Index.ValueRO(rowVal).Interface()
				if valueRO == nil {
					args = append(args, nil)
				} else {
					data, err := json.Marshal(valueRO)
					if err != nil {
						// TODO(jpj): if errors.Wrap makes it into the stdlib, use it here
						err = fmt.Errorf("cannot marshal field %q: %v", input.col.Field.Name, err)
						return nil, err
					}
					args = append(args, data)
				}
			} else {
				args = append(args, input.col.Index.ValueRO(rowVal).Interface())
			}
		} else {
			args = append(args, argv[input.argIndex])
		}
	}

	return args, nil
}

func (stmt *Stmt) expectedTypeName() string {
	return fmt.Sprintf("%s.%s", stmt.rowType.PkgPath(), stmt.rowType.Name())
}

// counterT is used for keeping track of placeholders
type counterT int

func (c *counterT) Next() int {
	*c++
	return int(*c)
}

// inputT describes an input to an SQL query.
//
// If col is non-nil, then it refers to the column/field
// used as the input for the corresponding placeholder in the
// SQL query.
//
// If col is nil, then argIndex is the index into the args
// array for the associated arg that will be used for the placeholder.
type inputT struct {
	col      *column.Info
	argIndex int // used only if col == nil
}

// jsonCell is used to unmarshal JSON cells into their destination type
type jsonCell struct {
	colname   string
	cellValue interface{}
	data      []byte
}

func newJSONCell(colname string, v interface{}) *jsonCell {
	return &jsonCell{
		colname:   colname,
		cellValue: v,
	}
}

func (jc *jsonCell) ScanValue() interface{} {
	return &jc.data
}

func (jc *jsonCell) Unmarshal() error {
	if len(jc.data) == 0 {
		// No JSON data to unmarshal, so set to the zero value
		// for this type. We know that jc.cellValue is a pointer,
		// so it is safe to call Elem() and set the value.
		valptr := reflect.ValueOf(jc.cellValue)
		val := valptr.Elem()
		val.Set(reflect.Zero(val.Type()))
		return nil
	}
	if err := json.Unmarshal(jc.data, jc.cellValue); err != nil {
		// TODO(jpj): if Wrap makes it into the stdlib, use it here
		return fmt.Errorf("cannot unmarshal JSON field %q: %v", jc.colname, err)
	}
	return nil
}
