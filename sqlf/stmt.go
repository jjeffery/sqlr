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

var errNotImplemented = errors.New("not implemented")

type InsertRowStmt struct {
	commonStmt
	autoIncrColumn *column.Info
}

func MustPrepareInsertRow(row interface{}, sql string) *InsertRowStmt {
	stmt, err := PrepareInsertRow(row, sql)
	if err != nil {
		panic(err)
	}
	return stmt
}

func prepareInsertRow(schema *Schema, row interface{}, sql string) (*InsertRowStmt, error) {
	stmt := &InsertRowStmt{}
	if err := stmt.prepareExecRow(schema, row, sql); err != nil {
		return nil, err
	}

	for _, col := range stmt.columns {
		if col.AutoIncrement {
			stmt.autoIncrColumn = col
			break
		}
	}

	if stmt.autoIncrColumn != nil {
		// Some DBs allow the auto-increment column to be specified.
		// Work out if this statment is doing this.
		for _, col := range stmt.inputs {
			if col == stmt.autoIncrColumn {
				// this statement is setting the auto-increment column explicitly
				stmt.autoIncrColumn = nil
				break
			}
		}
	}

	return stmt, nil
}

func PrepareInsertRow(row interface{}, sql string) (*InsertRowStmt, error) {
	return prepareInsertRow(DefaultSchema, row, sql)
}

func (stmt *InsertRowStmt) Exec(db Execer, row interface{}) error {

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
			return nil
		}
		// TODO: could catch a panic here if the type is not int8, 1nt16, int32, int64
		field.SetInt(n)
	}
	return nil
}

type UpdateRowStmt struct {
	commonStmt
}

func MustPrepareUpdateRow(row interface{}, sql string) *UpdateRowStmt {
	stmt, err := PrepareUpdateRow(row, sql)
	if err != nil {
		panic(err)
	}
	return stmt
}

func prepareUpdateRow(schema *Schema, row interface{}, sql string) (*UpdateRowStmt, error) {
	stmt := &UpdateRowStmt{}
	if err := stmt.prepareExecRow(schema, row, sql); err != nil {
		return nil, err
	}
	return stmt, nil
}

