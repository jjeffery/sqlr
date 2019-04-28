package sqlr

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func TestDB1(t *testing.T) {
	ctx := context.Background()
	db := sqliteDB(t)
	defer db.Close()

	mustExec(t, db, `
		create table test_table(
			id integer primary key autoincrement,
			string_column text,
			int_column integer
		)
	`)
	type Row struct {
		ID     int    `sql:"primary key autoincrement"`
		String string `sql:"string_column"`
		Number int    `sql:"int_column"`
	}

	schema := NewSchema(ForDB(db))
	sess := NewSession(ctx, db, schema)
	defer sess.Close()

	// insert three rows, IDs are automatically generated (1, 2, 3)
	for i, s := range []string{"AAAA", "BBBB", "CCCC"} {
		row := Row{
			String: s,
			Number: i,
		}
		_, err := sess.Row(&row).Exec(`insert into test_table({}) values({})`)
		if err != nil {
			t.Fatal("insert: ", err)
		}
	}

	{
		var rows []Row
		n, err := sess.Select(&rows, "select {} from test_table order by {}")
		if err != nil {
			t.Fatal("sqlrow.Select:", err)
		}
		if want := 3; n != want {
			t.Errorf("expected %d, got %d", want, n)
		}
	}

	{
		var rows []Row
		n, err := sess.Select(&rows, "select id, int_column, string_column from test_table order by {}")
		if err != nil {
			t.Fatal("sqlrow.Select:", err)
		}
		if want := 3; n != want {
			t.Errorf("expected %d, got %d", want, n)
		}
	}

	{
		var row Row
		n, err := sess.Select(&row, "select {} from test_table order by {}")
		if err != nil {
			t.Fatal("sqlrow.Select:", err)
		}
		if want := 3; n != want {
			t.Errorf("expected %d, got %d", want, n)
		}
		if want := "AAAA"; row.String != want {
			t.Errorf("want %q, got %q", want, row.String)
		}
		result, err := sess.Row(&row).Exec("update test_table set {} where {} and int_column = ?", 0)
		if err != nil {
			t.Fatal("sqlrow.Update:", err)
		}
		count, err := result.RowsAffected()
		wantNoError(t, err)
		if got, want := count, int64(1); got != want {
			t.Errorf("got=%d, want=%d", got, want)
		}
	}

	{
		var rows []Row
		n, err := sess.Select(&rows, "select {} from test_table where string_column in (?)", []string{
			"AAAA",
			"BBBB",
			"CCCC",
		})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := n, 3; got != want {
			t.Errorf("got = %d, want = %d", got, want)
		}
	}

	{
		expected := 5
		actual := len(schema.cache.stmts)
		if actual != expected {
			t.Errorf("statement cache: expected = %d, actual = %d", expected, actual)
		}
		for k, stmt := range schema.cache.stmts {
			t.Logf("%v=%v", k, stmt)
		}
	}
}

