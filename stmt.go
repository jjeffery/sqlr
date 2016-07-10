package sqlf

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jjeffery/sqlf/private/column"
	"github.com/jjeffery/sqlf/private/scanner"
)

// commonStmt contains fields and methods common to
// insert, update, delete and select statements.
type commonStmt struct {
	query string
	table *TableInfo
	err   error
}

func (stmt commonStmt) expectedTypeString() string {
	return fmt.Sprintf("%s.%s", stmt.table.rowType.PkgPath(), stmt.table.rowType.Name())
}

func (stmt commonStmt) expectTable() error {
	if stmt.table == nil {
		return errors.New("cannot determine row type from query")
	}
	return nil
}

// execRowStmt contains fields and methods common to insert,
// updated and delete row statements.
type execRowStmt struct {
	commonStmt
	inputs []*column.Info
}

func (stmt execRowStmt) getArgs(row interface{}) ([]interface{}, error) {
	if stmt.table == nil {
		return nil, errors.New("table not specified")
	}
	var args []interface{}

	rowVal := reflect.ValueOf(row)
	for rowVal.Type().Kind() == reflect.Ptr {
		rowVal = rowVal.Elem()
	}
	if rowVal.Type() != stmt.table.rowType {
		return nil, stmt.errorType()
	}

	for _, input := range stmt.inputs {
		args = append(args, input.Index.ValueRO(rowVal).Interface())
	}

	return args, nil
}

func (stmt execRowStmt) doExec(db Execer, row interface{}) (sql.Result, error) {
	if stmt.err != nil {
		return nil, stmt.err
	}
	if err := stmt.expectTable(); err != nil {
		return nil, err
	}
	args, err := stmt.getArgs(row)
	if err != nil {
		return nil, err
	}
	fmt.Printf("query: %s %v\n", stmt.query, args)
	return db.Exec(stmt.query, args...)
}

func (stmt execRowStmt) getRowValue(row interface{}) (reflect.Value, error) {
	rowVal := reflect.ValueOf(row)
	for rowVal.Type().Kind() == reflect.Ptr {
		rowVal = rowVal.Elem()
	}
	if rowVal.Type() != stmt.table.rowType {
		return reflect.Value{}, stmt.errorType()
	}
	return rowVal, nil
}

func (stmt execRowStmt) errorType() error {
	expectedType := stmt.expectedTypeString()
	return fmt.Errorf("expected type %s or *(%s)", expectedType, expectedType)
}

// InsertRowStmt inserts a single row.
type InsertRowStmt struct {
	execRowStmt
}

func (stmt *InsertRowStmt) Exec(db Execer, row interface{}) error {
	// find the auto-increment column, if any
	var autoInc *column.Info
	for _, fi := range stmt.table.fields {
		if fi.AutoIncrement {
			autoInc = fi
			break
		}
	}

	// field for setting the auto-increment value
	var field reflect.Value
	if autoInc != nil {
		// Some DBs allow the auto-increment column to be specified.
		// Work out if this statment is doing this.
		autoIncInserted := false
		for _, ci := range stmt.inputs {
			if ci == autoInc {
				// this statement is setting the auto-increment column explicitly
				autoIncInserted = true
				break
			}
		}

		if !autoIncInserted {
			rowVal := reflect.ValueOf(row)
			field = autoInc.Index.ValueRW(rowVal)
			if !field.CanSet() {
				return fmt.Errorf("cannot set auto-increment value for type %s", rowVal.Type().Name())
			}
		}
	}

	result, err := stmt.doExec(db, row)
	if err != nil {
		return err
	}

	if field.IsValid() {
		n, err := result.LastInsertId()
		if err != nil {
			return nil
		}
		// TODO: could catch a panic here if the type is not int8, 1nt16, int32, int64
		field.SetInt(n)
	}
	return nil
}