func PrepareUpdateRow(row interface{}, sql string) (*UpdateRowStmt, error) {
	return prepareUpdateRow(DefaultSchema, row, sql)
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

type GetRowStmt struct {
	commonStmt
}

func MustPrepareGetRow(row interface{}, sql string) *GetRowStmt {
	stmt, err := PrepareGetRow(row, sql)
	if err != nil {
		panic(err)
	}
	return stmt
}

func PrepareGetRow(row interface{}, sql string) (*GetRowStmt, error) {
	stmt := &GetRowStmt{}
	if err := stmt.prepareCommon(DefaultSchema, row, sql); err != nil {
		return nil, err
	}
	return stmt, nil
}

func prepareGetRow(schema *Schema, row interface{}, sql string) (*GetRowStmt, error) {
	stmt := &GetRowStmt{}
	if err := stmt.prepareCommon(schema, row, sql); err != nil {
		return nil, err
	}
	return stmt, nil
}

func (stmt *GetRowStmt) errorPtrType() error {
	expectedType := stmt.expectedTypeString()
	return fmt.Errorf("expected dest to be *%s", expectedType)
}

func (stmt *GetRowStmt) Get(db Queryer, dest interface{}) (int, error) {
	destValue := reflect.ValueOf(dest)

	if destValue.Kind() != reflect.Ptr {
		return 0, stmt.errorPtrType()
	}
	if destValue.IsNil() {
		return 0, errors.New("nil pointer passed")
	}

	rowValue := reflect.Indirect(destValue)
	rowType := rowValue.Type()
	if rowType != stmt.rowType {
		return 0, stmt.errorPtrType()
	}

	args, err := stmt.getArgs(dest)
	if err != nil {
		return 0, nil
	}

	stmt.Printf("query=%q, args=%v\n", stmt.query, args)
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

type SelectStmt struct {
	commonStmt
}

func MustPrepareSelect(row interface{}, sql string) *SelectStmt {
	stmt, err := PrepareSelect(row, sql)
	if err != nil {
		panic(err)
	}
	return stmt
}

func prepareSelect(schema *Schema, row interface{}, sql string) (*SelectStmt, error) {
	stmt := &SelectStmt{}
	if err := stmt.prepareCommon(schema, row, sql); err != nil {
		return nil, err
	}
	if len(stmt.inputs) > 0 {
		return nil, errors.New("unexpected inputs in query")
	}
	return stmt, nil
}

func PrepareSelect(row interface{}, sql string) (*SelectStmt, error) {
	return prepareSelect(DefaultSchema, row, sql)
}

func (stmt *SelectStmt) Select(db Queryer, dest interface{}, args ...interface{}) error {
	return errNotImplemented
}

type commonStmt struct {
	rowType    reflect.Type
	query      string
	dialect    Dialect
	convention Convention
	logger     Logger
	columns    []*column.Info
	inputs     []*column.Info
	outputs    []*column.Info
	err        error
}

// String prints the SQL query associated with the statement.
func (stmt *commonStmt) String() string {
	return stmt.query
}

func (stmt *commonStmt) Printf(format string, args ...interface{}) {
	if stmt.logger != nil {
		stmt.logger.Printf(format, args...)
	}
}

func (stmt *commonStmt) prepareCommon(schema *Schema, row interface{}, sql string) error {
	stmt.rowType = reflect.TypeOf(row)
	if stmt.rowType.Kind() == reflect.Ptr {
		stmt.rowType = stmt.rowType.Elem()
	}
	stmt.columns = column.ListForType(stmt.rowType)
	stmt.convention = schema.convention()
	stmt.dialect = schema.dialect()
	stmt.logger = schema.Logger
	if err := stmt.scanSQL(sql); err != nil {
		return err
	}
	stmt.Printf("prepared=%q", stmt.query)
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

func (stmt *commonStmt) addColumns(cols Columns) {
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

func (stmt *commonStmt) scanSQL(query string) error {
	query = strings.TrimSpace(query)
	scan := scanner.New(strings.NewReader(query))
	columns := newColumns(stmt.columns, stmt.convention, stmt.dialect)
	var insertColumns *Columns
	placeholderCount := 0
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
			placeholderCount++
			buf.WriteString(stmt.dialect.Placeholder(placeholderCount))
		case scanner.IDENT:
			if lit[0] == '{' {
				if !clause.acceptsColumns() {
					// invalid place to insert columns
					return fmt.Errorf("cannot expand %q in %q clause", lit, clause)
				}
				lit = strings.TrimSpace(scanner.Unquote(lit))
				if clause == clauseInsertValues {
					if lit != "" {
						return fmt.Errorf("columns for %q clause always match the %q clause",
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
			}
			if scanner.IsQuoted(lit) {
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

func (stmt commonStmt) doExec(db Execer, row interface{}) (sql.Result, error) {
	args, err := stmt.getArgs(row)
	if err != nil {
		return nil, err
	}
	stmt.Printf("query=%q, args=%v\n", stmt.query, args)
	return db.Exec(stmt.query, args...)
}

func (stmt commonStmt) getArgs(row interface{}) ([]interface{}, error) {
	var args []interface{}

	rowVal := reflect.ValueOf(row)
	for rowVal.Type().Kind() == reflect.Ptr {
		rowVal = rowVal.Elem()
	}
	if rowVal.Type() != stmt.rowType {
		return nil, stmt.errorType()
	}

	for _, input := range stmt.inputs {
		args = append(args, input.Index.ValueRO(rowVal).Interface())
	}

	return args, nil
}

func (stmt commonStmt) errorType() error {
	expectedType := stmt.expectedTypeString()
	return fmt.Errorf("expected type %s or *(%s)", expectedType, expectedType)
}

func (stmt commonStmt) expectedTypeString() string {
	return fmt.Sprintf("%s.%s", stmt.rowType.PkgPath(), stmt.rowType.Name())
}
