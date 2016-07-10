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

type execRowStmt struct {
	query  string
	table  *TableInfo
	inputs []*column.Info
	err    error
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
	expectedType := fmt.Sprintf("%s.%s", stmt.table.rowType.PkgPath(), stmt.table.rowType.Name())
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
			if tn.clause == clauseUpdateTable {
				stmt.table = tn.table
			}
		}
		if cols, ok := arg.(Columns); ok {
			if stmt.table == nil && (cols.clause == clauseUpdateSet || cols.clause == clauseUpdateWhere) {
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