// InsertRowPrintf creates a statement that will insert a single row in the database.
// The statement query is constructed using a familiar "printf" style syntax.
func InsertRowPrintf(format string, args ...interface{}) *InsertRowStmt {
	stmt := &InsertRowStmt{}

	for _, arg := range args {
		if tn, ok := arg.(TableName); ok {
			if tn.clause == clauseInsertInto {
				stmt.table = tn.table
			}
		}
		if cols, ok := arg.(Columns); ok {
			if stmt.table == nil && cols.clause == clauseInsertColumns {
				stmt.table = cols.table
			}
			if cols.clause.isInput() {
				// input parameters for the INSERT statement
				stmt.inputs = append(stmt.inputs, cols.filtered()...)
			}
		}
	}

	// generate the SQL statement and scan it
	stmt.query = fmt.Sprintf(format, args...)
	stmt.query = scanSQL(stmt.table.dialect, stmt.query)
	return stmt
}

// UpdateRowStmt updates a single row.
type UpdateRowStmt struct {
	execRowStmt
}

func (stmt *UpdateRowStmt) Exec(db Execer, row interface{}) (int, error) {
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

// UpdateRowPrintf creates a statement that will update a single row in the database.
// The statement query is constructed using a familiar "printf" style syntax.
func UpdateRowPrintf(format string, args ...interface{}) *UpdateRowStmt {
	stmt := &UpdateRowStmt{}

	for _, arg := range args {
		if tn, ok := arg.(TableName); ok {
			if tn.clause.matchAny(clauseUpdateTable, clauseDeleteTable) {
				stmt.table = tn.table
			}
		} else if cols, ok := arg.(Columns); ok {
			if stmt.table == nil && cols.clause.matchAny(clauseUpdateSet, clauseUpdateWhere, clauseDeleteWhere) {
				stmt.table = cols.table
			}
			if cols.clause.isInput() {
				// input parameters for the UPDATE statement
				stmt.inputs = append(stmt.inputs, cols.filtered()...)
			}
		}
	}

	// generate the SQL statement and scan it
	stmt.query = fmt.Sprintf(format, args...)
	stmt.query = scanSQL(stmt.table.dialect, stmt.query)
	return stmt
}

type SelectStmt struct {
	commonStmt
	columns []*column.Info
}

func (stmt *SelectStmt) Select(db Queryer, dest interface{}, args ...interface{}) error {
	if stmt.err != nil {
		return stmt.err
	}
	if err := stmt.expectTable(); err != nil {
		return err
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
	if rowType != stmt.table.rowType {
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
	expectedType := stmt.expectedTypeString()
	return fmt.Errorf("Expected rows to be pointer to []%s or []*%s", expectedType, expectedType)
}

func SelectRowsPrintf(format string, args ...interface{}) *SelectStmt {
	stmt := &SelectStmt{}

	for _, arg := range args {
		if tn, ok := arg.(TableName); ok {
			if stmt.table == nil && tn.clause.matchAny(clauseSelectFrom) {
				stmt.table = tn.table
			}
		} else if cols, ok := arg.(Columns); ok {
			if cols.clause.isInput() {
				stmt.err = errors.New("unexpected arg to SelectRowsPrintf")
				// input parameters for the SELECT statement
			}
			if cols.clause == clauseSelectColumns {
				if stmt.table == nil {
					stmt.table = cols.table
				}
				stmt.columns = append(stmt.columns, cols.filtered()...)
			}
		}
	}

	// generate the SQL statement and scan it
	stmt.query = fmt.Sprintf(format, args...)
	stmt.query = scanSQL(stmt.table.dialect, stmt.query)
	return stmt
}

// Scan the SQL query and convert according to the dialect.
func scanSQL(dialect Dialect, query string) string {
	query = strings.TrimSpace(query)
	scan := scanner.New(strings.NewReader(query))
	placeholderCount := 0
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
			placeholderCount++
			buf.WriteString(dialect.Placeholder(placeholderCount))
		case scanner.IDENT:
			if scanner.IsQuoted(lit) {
				lit = scanner.Unquote(lit)
				buf.WriteString(dialect.Quote(lit))
			} else {
				buf.WriteString(lit)
			}
		}
	}
	return buf.String()
}
