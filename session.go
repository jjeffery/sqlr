package sqlr

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jjeffery/kv"
	"github.com/jjeffery/sqlr/private/wherein"
)

// A Session is a request-scoped database session. It can execute
// queries and it can construct strongly-typed query functions.
type Session struct {
	context context.Context
	cancel  func()
	querier Querier
	schema  *Schema

	// cache of query functions for this session
	queryFuncs map[reflect.Type]reflect.Value

	// map of row handler callback functions
	rowHandlers map[reflect.Type][]func(reflect.Value)
}

// NewSession returns a new, request-scoped session.
//
// Although it is not mandatory, it is a good practice to
// call a session's Close method at the end of a request.
func NewSession(ctx context.Context, querier Querier, schema *Schema) *Session {
	if ctx == nil {
		ctx = context.Background()
	}
	if querier == nil {
		panic("querier cannot be nil")
	}
	if schema == nil {
		panic("schema cannot be nil")
	}
	ctx, cancel := context.WithCancel(ctx)
	return &Session{
		context: ctx,
		cancel:  cancel,
		querier: querier,
		schema:  schema,
	}
}

// Close releases resources associated with the session. Any attempt to
// query using the session will fail after Close has been called.
//
// Because a session is request-scoped, it should never be used once
// a request has completed. Calling a session's Close method at the end
// of a request is an effective way to release resources associated with
// the session and to ensure that it can no longer be used.
//
// Close implements the io.Closer interface. It always returns nil.
func (sess *Session) Close() error {
	sess.cancel()
	sess.queryFuncs = nil
	return nil
}

// Exec executes a query without returning any rows. The args are for any placeholder parameters in the query.
func (sess *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	return sess.execForRow(&struct{}{}, query, args...)
}

// execForRow executes a query on a row without returning any rows. The args are for any placeholder parameters in the query.
//
// Exec is a general-purpose row-based query function. For simple insert and update operations, consider
// using the InsertRow and UpdateRow methods respectively.
func (sess *Session) execForRow(row interface{}, query string, args ...interface{}) (sql.Result, error) {
	stmt, err := sess.schema.Prepare(row, query)
	if err != nil {
		return nil, err
	}
	return stmt.exec(sess.context, sess.querier, row, args...)
}

// InsertRow inserts one row into the database.
//
// If the row has an auto-increment field, then that field is updated
// with the value of the auto-increment column.
func (sess *Session) InsertRow(row interface{}) error {
	tbl := sess.schema.TableFor(row)

	// if we are going to update any fields, make sure we have a pointer
	if tbl.createdAt != nil || tbl.updatedAt != nil || tbl.version != nil || tbl.autoincr != nil {
		// We will want to modify row, so check that it can be modified.
		// Unfortunately this is a runtime check and cannot be determined at compile time.
		// TODO(jpj): considered creating MakeInsert and MakeUpdate functions similar to
		// MakeQuery, and these would have stricter type checking. The thinking is that would
		// create too much additional work and ceremony. This API will probably work just fine.
		rowValue := tbl.mustGetRowValue(row)
		if !rowValue.CanAddr() {
			var names []string
			if tbl.autoincr != nil {
				names = append(names, tbl.autoincr.info.FieldNames)
			}
			if tbl.createdAt != nil {
				names = append(names, tbl.createdAt.info.FieldNames)
			}
			if tbl.updatedAt != nil {
				names = append(names, tbl.updatedAt.info.FieldNames)
			}
			if tbl.version != nil {
				names = append(names, tbl.version.info.FieldNames)
			}
			var msg string
			if len(names) == 1 {
				msg = fmt.Sprintf("InsertRow requires *%s to update field %s", tbl.rowType, names[0])
			} else {
				msg = fmt.Sprintf("InsertRow requires *%s to update fields %s", tbl.rowType, strings.Join(names, ", "))
			}
			return errors.New(msg)
		}

		// Set the CreatedAt, UpdatedAt values of the field.
		// TODO(jpj): should probably put back to previous values if insert is unsuccessful
		if tbl.createdAt != nil || tbl.updatedAt != nil {
			now := time.Now()
			nowValue := reflect.ValueOf(now)
			if tbl.createdAt != nil {
				createdAtValue := tbl.createdAt.info.Index.ValueRW(rowValue)
				createdAtValue.Set(nowValue)
			}
			if tbl.updatedAt != nil {
				updatedAtValue := tbl.updatedAt.info.Index.ValueRW(rowValue)
				updatedAtValue.Set(nowValue)
			}
		}

		if tbl.version != nil {
			versionValue := tbl.version.info.Index.ValueRW(rowValue)
			versionValue.SetInt(1)
		}

		if tbl.autoincr != nil {
			if isPostgres(sess.schema.dialect) {
				if err := sess.postgresInsertRow(row, tbl, rowValue); err != nil {
					return err
				}
				// success = true
				return nil
			} else {
				if err := sess.autoincrInsertRow(row, tbl, rowValue); err != nil {
					return err
				}
				// success = true
				return nil
			}
		}
	}

	// no autoincr column, so just a standard insert
	query := fmt.Sprintf("insert into %s({}) values({})", sess.schema.dialect.Quote(tbl.tableName))
	stmt, err := sess.schema.Prepare(row, query)
	if err != nil {
		return err
	}
	_, err = stmt.exec(sess.context, sess.querier, row)
	if err != nil {
		return tbl.wrapRowError(err, row, "cannot insert row")
	}
	// success = true
	return nil
}

