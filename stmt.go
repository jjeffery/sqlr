package sqlr

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/jjeffery/sqlr/private/scanner"
	"github.com/jjeffery/sqlr/private/wherein"
)

// Stmt is a prepared statement. A Stmt is safe for concurrent use by multiple goroutines.
//
// Stmt is important for the implementation, but currently does not export many public methods.
// Currently the only public operation is to print the SQL. It may be removed from the public API
// in a future version.
type Stmt struct {
	schema    *Schema
	tbl       *Table
	queryType queryType
	query     string
	dialect   Dialect
	inputs    []inputSource
	argCount  int      // the number of args expected in addition to fields from the row
	output    struct { // outputs from a select query are determined the first time it is run
		mutex   sync.RWMutex
		columns []*Column
	}
	autoIncrColumn *Column
}

// inputSource describes where to source the input to an SQL query. (There is
// one input for each placeholder in the query).
//
// If col is non-nil, then the input should be sourced from the field
// associated with the column.
//
// If col is nil, then argIndex is the index into the args array, and the
// corresponding arg should be used as input.
type inputSource struct {
	col      *Column
	argIndex int // used only if col == nil
}

// newStmt creates a new statement for the schema, table and query.
func newStmt(schema *Schema, tbl *Table, sql string) (*Stmt, error) {
	stmt := &Stmt{
		schema:  schema,
		dialect: schema.getDialect(),
		tbl:     tbl,
	}
	if err := stmt.scanSQL(sql); err != nil {
		return nil, err
	}

	if stmt.queryType == queryInsert {
		for _, col := range tbl.Columns() {
			if col.AutoIncrement() {
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

func (stmt *Stmt) exec(ctx context.Context, db Querier, row interface{}, args ...interface{}) (sql.Result, error) {
	args, err := stmt.getArgs(row, args)
	if err != nil {
		return nil, err
	}
	expandedQuery, expandedArgs, err := wherein.Expand(stmt.query, args)
	if err != nil {
		return nil, err
	}
	result, err := db.ExecContext(ctx, expandedQuery, expandedArgs...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// selectRows executes the prepared query statement with the given arguments and
// returns the query results in rows. If rows is a pointer to a slice of structs
// then one item is added to the slice for each row returned by the query. If row
// is a pointer to a struct then that struct is filled with the result of the first
// row returned by the query. In both cases Select returns the number of rows returned
// by the query.
//
// This used to be a public method, but has been deprecated in favour of Session.Select.
func (stmt *Stmt) selectRows(ctx context.Context, db Querier, rows interface{}, args ...interface{}) (int, error) {
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
	if destType == stmt.tbl.RowType() {
		// pointer to row struct, so only fetch one row
		return stmt.selectOne(ctx, db, rows, destValue, args)
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
	if rowType != stmt.tbl.RowType() {
		return 0, errorPtrType()
	}

	expandedQuery, expandedArgs, err := wherein.Expand(stmt.query, args)
	if err != nil {
		return 0, err
	}
	sqlRows, err := db.QueryContext(ctx, expandedQuery, expandedArgs...)
	if err != nil {
		return 0, err
	}
	defer sqlRows.Close()
	outputs, err := stmt.getOutputs(sqlRows)
	if err != nil {
		return 0, err
	}

	var rowCount = 0
	scanValues := make([]interface{}, len(stmt.tbl.Columns()))

	for sqlRows.Next() {
		rowCount++
		rowValuePtr := reflect.New(rowType)
		rowValue := reflect.Indirect(rowValuePtr)
		var jsonCells []*jsonCell
		for i, col := range outputs {
			cellValue := col.info.Index.ValueRW(rowValue)
			cellPtr := cellValue.Addr().Interface()
			if col.JSON() {
				jc := newJSONCell(col.info.Field.Name, cellPtr)
				jsonCells = append(jsonCells, jc)
				scanValues[i] = jc.ScanValue()
			} else {
				scanValues[i] = newNullCell(col.info.Field.Name, cellValue, cellPtr)
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

	if err := sqlRows.Err(); err != nil {
		return 0, err
	}

	// If the slice is nil, return an empty slice. This way the returned slice is
	// always non-nil for a successful call.
	if sliceValue.IsNil() {
		if isPtr {
			sliceValue.Set(reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(rowType)), 0, 0))
		} else {
			sliceValue.Set(reflect.MakeSlice(reflect.SliceOf(rowType), 0, 0))
		}
	}

	return rowCount, nil
}

// TODO(jpj): need to merge the common code in Select and selectOne

func (stmt *Stmt) selectOne(ctx context.Context, db Querier, dest interface{}, rowValue reflect.Value, args []interface{}) (int, error) {
	expandedQuery, expandedArgs, err := wherein.Expand(stmt.query, args)
	if err != nil {
		return 0, err
	}
	rows, err := db.QueryContext(ctx, expandedQuery, expandedArgs...)
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
		cellValue := col.info.Index.ValueRW(rowValue)
		cellPtr := cellValue.Addr().Interface()
		if col.JSON() {
			jc := newJSONCell(col.info.Field.Name, cellPtr)
			jsonCells = append(jsonCells, jc)
			scanValues[i] = jc.ScanValue()
		} else {
			scanValues[i] = newNullCell(col.info.Field.Name, cellValue, cellPtr)
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

func (stmt *Stmt) getOutputs(rows *sql.Rows) ([]*Column, error) {
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

	columnMap := make(map[string]*Column)
	for _, col := range stmt.tbl.Columns() {
		columnMap[col.Name()] = col
	}

	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	outputs = make([]*Column, len(columnNames))
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
		lowerColumnMap := make(map[string]*Column)
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
			delete(columnMap, col.Name())
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

func (stmt *Stmt) scanSQL(query string) error {
	query = strings.TrimSpace(query)
	scan := scanner.New(strings.NewReader(query))
	columns := newColumns(stmt.tbl.Columns())
	var counter int
	counterNext := func() int { counter++; return counter }
	var insertColumns *columnList
	var clause sqlClause
	var buf bytes.Buffer
	rename := func(name string) string {
		if newName, ok := stmt.schema.renameIdent(name); ok {
			return newName
		}
		return name
	}

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
			// TODO(jpj): should parse the placeholder in case it is positional
			// instead of just allocating it a number assuming it is not positional
			buf.WriteString(stmt.dialect.Placeholder(counterNext()))
			stmt.inputs = append(stmt.inputs, inputSource{argIndex: stmt.argCount})
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
					buf.WriteString(cols.String(stmt.dialect, counterNext))
					stmt.addInputColumns(cols)
				} else {
					cols, err := columns.Parse(clause, lit)
					if err != nil {
						return fmt.Errorf("cannot expand %q in %q clause: %v", lit, clause, err)
					}
					buf.WriteString(cols.String(stmt.dialect, counterNext))
					stmt.addInputColumns(cols)
					if clause == clauseInsertColumns {
						insertColumns = &cols
					}
				}
			} else if scanner.IsQuoted(lit) {
				lit = rename(scanner.Unquote(lit))
				buf.WriteString(stmt.dialect.Quote(lit))
			} else {
				lit = rename(lit)
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

func (stmt *Stmt) addInputColumns(cols columnList) {
	if cols.clause.isInput() {
		for _, col := range cols.filtered() {
			stmt.inputs = append(stmt.inputs, inputSource{col: col})
		}
	}
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
	if rowVal.Type() != stmt.tbl.RowType() {
		// should never happen, calling functions have already checked
		expectedType := stmt.expectedTypeName()
		return nil, fmt.Errorf("expected type %s or *(%s)", expectedType, expectedType)
	}

	for _, input := range stmt.inputs {
		if input.col != nil {
			colVal := input.col.info.Index.ValueRO(rowVal)
			if input.col.JSON() {
				// marshal field contents into JSON and pass as a byte array
				valueRO := colVal.Interface()
				if input.col.EmptyNull() && reflect.DeepEqual(valueRO, input.col.zeroValue) {
					args = append(args, nil)
				} else if valueRO == nil {
					args = append(args, nil)
				} else {
					data, err := json.Marshal(valueRO)
					if err != nil {
						// TODO(jpj): if errors.Wrap makes it into the stdlib, use it here
						err = fmt.Errorf("cannot marshal field %q: %v", input.col.info.Field.Name, err)
						return nil, err
					}
					args = append(args, data)
				}
			} else if input.col.EmptyNull() {
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
	rowType := stmt.tbl.RowType()
	return fmt.Sprintf("%s.%s", rowType.PkgPath(), rowType.Name())
}
