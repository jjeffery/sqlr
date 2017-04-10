package sqlrow

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jjeffery/sqlrow/private/column"
	"github.com/jjeffery/sqlrow/private/scanner"
	"github.com/jjeffery/sqlrow/private/wherein"
)

// Stmt is a prepared statement. A Stmt is safe for concurrent use by multiple goroutines.
type Stmt struct {
	rowType    reflect.Type
	queryType  queryType
	query      string
	dialect    Dialect
	convention columnNamer
	argCount   int
	columns    []*column.Info
	inputs     []inputT
	output     struct {
		mutex   sync.RWMutex
		columns []*column.Info
	}
	autoIncrColumn *column.Info
}

// TODO(SELECT): inferRowType should handle scalars: string, int, float, time.Time and types
// based on these types.

func inferRowType(row interface{}) (reflect.Type, error) {
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
		return nil, errors.New("expected arg to refer to a struct type")
	}
	return rowType, nil
}

func newStmt(dialect Dialect, convention columnNamer, rowType reflect.Type, sql string) (*Stmt, error) {
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
			if col.Tag.AutoIncrement {
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
	expandedQuery, expandedArgs, err := wherein.Expand(stmt.query, args)
	if err != nil {
		return 0, err
	}
	result, err := db.Exec(expandedQuery, expandedArgs...)
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

	expandedQuery, expandedArgs, err := wherein.Expand(stmt.query, args)
	if err != nil {
		return 0, err
	}
	sqlRows, err := db.Query(expandedQuery, expandedArgs...)
	if err != nil {
		return 0, err
	}
	defer sqlRows.Close()
	outputs, err := stmt.getOutputs(sqlRows)
	if err != nil {
		return 0, err
	}

	var rowCount = 0
	scanValues := make([]interface{}, len(stmt.columns))

	for sqlRows.Next() {
		rowCount++
		rowValuePtr := reflect.New(rowType)
		rowValue := reflect.Indirect(rowValuePtr)
		var jsonCells []*jsonCell
		for i, col := range outputs {
			cellValue := col.Index.ValueRW(rowValue)
			cellPtr := cellValue.Addr().Interface()
			if col.Tag.JSON {
				jc := newJSONCell(col.Field.Name, cellPtr)
				jsonCells = append(jsonCells, jc)
				scanValues[i] = jc.ScanValue()
			} else if col.Tag.EmptyNull {
				scanValues[i] = newNullCell(col.Field.Name, cellValue, cellPtr)
			} else {
				scanValues[i] = cellPtr
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
	expandedQuery, expandedArgs, err := wherein.Expand(stmt.query, args)
	if err != nil {
		return 0, err
	}
	rows, err := db.Query(expandedQuery, expandedArgs...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	outputs, err := stmt.getOutputs(rows)
	if err != nil {
		return 0, err
	}

	scanValues := make([]interface{}, len(outputs))
	var jsonCells []*jsonCell

	if !rows.Next() {
		// no rows returned
		return 0, nil
	}

	// at least one row returned
	rowCount := 1

	for i, col := range outputs {
		cellValue := col.Index.ValueRW(rowValue)
		cellPtr := cellValue.Addr().Interface()
		if col.Tag.JSON {
			jc := newJSONCell(col.Field.Name, cellPtr)
			jsonCells = append(jsonCells, jc)
			scanValues[i] = jc.ScanValue()
		} else if col.Tag.EmptyNull {
			scanValues[i] = newNullCell(col.Field.Name, cellValue, cellPtr)
		} else {
			scanValues[i] = cellPtr
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

func (stmt *Stmt) getOutputs(rows *sql.Rows) ([]*column.Info, error) {
	stmt.output.mutex.RLock()
	outputs := stmt.output.columns
	stmt.output.mutex.RUnlock()
	if outputs != nil {
		// already worked out
		return outputs, nil
	}
	stmt.output.mutex.Lock()
	defer stmt.output.mutex.Unlock()
	// test again once write lock acquired
	if stmt.output.columns != nil {
		return stmt.output.columns, nil
	}

	columnMap := make(map[string]*column.Info)
	for _, col := range stmt.columns {
		columnName := stmt.convention.ColumnName(col)
		columnMap[columnName] = col
	}

	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	outputs = make([]*column.Info, len(columnNames))
	var columnNotFound = false
	for i, columnName := range columnNames {
		col := columnMap[columnName]
		if col == nil {
			columnNotFound = true
			continue
		}
		outputs[i] = col
		delete(columnMap, columnName)
	}

	if columnNotFound {
		// One or more column names not found. The first loop
		// was case sensitive. Try again case-insensitive.
		// Build a map of lower-case column names for the remaining,
		// unmatched columns and then try again.
		var unknownColumnNames []string
		lowerColumnMap := make(map[string]*column.Info)
		for k, v := range columnMap {
			lowerColumnMap[strings.ToLower(k)] = v
		}
		for i, columnName := range columnNames {
			if outputs[i] != nil {
				continue
			}
			columnNameLower := strings.ToLower(columnName)
			col := lowerColumnMap[columnNameLower]
			if col == nil {
				unknownColumnNames = append(unknownColumnNames, columnName)
				continue
			}
			outputs[i] = col
			delete(lowerColumnMap, columnNameLower)
			delete(columnMap, stmt.convention.ColumnName(col))
		}

		if len(unknownColumnNames) == 1 {
			return nil, fmt.Errorf("unknown column name=%q", unknownColumnNames[0])
		}
		if len(unknownColumnNames) > 0 {
			return nil, fmt.Errorf("unknown columns names=%q", strings.Join(unknownColumnNames, ","))
		}
	}
	if len(columnMap) > 0 {
		missingColumnNames := make([]string, 0, len(columnMap))
		for columnName := range columnMap {
			missingColumnNames = append(missingColumnNames, columnName)
		}
		if len(missingColumnNames) == 1 {
			return nil, fmt.Errorf("missing column name=%q", missingColumnNames[0])
		}
		return nil, fmt.Errorf("missing columns names=%s", strings.Join(missingColumnNames, ","))
	}

	stmt.output.columns = outputs
	return stmt.output.columns, nil
}

func (stmt *Stmt) addColumns(cols columnsT) {
	if cols.clause.isInput() {
		for _, col := range cols.filtered() {
			stmt.inputs = append(stmt.inputs, inputT{col: col})
		}
	}

	// Note that we don't record output columns anymore prior to running the query.
	// This way it does not matter if columns are specified explicitly or
	// if they are expanded out using the {} mechanism.
}

func (stmt *Stmt) scanSQL(query string) error {
	query = strings.TrimSpace(query)
	scan := scanner.New(strings.NewReader(query))
	var counter counterT
	columns := newColumns(stmt.columns)
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
			// TODO(SELECT): should parse the placeholder in case it is positional
			// instead of just allocating it a number assuming it is not positional
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
					buf.WriteString(cols.String(stmt.dialect, stmt.convention, counter.Next))
					stmt.addColumns(cols)
				} else {
					cols, err := columns.Parse(clause, lit)
					if err != nil {
						return fmt.Errorf("cannot expand %q in %q clause: %v", lit, clause, err)
					}
					buf.WriteString(cols.String(stmt.dialect, stmt.convention, counter.Next))
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
			colVal := input.col.Index.ValueRO(rowVal)
			if input.col.Tag.JSON {
				// marshal field contents into JSON and pass as a byte array
				valueRO := colVal.Interface()
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
			} else if input.col.Tag.EmptyNull {
				// TODO: store zero value with the column
				zero := reflect.Zero(colVal.Type()).Interface()
				ival := colVal.Interface()
				if ival == zero {
					args = append(args, nil)
				} else {
					args = append(args, ival)
				}
			} else {
				args = append(args, colVal.Interface())
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

var (
	timeType = reflect.TypeOf(time.Time{})
	timeZero = reflect.Zero(reflect.TypeOf(time.Time{}))
)

// newNullCell returns a scannable value for fields that are configured
// so that a null value means to store an empty value. These fields should
// have a backing field type of int, uint, bool, float or string.
func newNullCell(colname string, cellValue reflect.Value, cellPtr interface{}) interface{} {
	switch cellValue.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return &nullIntCell{colname: colname, cellValue: cellValue}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return &nullUintCell{colname: colname, cellValue: cellValue}
	case reflect.Float32, reflect.Float64:
		return &nullFloatCell{colname: colname, cellValue: cellValue}
	case reflect.Bool:
		return &nullBoolCell{colname: colname, cellValue: cellValue}
	case reflect.String:
		return &nullStringCell{colname: colname, cellValue: cellValue}
	case reflect.Struct:
		if cellValue.Type() == timeType {
			return &nullTimeCell{colname: colname, cellValue: cellValue}
		}
		return cellPtr
	default:
		// other valid types include pointer and slice, which
		// can handle a null value without resorting to reflection
		return cellPtr
	}
}

type nullIntCell struct {
	colname   string
	cellValue reflect.Value
	bits      int
}

func (nc *nullIntCell) Scan(v interface{}) (err error) {
	defer func() {
		// handle panic if SetFloat overflows
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot scan column %q: %v", nc.colname, r)
		}
	}()
	var nullable sql.NullInt64
	if err = nullable.Scan(v); err != nil {
		return fmt.Errorf("cannot scan column %q: %v", nc.colname, err)
	}
	if nullable.Valid {
		nc.cellValue.SetInt(nullable.Int64)
	} else {
		nc.cellValue.SetInt(0)
	}
	return nil
}

type nullUintCell struct {
	colname   string
	cellValue reflect.Value
}

func (nc *nullUintCell) Scan(v interface{}) (err error) {
	defer func() {
		// handle panic if SetFloat overflows
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot scan column %q: %v", nc.colname, r)
		}
	}()
	var nullable sql.NullInt64
	if err = nullable.Scan(v); err != nil {
		return fmt.Errorf("cannot scan column %q: %v", nc.colname, err)
	}
	if nullable.Valid {
		nc.cellValue.SetUint(uint64(nullable.Int64))
	} else {
		nc.cellValue.SetUint(0)
	}
	return nil
}

type nullFloatCell struct {
	colname   string
	cellValue reflect.Value
	bits      int
}

func (nc *nullFloatCell) Scan(v interface{}) (err error) {
	defer func() {
		// handle panic if SetFloat overflows
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot scan column %q: %v", nc.colname, r)
		}
	}()
	var nullable sql.NullFloat64
	if err := nullable.Scan(v); err != nil {
		return fmt.Errorf("cannot scan column %q: %v", nc.colname, err)
	}
	if nullable.Valid {
		nc.cellValue.SetFloat(nullable.Float64)
	} else {
		nc.cellValue.SetFloat(0.0)
	}
	return nil
}

type nullBoolCell struct {
	colname   string
	cellValue reflect.Value
}

func (nc *nullBoolCell) Scan(v interface{}) error {
	var nullable sql.NullBool
	if err := nullable.Scan(v); err != nil {
		return fmt.Errorf("cannot scan column %q: %v", nc.colname, err)
	}
	if nullable.Valid {
		nc.cellValue.SetBool(nullable.Bool)
	} else {
		nc.cellValue.SetBool(false)
	}
	return nil
}

type nullStringCell struct {
	colname   string
	cellValue reflect.Value
}

func (nc *nullStringCell) Scan(v interface{}) error {
	var nullable sql.NullString
	if err := nullable.Scan(v); err != nil {
		return fmt.Errorf("cannot scan column %q: %v", nc.colname, err)
	}
	if nullable.Valid {
		nc.cellValue.SetString(nullable.String)
	} else {
		nc.cellValue.SetString("")
	}
	return nil
}

type nullTimeCell struct {
	colname   string
	cellValue reflect.Value
}

func (nc *nullTimeCell) Scan(v interface{}) error {
	if v == nil {
		nc.cellValue.Set(timeZero)
		return nil
	}
	switch v.(type) {
	case time.Time:
		nc.cellValue.Set(reflect.ValueOf(v))
		return nil
	}

	return fmt.Errorf("cannot scan column %q: type %q is not compatible with time.Time", nc.colname, reflect.TypeOf(v))
}