func (sess *Session) autoincrInsertRow(row interface{}, tbl *Table, rowValue reflect.Value) error {
	query := fmt.Sprintf("insert into %s({}) values({})", sess.schema.dialect.Quote(tbl.tableName))
	stmt, err := sess.schema.Prepare(row, query)
	if err != nil {
		return err
	}
	result, err := stmt.exec(sess.context, sess.querier, row)
	if err != nil {
		return tbl.wrapRowError(err, row, "cannot insert row")
	}
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return tbl.wrapRowError(err, row, "cannot retrieve last insert id")
	}
	// already checked previously that this field can be set
	field := tbl.autoincr.info.Index.ValueRW(rowValue)
	field.SetInt(lastInsertID)
	return nil
}

func (sess *Session) postgresInsertRow(row interface{}, tbl *Table, rowValue reflect.Value) error {
	query := fmt.Sprintf(
		"insert into %s({}) values({}) returning %s",
		sess.schema.dialect.Quote(tbl.tableName),
		sess.schema.dialect.Quote(tbl.autoincr.columnName),
	)
	stmt, err := sess.schema.Prepare(row, query)
	if err != nil {
		return err
	}
	args, err := stmt.getArgs(row, nil)
	if err != nil {
		return err
	}
	rows, err := sess.querier.QueryContext(sess.context, stmt.String(), args...)
	if err != nil {
		return tbl.wrapRowError(err, row, "cannot insert row")
	}
	defer rows.Close()
	// expecting one row, one column
	rows.Next()
	var lastInsertID int64
	if err := rows.Scan(&lastInsertID); err != nil {
		return tbl.wrapRowError(err, row, "cannot retrieve last insert id")
	}
	// already checked previously that this field can be set
	field := tbl.autoincr.info.Index.ValueRW(rowValue)
	field.SetInt(lastInsertID)
	return nil
}

// UpdateRow updates one row in the database. It returns the number
// of rows updated, which should be zero or one.
//
// If the row has a version field, then that field is incremented
// during the update. If the row being updated does not match the
// original value of the version field, then an OptimisticLockingError
// will be returned.
func (sess *Session) UpdateRow(row interface{}) (int, error) {
	tbl := sess.schema.TableFor(row)

	// if we are going to update any fields, make sure we have a pointer
	if tbl.updatedAt != nil || tbl.version != nil {
		// We will want to modify row, so check that it can be modified.
		// Unfortunately this is a runtime check and cannot be determined at compile time.
		// TODO(jpj): considered creating MakeInsert and MakeUpdate functions similar to
		// MakeQuery, and these would have stricter type checking. The thinking is that would
		// create too much additional work and ceremony. This API will probably work just fine.
		rowValue := tbl.mustGetRowValue(row)
		if !rowValue.CanAddr() {
			var names []string
			if tbl.version != nil {
				names = append(names, tbl.version.info.FieldNames)
			}
			if tbl.updatedAt != nil {
				names = append(names, tbl.updatedAt.info.FieldNames)
			}
			var msg string
			if len(names) == 1 {
				msg = fmt.Sprintf("UpdateRow requires *%s to update field %s", tbl.rowType, names[0])
			} else {
				msg = fmt.Sprintf("UpdateRow requires *%s to update fields %s", tbl.rowType, strings.Join(names, ", "))
			}
			return 0, errors.New(msg)
		}

		if tbl.updatedAt != nil {
			// Set the UpdatedAt value of the field.
			// TODO(jpj): should probably put back to previous values if insert is unsuccessful
			now := time.Now()
			nowValue := reflect.ValueOf(now)
			updatedAtValue := tbl.updatedAt.info.Index.ValueRW(rowValue)
			updatedAtValue.Set(nowValue)
		}

		if tbl.version != nil {
			n, err := sess.updateRowVersioned(row, tbl, rowValue)
			if err != nil {
				return 0, err
			}
			// success = true
			return n, nil
		}
	}

	// no version column, so just a standard update
	query := fmt.Sprintf("update %s set {} where {}", sess.schema.dialect.Quote(tbl.tableName))
	stmt, err := sess.schema.Prepare(row, query)
	if err != nil {
		return 0, err
	}
	result, err := stmt.exec(sess.context, sess.querier, row)
	if err != nil {
		return 0, tbl.wrapRowError(err, row, "cannot update row")
	}
	rowsUpdated, err := result.RowsAffected()
	if err != nil {
		return 0, tbl.wrapRowError(err, row, "cannot retrieve rows updated")
	}
	// success = true
	return int(rowsUpdated), nil
}