func TestJsonMarshaling(t *testing.T) {
	ctx := context.Background()
	db := sqliteDB(t)
	defer db.Close()

	mustExec(t, db, `
		create table test_table(
			id integer primary key autoincrement,
			name text,
			keyvals text
		)
	`)
	type KV struct {
		Key   string
		Value interface{}
	}
	type Row struct {
		ID      int `sql:"primary key autoincrement" table:"test_table"`
		Name    string
		Keyvals []KV `sql:"json"`
	}

	row := Row{
		Name: "first row",
		Keyvals: []KV{
			{"k1", "v1"},
			{"k2", 2},
		},
	}

	schema := NewSchema(ForDB(db))
	sess := NewSession(ctx, db, schema)

	err := sess.InsertRow(&row)
	wantNoError(t, err)

	{
		var row2 Row
		n, err := sess.Select(&row2, "select {} from test_table where {}", 1)
		wantNoError(t, err)
		if got, want := n, 1; got != want {
			t.Fatalf("got=%d, want=%d", got, want)
		}

		if got, want := fmt.Sprintf("%+v", row2), fmt.Sprintf("%+v", row); got != want {
			t.Fatalf("got=%s, want=%s", got, want)
		}
	}

	{
		var rows []Row
		n, err := sess.Select(&rows, "select {} from test_table where {}", 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if n != 1 {
			t.Fatalf("expected one row, got %d", n)
		}

		row2 := rows[0]
		expected := fmt.Sprintf("%+v", row)
		actual := fmt.Sprintf("%+v", row2)
		if expected != actual {
			t.Fatalf("expected %v, got %v", expected, actual)
		}
	}
}

func TestJsonNullMarshaling(t *testing.T) {
	ctx := context.Background()
	db := sqliteDB(t)
	defer db.Close()

	mustExec(t, db, `
		create table test_table(
			id integer primary key,
			keyvals text
		)
	`)
	type KV struct {
		Key   string
		Value interface{}
	}
	type Row struct {
		ID      int  `sql:"primary key autoincrement"`
		Keyvals []KV `sql:"json null"`
	}

	row := Row{
		ID:      1,
		Keyvals: nil,
	}

	schema := NewSchema(ForDB(db))
	sess := NewSession(ctx, db, schema)

	if _, err := sess.Row(&row).Exec("insert into test_table({}) values({})"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// check that the column is null
	{
		var keyvals sql.NullString
		err := db.QueryRow("select keyvals from test_table where id = 1").Scan(&keyvals)
		wantNoError(t, err)
		if keyvals.Valid {
			t.Fatalf("want=null, got=%q", keyvals.String)
		}
	}

	{
		var row2 Row
		n, err := sess.Select(&row2, "select {} from test_table where {}", 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if n != 1 {
			t.Fatalf("expected one row, got %d", n)
		}

		expected := fmt.Sprintf("%+v", row)
		actual := fmt.Sprintf("%+v", row2)
		if expected != actual {
			t.Fatalf("expected %v, got %v", expected, actual)
		}
	}
}

func TestRace(t *testing.T) {
	ctx := context.Background()
	db := postgresDB(t)
	defer db.Close()

	mustExec(t, db, `
		drop table if exists t1;
		create table t1 (
			id integer primary key,
			name text
		);
	`)
	defer mustExec(t, db, `drop table if exists t1`)

	type Row1 struct {
		ID   int `sql:"primary key"`
		Name string
	}

	schema := NewSchema(ForDB(db))

	var wg sync.WaitGroup

	const loops = 10

	for i := 0; i < loops; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sess := NewSession(ctx, db, schema)
			for j := 0; j < loops; j++ {
				id := i*loops + j
				row := Row1{
					ID:   id,
					Name: fmt.Sprintf("Row #%d", id),
				}
				if _, err := sess.Row(row).Exec("insert into t1({}) values({})"); err != nil {
					t.Errorf("cannot insert row %d: %v", id, err)
					return
				}

				var rows []Row1
				if _, err := sess.Select(&rows, "select {} from t1 order by id desc limit ?", id); err != nil {
					t.Errorf("%d: cannot query rows: %v", id, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	{
		expected := 2
		actual := len(schema.cache.stmts)
		if actual != expected {
			t.Errorf("statement cache: expected = %d, actual = %d", expected, actual)
		}
		for k, stmt := range schema.cache.stmts {
			t.Logf("%v=%v", k, stmt)
		}
	}
}

func TestNullable(t *testing.T) {
	ctx := context.Background()
	db := postgresDB(t)
	defer db.Close()

	if _, err := db.Exec(`drop table if exists nullable_types;`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(`drop table if exists nullable_types`)
	}()
	mustExec(t, db, `
		create table nullable_types(
			id integer not null primary key,
			i integer null,
			i8 integer null,
			i16 integer null,
			i32 integer null,
			i64 integer null,
			u integer null,
			uptr integer null,
			u8 integer null,
			u16 integer null,
			u32 integer null,
			u64 integer null,
			f32 double precision null,
			f64 double precision null,
			b boolean null,
			s text null,
			t timestamp with time zone null
		);
	`)

	schema := NewSchema(ForDB(db))
	sess := NewSession(ctx, db, schema)

	type Row struct {
		Id  int       `sql:"primary key"`
		I   int       `sql:"omitempty"`
		I8  int8      `sql:"emptynull"`
		I16 int16     `sql:"null"`
		I32 int32     `sql:"null"`
		I64 int64     `sql:"null"`
		U   uint      `sql:"null"`
		U8  uint8     `sql:"null"`
		U16 uint16    `sql:"null"`
		U32 uint32    `sql:"null"`
		U64 uint64    `sql:"null"`
		F32 float32   `sql:"null"`
		F64 float64   `sql:"null"`
		B   bool      `sql:"null"`
		S   string    `sql:"null"`
		T   time.Time `sql:"null"`
	}

	if _, err := db.Exec(`insert into nullable_types(id) values(1)`); err != nil {
		t.Fatal(err)
	}

	var row Row
	n, err := sess.Select(&row, "select {} from nullable_types where id = $1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := n, 1; want != got {
		t.Fatalf("want=%v, got=%v", want, got)
	}
	if got, want := row.I, 0; got != want {
		t.Fatalf("want=%v, got=%v", want, got)
	}
	if got, want := row.I8, int8(0); got != want {
		t.Fatalf("want=%v, got=%v", want, got)
	}
	if got, want := row.I16, int16(0); got != want {
		t.Fatalf("want=%v, got=%v", want, got)
	}
	if got, want := row.I32, int32(0); got != want {
		t.Fatalf("want=%v, got=%v", want, got)
	}
	if got, want := row.I64, int64(0); got != want {
		t.Fatalf("want=%v, got=%v", want, got)
	}

	if got, want := row.U, uint(0); got != want {
		t.Fatalf("want=%v, got=%v", want, got)
	}
	if got, want := row.U8, uint8(0); got != want {
		t.Fatalf("want=%v, got=%v", want, got)
	}
	if got, want := row.U16, uint16(0); got != want {
		t.Fatalf("want=%v, got=%v", want, got)
	}
	if got, want := row.U32, uint32(0); got != want {
		t.Fatalf("want=%v, got=%v", want, got)
	}
	if got, want := row.U64, uint64(0); got != want {
		t.Fatalf("want=%v, got=%v", want, got)
	}

	if got, want := row.F32, float32(0); got != want {
		t.Fatalf("want=%v, got=%v", want, got)
	}
	if got, want := row.F64, float64(0); got != want {
		t.Fatalf("want=%v, got=%v", want, got)
	}

	if got, want := row.B, false; got != want {
		t.Fatalf("want=%v, got=%v", want, got)
	}
	if got, want := row.S, ""; got != want {
		t.Fatalf("want=%v, got=%v", want, got)
	}
	if got, want := row.T, (time.Time{}); !got.Equal(want) {
		t.Fatalf("want=%v, got=%v", want, got)
	}

	row2 := row
	row2.Id = 2
	if _, err := sess.Row(row2).Exec("insert into nullable_types({}) values({})"); err != nil {
		t.Fatal(err)
	}

	const sql = `
		select {} from nullable_types
		where i is null
		and i8 is null
		and i16 is null
		and i32 is null
		and i64 is null
		and u is null
		and u8 is null
		and u16 is null
		and u32 is null
		and u64 is null
		and b is null
		and s is null
		and t is null
	`
	var rows []Row
	n, err = sess.Select(&rows, sql)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := n, 2; got != want {
		t.Errorf("want=%v, got=%v", want, got)
	}
}

func TestScalarQuery(t *testing.T) {
	db := postgresDB(t)
	defer db.Close()

	mustExec(t, db, `drop table if exists numbers`)
	defer mustExec(t, db, `drop table if exists numbers`)
	mustExec(t, db, `create table numbers(number int)`)
	const rowCount = 6
	for i := 0; i < rowCount; i++ {
		mustExec(t, db, fmt.Sprintf("insert into numbers(number) values(%d)", i))
	}

	schema := NewSchema(ForDB(db))
	sess := NewSession(context.Background(), db, schema)

	{
		var q func(query string, args ...interface{}) (int, error)
		sess.MakeQuery(&q)

		count, err := q("select count(*) from numbers")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got, want := count, rowCount; got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
	}

	{
		var q func(query string, args ...interface{}) (int32, error)
		sess.MakeQuery(&q)

		count, err := q("select count(*) from numbers")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got, want := count, int32(rowCount); got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
	}

	{
		var q func(query string, args ...interface{}) (int64, error)
		sess.MakeQuery(&q)

		count, err := q("select count(*) from numbers")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got, want := count, int64(rowCount); got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
	}

	{
		var q func(query string, args ...interface{}) (uint, error)
		sess.MakeQuery(&q)

		count, err := q("select count(*) from numbers")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got, want := count, uint(rowCount); got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
	}

	{
		var q func(query string, args ...interface{}) (uint32, error)
		sess.MakeQuery(&q)

		count, err := q("select count(*) from numbers")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got, want := count, uint32(rowCount); got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
	}

	{
		var q func(query string, args ...interface{}) (uint64, error)
		sess.MakeQuery(&q)

		count, err := q("select count(*) from numbers")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got, want := count, uint64(rowCount); got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
	}
}

func TestQuery(t *testing.T) {
	db := postgresDB(t)
	defer db.Close()

	mustExec(t, db, `drop table if exists widget;`)
	defer func() {
		mustExec(t, db, `drop table if exists widget`)
	}()
	mustExec(t, db, `
		create table widget(
			id integer not null primary key,
			name text not null
		);
	`)

	schema := NewSchema(ForDB(db))
	sess := NewSession(context.Background(), db, schema)

	type Widget struct {
		ID   int    `sql:"primary key"`
		Name string `sql:"natural key"`
	}
	type WidgetThunk func() (*Widget, error)

	const rowCount = 6
	var widget Widget

	for i := 0; i < rowCount; i++ {
		widget.ID = i
		widget.Name = fmt.Sprintf("Widget %d", i)
		if _, err := sess.Row(widget).Exec(`insert widget`); err != nil {
			t.Fatal(err)
		}
	}

	var dao struct {
		get        func(id int) (*Widget, error)
		load       func(id int) WidgetThunk
		getMany    func(ids ...int) ([]*Widget, error)
		selectRow  func(query string, args ...interface{}) (*Widget, error)
		selectRows func(query string, args ...interface{}) ([]*Widget, error)
	}

	sess.MakeQuery(&dao.get, &dao.getMany, &dao.selectRow, &dao.selectRows, &dao.load)

	for i := 0; i < rowCount; i++ {
		w, err := dao.get(i)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := w.Name, fmt.Sprintf("Widget %d", i); got != want {
			t.Errorf("got=%v, want=%v", got, want)
		}
	}

	{
		ids := make([]int, rowCount)
		for i := 0; i < rowCount; i++ {
			ids[i] = i
		}
		widgets, err := dao.getMany(ids...)
		if err != nil {
			t.Fatal(err)
		}
		sort.Slice(widgets, func(i, j int) bool {
			return widgets[i].ID < widgets[j].ID
		})
		for i, w := range widgets {
			if got, want := w.Name, fmt.Sprintf("Widget %d", i); got != want {
				t.Errorf("got=%v, want=%v", got, want)
			}
		}
	}

	// empty getMany should return empty slice
	{
		var ids []int
		widgets, err := dao.getMany(ids...)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(widgets), 0; got != want {
			t.Errorf("got=%v, want=%v", got, want)
		}
	}

	{
		widgets, err := dao.selectRows("select {} from widget order by id")
		if err != nil {
			t.Fatal(err)
		}
		for i, w := range widgets {
			if got, want := w.Name, fmt.Sprintf("Widget %d", i); got != want {
				t.Errorf("got=%v, want=%v", got, want)
			}
		}
	}

	for i := 0; i < rowCount; i++ {
		pattern := fmt.Sprintf("%%%d", i)
		w, err := dao.selectRow(`select {} from widget where name like ?`, pattern)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := w.Name, fmt.Sprintf("Widget %d", i); got != want {
			t.Errorf("got=%v, want=%v", got, want)
		}
	}

	{
		thunks := make([]WidgetThunk, rowCount)
		for i := 0; i < rowCount; i++ {
			thunks[i] = dao.load(i)
		}

		for i := 0; i < rowCount; i++ {
			w, err := thunks[i]()
			if err != nil {
				t.Errorf("want=no error, got=%v", err)
				continue
			}
			if got, want := w.Name, fmt.Sprintf("Widget %d", i); got != want {
				t.Errorf("got=%v, want=%v", got, want)
			}
		}
	}
}

func TestHandleRows(t *testing.T) {
	db := postgresDB(t)
	defer db.Close()

	mustExec(t, db, `drop table if exists widget;`)
	defer func() {
		mustExec(t, db, `drop table if exists widget`)
	}()
	mustExec(t, db, `
		create table widget(
			id integer not null primary key,
			name text not null
		);
	`)

	schema := NewSchema(ForDB(db))
	sess := NewSession(context.Background(), db, schema)

	type Widget struct {
		ID   int    `sql:"primary key"`
		Name string `sql:"natural key"`
	}
	type WidgetThunk func() (*Widget, error)

	const rowCount = 6
	var widget Widget

	for i := 0; i < rowCount; i++ {
		widget.ID = i
		widget.Name = fmt.Sprintf("Widget %d", i)
		if _, err := sess.Row(widget).Exec(`insert widget`); err != nil {
			t.Fatal(err)
		}
	}

	var dao struct {
		get        func(id int) (*Widget, error)
		load       func(id int) WidgetThunk
		getMany    func(ids ...int) ([]*Widget, error)
		selectRow  func(query string, args ...interface{}) (*Widget, error)
		selectRows func(query string, args ...interface{}) ([]*Widget, error)
	}

	sess.MakeQuery(&dao.get, &dao.getMany, &dao.selectRow, &dao.selectRows, &dao.load)

	var handledRows []*Widget
	sess.HandleRows(func(rows []*Widget) {
		handledRows = append(handledRows, rows...)
	})

	for i := 0; i < rowCount; i++ {
		handledRows = nil
		w, err := dao.get(i)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(handledRows), 1; got != want {
			t.Errorf("got=%v, want=%v", got, want)
		}
		if got, want := handledRows[0], w; got != want {
			t.Errorf("got=%v, want=%v", got, want)
		}
	}

	{
		ids := make([]int, rowCount)
		for i := 0; i < rowCount; i++ {
			ids[i] = i
		}
		handledRows = nil
		widgets, err := dao.getMany(ids...)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(handledRows), len(widgets); got != want {
			t.Errorf("got=%v, want=%v", got, want)
		}
		for i, w := range widgets {
			if got, want := handledRows[i], w; got != want {
				t.Errorf("got=%v, want=%v", got, want)
			}
		}
	}
	{
		handledRows = nil
		widgets, err := dao.selectRows("select {} from widget order by id")
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(handledRows), len(widgets); got != want {
			t.Errorf("got=%v, want=%v", got, want)
		}
		for i, w := range widgets {
			if got, want := handledRows[i], w; got != want {
				t.Errorf("got=%v, want=%v", got, want)
			}
		}
	}

	for i := 0; i < rowCount; i++ {
		handledRows = nil
		pattern := fmt.Sprintf("%%%d", i)
		w, err := dao.selectRow(`select {} from widget where name like ?`, pattern)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(handledRows), 1; got != want {
			t.Errorf("got=%v, want=%v", got, want)
		}
		if got, want := handledRows[0], w; got != want {
			t.Errorf("got=%v, want=%v", got, want)
		}
	}

	{
		handledRows = nil
		thunks := make([]WidgetThunk, rowCount)
		for i := 0; i < rowCount; i++ {
			thunks[i] = dao.load(i)
		}

		for i := 0; i < rowCount; i++ {
			w, err := thunks[i]()
			if err != nil {
				t.Errorf("want=no error, got=%v", err)
				continue
			}
			if got, want := handledRows[i], w; got != want {
				t.Errorf("got=%v, want=%v", got, want)
			}
		}
	}
}

func TestInsertRow_NoAutoIncr(t *testing.T) {
	db := postgresDB(t)
	defer db.Close()

	mustExec(t, db, `drop table if exists no_auto_incr;`)
	defer mustExec(t, db, `drop table if exists no_auto_incr;`)
	mustExec(t, db, `
		create table no_auto_incr(
			id int primary key not null, 
			name text, 
			created_at timestamp with time zone, 
			updated_at timestamp with time zone
		)`,
	)

	type NoAutoIncr struct {
		ID        int `sql:"primary key"`
		Name      string
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	schema := NewSchema(WithDialect(Postgres))
	sess := NewSession(context.Background(), db, schema)

	rows := []NoAutoIncr{
		NoAutoIncr{
			ID:   1,
			Name: "row 1",
		},
	}

	for _, row := range rows {

		started := time.Now()
		if err := sess.InsertRow(&row); err != nil {
			t.Fatalf("want no error, got %v", err)
		}

		if row.CreatedAt.Before(started) {
			t.Errorf("wanted created_at < started, got %v", row.CreatedAt)
		}
		if !row.UpdatedAt.Equal(row.CreatedAt) {
			t.Errorf("wanted updated_at = created_at, got %v", row.UpdatedAt)
		}

		var getRow func(id int) (*NoAutoIncr, error)
		sess.MakeQuery(&getRow)

		row2, err := getRow(1)
		if err != nil {
			t.Fatalf("want no error, got %v", err)
		}

		if row2 == nil {
			t.Fatalf("want non-nil, got nil")
		}
		if got, want := row2.Name, row.Name; got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
		if got, want := row2.CreatedAt.Format(time.RFC3339), row.CreatedAt.Format(time.RFC3339); got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
		if got, want := row2.UpdatedAt.Format(time.RFC3339), row.UpdatedAt.Format(time.RFC3339); got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
	}
}

func TestInsertRow_Serial(t *testing.T) {
	db := postgresDB(t)
	defer db.Close()

	mustExec(t, db, `drop table if exists auto_incr;`)
	defer mustExec(t, db, `drop table if exists auto_incr;`)
	mustExec(t, db, `
		create table auto_incr(
			id serial primary key not null, 
			name text, 
			created_at timestamp with time zone, 
			updated_at timestamp with time zone
		)`,
	)

	type AutoIncr struct {
		ID        int `sql:"primary key autoincr"`
		Name      string
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	schema := NewSchema(WithDialect(Postgres))
	sess := NewSession(context.Background(), db, schema)

	rows := []AutoIncr{
		AutoIncr{
			Name: "row 1",
		},
		AutoIncr{
			Name: "row 2",
		},
	}

	for i, row := range rows {

		started := time.Now()
		if err := sess.InsertRow(&row); err != nil {
			t.Fatalf("want no error, got %v", err)
		}

		if row.CreatedAt.Before(started) {
			t.Errorf("wanted created_at < started, got %v", row.CreatedAt)
		}
		if !row.UpdatedAt.Equal(row.CreatedAt) {
			t.Errorf("wanted updated_at = created_at, got %v", row.UpdatedAt)
		}

		var getRow func(id int) (*AutoIncr, error)
		sess.MakeQuery(&getRow)

		row2, err := getRow(i + 1)
		if err != nil {
			t.Fatalf("want no error, got %v", err)
		}

		if row2 == nil {
			t.Fatalf("want non-nil, got nil")
		}
		if got, want := row2.Name, row.Name; got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
		if got, want := row2.CreatedAt.Format(time.RFC3339), row.CreatedAt.Format(time.RFC3339); got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
		if got, want := row2.UpdatedAt.Format(time.RFC3339), row.UpdatedAt.Format(time.RFC3339); got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
	}
}

func TestInsertRow_AutoIncr(t *testing.T) {
	db := sqliteDB(t)
	defer db.Close()

	mustExec(t, db, `drop table if exists auto_incr;`)
	defer mustExec(t, db, `drop table if exists auto_incr;`)
	mustExec(t, db, `
		create table auto_incr(
			id integer primary key autoincrement, 
			name text, 
			created_at datetime, 
			updated_at datetime
		)`,
	)

	type AutoIncr struct {
		ID        int `sql:"primary key autoincr"`
		Name      string
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	schema := NewSchema(WithDialect(SQLite))
	sess := NewSession(context.Background(), db, schema)

	rows := []AutoIncr{
		AutoIncr{
			Name: "row 1",
		},
		AutoIncr{
			Name: "row 2",
		},
	}

	for i, row := range rows {

		started := time.Now()
		if err := sess.InsertRow(&row); err != nil {
			t.Fatalf("want no error, got %v", err)
		}

		if row.CreatedAt.Before(started) {
			t.Errorf("wanted created_at < started, got %v", row.CreatedAt)
		}
		if !row.UpdatedAt.Equal(row.CreatedAt) {
			t.Errorf("wanted updated_at = created_at, got %v", row.UpdatedAt)
		}

		var getRow func(id int) (*AutoIncr, error)
		sess.MakeQuery(&getRow)

		row2, err := getRow(i + 1)
		if err != nil {
			t.Fatalf("want no error, got %v", err)
		}

		if row2 == nil {
			t.Fatalf("want non-nil, got nil")
		}
		if got, want := row2.Name, row.Name; got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
		if got, want := row2.CreatedAt.Format(time.RFC3339), row.CreatedAt.Format(time.RFC3339); got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
		if got, want := row2.UpdatedAt.Format(time.RFC3339), row.UpdatedAt.Format(time.RFC3339); got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
	}
}

func TestUpdateRow_NoVersion_NoUpdatedAt(t *testing.T) {
	db := postgresDB(t)
	defer db.Close()

	mustExec(t, db, `drop table if exists no_version_no_updated_at;`)
	defer mustExec(t, db, `drop table if exists no_version_no_updated_at;`)
	mustExec(t, db, `
		create table no_version_no_updated_at(
			id int primary key not null, 
			name text,
			counter int
		)`,
	)

	type NoVersionNoUpdatedAt struct {
		ID      int `sql:"primary key"`
		Name    string
		Counter int
	}

	schema := NewSchema(WithDialect(Postgres))
	sess := NewSession(context.Background(), db, schema)

	rows := []NoVersionNoUpdatedAt{
		NoVersionNoUpdatedAt{
			ID:   1,
			Name: "row 1",
		},
	}

	for _, row := range rows {
		if err := sess.InsertRow(&row); err != nil {
			t.Fatalf("want no error, got %v", err)
		}

		var getRow func(id int) (*NoVersionNoUpdatedAt, error)
		sess.MakeQuery(&getRow)

		row2, err := getRow(row.ID)
		if err != nil {
			t.Fatalf("want no error, got %v", err)
		}

		if row2 == nil {
			t.Fatalf("want non-nil, got nil")
		}
		if got, want := row2.Name, row.Name; got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}

		row2.Counter++
		rowCount, err := sess.UpdateRow(row2)
		if err != nil {
			t.Fatalf("got=%v, want=nil", err)
		}
		if got, want := rowCount, 1; got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
	}
}

func TestUpdateRow_NoVersion_UpdatedAt(t *testing.T) {
	db := postgresDB(t)
	defer db.Close()

	mustExec(t, db, `drop table if exists no_version_updated_at;`)
	defer mustExec(t, db, `drop table if exists no_version_updated_at;`)
	mustExec(t, db, `
		create table no_version_updated_at(
			id int primary key not null, 
			name text,
			counter int,
			created_at timestamp with time zone, 
			updated_at timestamp with time zone
		)`,
	)

	type NoVersionUpdatedAt struct {
		ID        int `sql:"primary key"`
		Name      string
		Counter   int
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	schema := NewSchema(WithDialect(Postgres))
	sess := NewSession(context.Background(), db, schema)

	rows := []NoVersionUpdatedAt{
		NoVersionUpdatedAt{
			ID:   1,
			Name: "row 1",
		},
	}

	for _, row := range rows {

		started := time.Now()
		if err := sess.InsertRow(&row); err != nil {
			t.Fatalf("want no error, got %v", err)
		}

		if row.CreatedAt.Before(started) {
			t.Errorf("wanted created_at < started, got %v", row.CreatedAt)
		}
		if !row.UpdatedAt.Equal(row.CreatedAt) {
			t.Errorf("wanted updated_at = created_at, got %v", row.UpdatedAt)
		}

		var getRow func(id int) (*NoVersionUpdatedAt, error)
		sess.MakeQuery(&getRow)

		row2, err := getRow(row.ID)
		if err != nil {
			t.Fatalf("want no error, got %v", err)
		}

		if row2 == nil {
			t.Fatalf("want non-nil, got nil")
		}
		if got, want := row2.Name, row.Name; got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
		if got, want := row2.CreatedAt.Format(time.RFC3339), row.CreatedAt.Format(time.RFC3339); got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
		if got, want := row2.UpdatedAt.Format(time.RFC3339), row.UpdatedAt.Format(time.RFC3339); got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}

		row2.Counter++
		updated := time.Now()
		rowCount, err := sess.UpdateRow(row2)
		if err != nil {
			t.Fatalf("got=%v, want=nil", err)
		}
		if got, want := rowCount, 1; got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
		if row2.UpdatedAt.Before(updated) {
			t.Errorf("wanted updated_at > %v, got %v", updated, row.UpdatedAt)
		}
	}
}

func TestUpdateRow_Version_UpdatedAt(t *testing.T) {
	db := postgresDB(t)
	defer db.Close()

	mustExec(t, db, `drop table if exists version_updated_at;`)
	defer mustExec(t, db, `drop table if exists version_updated_at;`)
	mustExec(t, db, `
		create table version_updated_at(
			id int primary key not null, 
			version int not null,
			name text,
			counter int,
			created_at timestamp with time zone, 
			updated_at timestamp with time zone
		)`,
	)

	type VersionUpdatedAt struct {
		ID        int   `sql:"primary key"`
		Version   int64 `sql:"version"`
		Name      string
		Counter   int
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	schema := NewSchema(WithDialect(Postgres))
	sess := NewSession(context.Background(), db, schema)

	rows := []VersionUpdatedAt{
		VersionUpdatedAt{
			ID:   1,
			Name: "row 1",
		},
	}

	for _, row := range rows {

		started := time.Now()
		if err := sess.InsertRow(&row); err != nil {
			t.Fatalf("want no error, got %v", err)
		}

		if row.CreatedAt.Before(started) {
			t.Errorf("wanted created_at < started, got %v", row.CreatedAt)
		}
		if !row.UpdatedAt.Equal(row.CreatedAt) {
			t.Errorf("wanted updated_at = created_at, got %v", row.UpdatedAt)
		}

		var getRow func(id int) (*VersionUpdatedAt, error)
		sess.MakeQuery(&getRow)

		row2, err := getRow(row.ID)
		if err != nil {
			t.Fatalf("want no error, got %v", err)
		}

		if row2 == nil {
			t.Fatalf("want non-nil, got nil")
		}
		if got, want := row2.Name, row.Name; got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
		if got, want := row2.CreatedAt.Format(time.RFC3339), row.CreatedAt.Format(time.RFC3339); got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
		if got, want := row2.UpdatedAt.Format(time.RFC3339), row.UpdatedAt.Format(time.RFC3339); got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
		if got, want := row2.Version, int64(1); got != want {
			t.Errorf("got=%v, want=%v", got, want)
		}

		row2.Counter++
		updated := time.Now()
		rowCount, err := sess.UpdateRow(row2)
		if err != nil {
			t.Fatalf("got=%v, want=nil", err)
		}
		if got, want := rowCount, 1; got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
		if row2.UpdatedAt.Before(updated) {
			t.Errorf("wanted updated_at > %v, got %v", updated, row.UpdatedAt)
		}
		if got, want := row2.Version, int64(2); got != want {
			t.Errorf("got=%v, want=%v", got, want)
		}

		// set the version to the wrong value
		row2.Version = 1
		rowCount, err = sess.UpdateRow(row2)
		if got, want := rowCount, 0; got != want {
			t.Fatalf("got=%v, want=%v", got, want)
		}
		if err == nil {
			t.Fatal("got=nil, want=non-nil error")
		}
		if got, want := err.Error(), `optimistic locking conflict rowType="sqlr.VersionUpdatedAt" ID=1 expectedVersion=1 actualVersion=2`; got != want {
			t.Errorf("got=%v, want=%v", got, want)
		}
	}
}

func TestTableName(t *testing.T) {
	type NamedRow struct {
		ID   int
		Name string
	}

	type StructTagRow struct {
		ID   int `table:"table_name"`
		Name string
	}

	tests := []struct {
		tests  string
		schema *Schema
		row    interface{}
		want   string
	}{
		{
			tests:  "anon class with struct tag",
			schema: NewSchema(WithNamingConvention(SnakeCase)),
			row: struct {
				ID   int `sql:"primary key" table:"table_name_in_struct_tag"`
				Name string
			}{},
			want: "table_name_in_struct_tag",
		},
		{
			tests:  "named type, no config or struct tag",
			schema: NewSchema(WithNamingConvention(SnakeCase)),
			row:    &NamedRow{},
			want:   "named_row",
		},
		{
			tests:  "named type with struct tag",
			schema: NewSchema(WithNamingConvention(SnakeCase)),
			row:    &StructTagRow{},
			want:   "table_name",
		},
		{
			tests: "config overrides named type",
			schema: NewSchema(
				WithTables(TablesConfig{
					(*NamedRow)(nil): TableConfig{TableName: "override_name"},
				}),
			),
			row:  &NamedRow{},
			want: "override_name",
		},
		{
			tests: "config overrides struct tag",
			schema: NewSchema(
				WithTables(TablesConfig{
					(*StructTagRow)(nil): TableConfig{TableName: "override_name"},
				}),
			),
			row:  &StructTagRow{},
			want: "override_name",
		},
	}

	for tn, tt := range tests {
		tbl := tt.schema.TableFor(tt.row)
		if got, want := tbl.Name(), tt.want; got != want {
			t.Fatalf("%d: got=%q want=%q", tn, got, want)
		}
	}
}

func TestExcludeColumns(t *testing.T) {
	db := postgresDB(t)
	defer db.Close()

	mustExec(t, db, `drop table if exists exclude_columns;`)
	defer mustExec(t, db, `drop table if exists exclude_columns;`)
	mustExec(t, db, `
		create table exclude_columns(
			id int primary key not null, 
			n1 int not null,
			n2 int not null,
			n3 int not null
		)`,
	)

	type Row struct {
		ID int `sql:"primary key" table:"exclude_columns"`
		N1 int
		N2 int
		N3 int
	}

	schema := NewSchema(WithDialect(Postgres))
	sess := NewSession(context.Background(), db, schema)

	err := sess.InsertRow(&Row{
		ID: 1,
		N1: 1,
		N2: 2,
		N3: 3,
	})
	wantNoError(t, err)

	var selectRow func(query string, args ...interface{}) (*Row, error)
	sess.MakeQuery(&selectRow)

	check := func(got, want int) {
		t.Helper()
		if got != want {
			t.Fatalf("got=%d want=%d", got, want)
		}
	}

	row, err := selectRow("select {} from exclude_columns")
	wantNoError(t, err)
	check(row.N1, 1)
	check(row.N2, 2)
	check(row.N3, 3)

	row, err = selectRow("select 0 as n1, {exclude n1} from exclude_columns")
	wantNoError(t, err)
	check(row.N1, 0)
	check(row.N2, 2)
	check(row.N3, 3)

	row, err = selectRow("select 10 as n1, 20 as n2, {exclude n1, n2} from exclude_columns")
	wantNoError(t, err)
	check(row.N1, 10)
	check(row.N2, 20)
	check(row.N3, 3)
}

// mustExec performs an SQL command, which must succeed or the test stops
func mustExec(t *testing.T, db *sql.DB, query string) {
	t.Helper()
	if _, err := db.Exec(query); err != nil {
		t.Fatal(err)
	}
}

// postgresDB returns a *sql.DB for accessing the test PostgreSQL database.
func postgresDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", "postgres://sqlr_test:sqlr_test@localhost/sqlr_test?sslmode=disable")
	if err != nil {
		t.Fatal("sql.Open:", err)
	}
	return db
}

// sqliteDB returns a *sql.DB for accessing a test SQLite database.
func sqliteDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("sql.Open:", err)
	}
	return db
}

func wantNoError(t *testing.T, err error, args ...interface{}) {
	t.Helper()
	if err != nil {
		args = append(args, err)
		t.Fatal(args...)
	}
}