func (sess *Session) updateRowVersioned(row interface{}, tbl *Table, rowValue reflect.Value) (int, error) {
	versionValue := tbl.version.info.Index.ValueRW(rowValue)
	oldVersion := versionValue.Int()
	newVersion := oldVersion + 1
	versionValue.SetInt(newVersion)
	var success bool

	// rollback to old version if not successful
	defer func() {
		if !success {
			versionValue.SetInt(oldVersion)
		}
	}()

	dialect := sess.schema.dialect
	query := fmt.Sprintf(
		"update %s set {} where {} and %s = ?",
		dialect.Quote(tbl.tableName),
		dialect.Quote(tbl.version.columnName),
	)
	stmt, err := sess.schema.Prepare(row, query)
	if err != nil {
		return 0, err
	}

	result, err := stmt.exec(sess.context, sess.querier, row, oldVersion)
	if err != nil {
		return 0, tbl.wrapRowError(err, row, "cannot update row")
	}

	rowsUpdated, err := result.RowsAffected()
	if err != nil {
		return 0, tbl.wrapRowError(err, row, "cannot get rows updated")
	}

	if rowsUpdated == 1 {
		success = true
		return 1, nil
	}

	if rowsUpdated > 1 {
		return 0, tbl.wrapRowError(err, row, "unexpected rows updated count").With(
			"rowsUpdated", rowsUpdated,
		)
	}

	// At this point no rows were updated, which indicates an optimistic locking error.
	query = fmt.Sprintf(
		"select %s from %s",
		dialect.Quote(tbl.version.columnName),
		dialect.Quote(tbl.tableName),
	)
	var args []interface{}
	for i, pkcol := range tbl.pk {
		if i == 0 {
			query += fmt.Sprintf(" where %s = %s", dialect.Quote(pkcol.columnName), dialect.Placeholder(i+1))
		} else {
			query += fmt.Sprintf(" and %s = %s", dialect.Quote(pkcol.columnName), dialect.Placeholder(i+1))
		}
		pkValue := pkcol.info.Index.ValueRO(rowValue)
		args = append(args, pkValue.Interface())
	}
	rows, err := sess.querier.QueryContext(sess.context, query, args...)
	if err != nil {
		return 0, tbl.wrapRowError(err, row, "cannot obtain version")
	}
	var currentVersion int64
	rows.Next()
	if err := rows.Scan(&currentVersion); err != nil {
		return 0, tbl.wrapRowError(err, row, "cannot scan version")
	}

	return 0, &OptimisticLockingError{
		Table:           tbl,
		Row:             row,
		ExpectedVersion: oldVersion,
		ActualVersion:   currentVersion,
	}
}

// OptimisticLockingError is an error generated during an Update
// or Upsert operation, where the value of the row struct version
// field does not match the value of the corresponding row in the
// database.
type OptimisticLockingError struct {
	Table           *Table
	Row             interface{}
	ExpectedVersion int64
	ActualVersion   int64
}

func (e *OptimisticLockingError) Error() string {
	var keyvals []interface{}
	if e != nil {
		if e.Table != nil && e.Row != nil {
			keyvals = e.Table.keyvals(e.Row)
		}
		if e.ExpectedVersion != e.ActualVersion {
			keyvals = append(keyvals, "expectedVersion", e.ExpectedVersion)
			keyvals = append(keyvals, "actualVersion", e.ActualVersion)
		}
	}
	return "optimistic locking conflict " + kv.List(keyvals).String()
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
func (sess *Session) Select(rows interface{}, query string, args ...interface{}) (int, error) {
	stmt, err := sess.schema.Prepare(rows, query)
	if err != nil {
		return 0, err
	}
	n, err := stmt.selectRows(sess.context, sess.querier, rows, args...)
	if err != nil {
		return n, err
	}
	if n > 0 {
		// TODO: perform notifications
		sess.callRowHandlers(stmt.tbl, rows, n)
	}
	return n, err
}

// Querier returns the database querier associated with this session.
func (sess *Session) Querier() Querier {
	return sess.querier
}

// Context returns the context associated with this session.
//
// Note that the context returned by this function is not identical to the
// context passed to the NewSession function. The context returned by this
// method is canceled when the Close method is called.
func (sess *Session) Context() context.Context {
	return sess.context
}

// Schema returns the schema used for this session.
func (sess *Session) Schema() *Schema {
	return sess.schema
}

// MakeQuery makes one or more functions that can be used to query in a type-safe manner.
// Each funcPtr is a pointer to a function that will be created by this function.
//
// If "Row" is the row type, and "RowID" is the primary key type for Row objects, then the
// function can be any one of the following signatures:
//  // Get one row based on its ID
//  func(id RowID) (*Row, error)
//
//  // Get multiple rows given multiple IDs
//  func(ids []RowID) ([]*Row, error)
//  func(ids ...RowID) ([]*Row, error)
//
//  // Get one row returning a thunk: batches multiple requests into
//  // one query using the dataloader pattern
//  func(id RowID) func() (*Row, error)
//
//  // Execute a query that will return multiple Row objects
//  func(query string, args ...interface{}) ([]*Row, error)
//
//  // Execute a query that will return a single Row object.
//  func(query string, args ...interface{}) (*Row, error)
//
//  // Execute a query that will return a single integer value.
//  // Useful for select count(*) queries.
//  func(query string, args ...interface{}) (int, error)
//  func(query string, args ...interface{}) (int64, error)
//
// If any of the funcPtr arguments are not pointers to a function, or do not fit
// one of the known function prototypes, then this function will panic.
func (sess *Session) MakeQuery(funcPtr ...interface{}) {
	if err := sess.makeQueries(funcPtr...); err != nil {
		panic(err)
	}
}

func (sess *Session) makeQueries(funcPtr ...interface{}) error {
	for _, fp := range funcPtr {
		if err := sess.makeQueryFunc(fp); err != nil {
			return err
		}
	}
	return nil
}

func (sess *Session) makeQueryFunc(funcPtr interface{}) error {
	funcPtrValue := reflect.ValueOf(funcPtr)
	funcPtrType := funcPtrValue.Type()
	if funcPtrType.Kind() != reflect.Ptr {
		return newError("expected pointer to function, got %s", funcPtrType.String())
	}
	funcValue := funcPtrValue.Elem()
	funcType := funcValue.Type()
	if funcType.Kind() != reflect.Func {
		return newError("expected pointer to function, got %s", funcPtrType.String())
	}

	// lookup the cache, and if a miss then create func and add to cache
	queryFunc, ok := sess.queryFuncs[funcType]
	if !ok {
		queryFuncFactory, err := makeQuery(funcType, sess.schema)
		if err != nil {
			return err
		}
		queryFunc = queryFuncFactory(sess)
		if sess.queryFuncs == nil {
			sess.queryFuncs = make(map[reflect.Type]reflect.Value)
		}
		sess.queryFuncs[funcType] = queryFunc
	}
	funcValue.Set(queryFunc)
	return nil
}

// Query performs a query that is not row-based. The query placeholders
// are converted to the format suitable for the SQL dialect, and any
// args that are slices are expanded (eg for WHERE IN (...) clauses).
func (sess *Session) Query(query string, args ...interface{}) (*sql.Rows, error) {
	var row struct{}
	stmt, err := sess.schema.Prepare(row, query)
	if err != nil {
		return nil, err
	}
	args, err = stmt.getArgs(row, args)
	if err != nil {
		return nil, err
	}
	expandedQuery, expandedArgs, err := wherein.Expand(stmt.query, args)
	if err != nil {
		return nil, err
	}
	return sess.querier.QueryContext(sess.context, expandedQuery, expandedArgs...)
}

// callRowHandlers calls the appropriate row handlers
func (sess *Session) callRowHandlers(tbl *Table, rows interface{}, rowCount int) {
	handlers := sess.rowHandlers[tbl.rowType]
	if len(handlers) == 0 {
		// nothing to do
		return
	}
	rowsPtrValue := reflect.ValueOf(rows)
	rowsValue := rowsPtrValue.Elem()
	if rowsValue.Kind() == reflect.Slice {
		rowsElemType := rowsValue.Type().Elem()
		if rowsElemType.Kind() == reflect.Ptr {
			// this is what we are after
			for _, callback := range handlers {
				callback(rowsValue)
			}
		}
		// This is where the result set is an array of struct, not
		// pointer to struct. It is not necessary possible to get
		// a slice of pointers to row structs without a lot of
		// copying. For now, the callbacks just do not get called,
		// and we'll see if that is a problem.
		return
	}

	if rowsValue.Kind() == reflect.Struct {
		// A single row has been returned
		sliceValue := reflect.MakeSlice(reflect.SliceOf(rowsPtrValue.Type()), 1, 1)
		sliceElemValue := sliceValue.Index(0)
		sliceElemValue.Set(rowsPtrValue)
		for _, callback := range handlers {
			callback(sliceValue)
		}
	}
}

// HandleRows supplies a callback function for the session to call when
// one or more rows are retrieved during a select query. The callback must
// have a function signature like the following:
//  func(rows []*RowType)
// Where RowType is a structure that represents a row. Whenever a select
// query is called that returns one or more RowType instances, then this
// callback will be called.
//
// If callback is not a function as described above, this method will panic.
func (sess *Session) HandleRows(callback interface{}) {
	if err := sess.handleRowsHelper(callback); err != nil {
		panic(err)
	}
}

func (sess *Session) handleRowsHelper(callback interface{}) error {
	cbValue := reflect.ValueOf(callback)
	if cbValue.Kind() != reflect.Func {
		return errors.New("HandleRows: expect callback to be func([]*RowStruct)")
	}
	cbType := cbValue.Type()
	if cbType.NumIn() != 1 || cbType.NumOut() != 0 {
		return errors.New("HandleRows: expect callback to be func([]*RowStruct)")
	}
	inSliceType := cbType.In(0)
	if inSliceType.Kind() != reflect.Slice {
		return errors.New("HandleRows: expect callback to be func([]*RowStruct)")
	}
	inPtrType := inSliceType.Elem()
	if inPtrType.Kind() != reflect.Ptr {
		return errors.New("HandleRows: expect callback to be func([]*RowStruct)")
	}
	inStructType := inPtrType.Elem()
	if inStructType.Kind() != reflect.Struct {
		return errors.New("HandleRows: expect callback to be func([]*RowStruct)")
	}

	// checks finished: callback is the correct type

	if sess.rowHandlers == nil {
		sess.rowHandlers = make(map[reflect.Type][]func(reflect.Value))
	}
	handlers := sess.rowHandlers[inStructType]
	handlers = append(handlers, func(rows reflect.Value) {
		inValues := [1]reflect.Value{rows}
		cbValue.Call(inValues[:])
	})
	sess.rowHandlers[inStructType] = handlers
	return nil
}

// SessionRow is is used to perform queries that apply to single row.
type SessionRow struct {
	Session *Session
	Row     interface{}
}

// Row returns a session row, which is then used to execute a query
// based on the contents of row.
func (sess *Session) Row(row interface{}) *SessionRow {
	return &SessionRow{
		Session: sess,
		Row:     row,
	}
}

// Exec executes a query using the row as parameters to the query.
func (row *SessionRow) Exec(query string, args ...interface{}) (sql.Result, error) {
	return row.Session.execForRow(row.Row, query, args...)
}

/**** TODO(jpj): decide whether we want to include these methods.
	We want the Exec method because it is more consistent to have
	every Exec() method accept the same arguments (query, args).
	There is no real need for these methods, however.
	For now keep the API smaller by leaving them out, and add them
	if there is a good reason in the future.

// Insert inserts the row into the database.
func (row *SessionRow) Insert() error {
	return row.Session.InsertRow(row.Row)
}

// Update updates the row in the database.
func (row *SessionRow) Update() (int, error) {
	return row.Session.UpdateRow(row.Row)
}

// Delete deletes the row from the database.
func (row *SessionRow) Delete() (int, error) {
	return 0, errors.New("not implemented")
}

****/
