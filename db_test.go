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
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("sql.Open:", err)
	}
	defer db.Close()

	_, err = db.Exec(`
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

	// insert three rows, IDs are automatically generated (1, 2, 3)
	for i, s := range []string{"AAAA", "BBBB", "CCCC"} {
		row := Row{
			String: s,
			Number: i,
		}
		_, err = schema.Exec(db, &row, `insert into test_table({}) values({})`)
		if err != nil {
			t.Fatal("insert: ", err)
		}
	}

	{
		var rows []Row
		n, err := schema.Select(db, &rows, "select {} from test_table order by {}")
		if err != nil {
			t.Fatal("sqlrow.Select:", err)
		}
		if want := 3; n != want {
			t.Errorf("expected %d, got %d", want, n)
		}
	}

	{
		var rows []Row
		n, err := schema.Select(db, &rows, "select id, int_column, string_column from test_table order by {}")
		if err != nil {
			t.Fatal("sqlrow.Select:", err)
		}
		if want := 3; n != want {
			t.Errorf("expected %d, got %d", want, n)
		}
	}

	{
		var row Row
		n, err := schema.Select(db, &row, "select {} from test_table order by {}")
		if err != nil {
			t.Fatal("sqlrow.Select:", err)
		}
		if want := 3; n != want {
			t.Errorf("expected %d, got %d", want, n)
		}
		if want := "AAAA"; row.String != want {
			t.Errorf("want %q, got %q", want, row.String)
		}
		n, err = schema.Exec(db, &row, "update test_table set {} where {} and int_column = ?", 0)
		if err != nil {
			t.Fatal("sqlrow.Update:", err)
		}
		if want := 1; n != want {
			t.Errorf("expected %d, got %d", want, n)
		}
	}

	{
		var rows []Row
		n, err := schema.Select(db, &rows, "select {} from test_table where string_column in (?)", []string{
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
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("sql.Open:", err)
	}
	defer db.Close()

	_, err = db.Exec(`
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
		ID      int `sql:"primary key autoincrement"`
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

	_, err = schema.Exec(db, &row, "insert into test_table({}) values({})")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	{
		var row2 Row
		n, err := schema.Select(db, &row2, "select {} from test_table where {}", 1)
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

	{
		var rows []Row
		n, err := schema.Select(db, &rows, "select {} from test_table where {}", 1)
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

func TestRace(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://sqlr_test:sqlr_test@localhost/sqlr_test?sslmode=disable")
	if err != nil {
		t.Fatal("sql.Open:", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		drop table if exists t1;
		create table t1 (
			id integer primary key,
			name text
		);
	`)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Exec(`drop table if exists t1`)

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
			for j := 0; j < loops; j++ {
				id := i*loops + j
				row := Row1{
					ID:   id,
					Name: fmt.Sprintf("Row #%d", id),
				}
				if _, err := schema.Exec(db, row, "insert into t1({}) values({})"); err != nil {
					t.Errorf("cannot insert row %d: %v", id, err)
					return
				}

				var rows []Row1
				if _, err := schema.Select(db, &rows, "select {} from t1 order by id desc limit ?", id); err != nil {
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
	db, err := sql.Open("postgres", "postgres://sqlr_test:sqlr_test@localhost/sqlr_test?sslmode=disable")
	if err != nil {
		t.Fatal("sql.Open:", err)
	}
	defer db.Close()

	if _, err := db.Exec(`drop table if exists nullable_types;`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(`drop table if exists nullable_types`)
	}()
	if _, err = db.Exec(`
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
	`); err != nil {
		t.Fatal(err)
	}

	schema := NewSchema(ForDB(db))

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
	n, err := schema.Select(db, &row, "select {} from nullable_types where id = $1", 1)
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
	if _, err := schema.Exec(db, row2, "insert into nullable_types({}) values({})"); err != nil {
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
	n, err = schema.Select(db, &rows, sql)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := n, 2; got != want {
		t.Errorf("want=%v, got=%v", want, got)
	}
}

func TestQuery(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://sqlr_test:sqlr_test@localhost/sqlr_test?sslmode=disable")
	if err != nil {
		t.Fatal("sql.Open:", err)
	}
	defer db.Close()

	if _, err := db.Exec(`drop table if exists widgets;`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		// _, _ = db.Exec(`drop table if exists widgets`)
	}()
	if _, err = db.Exec(`
		create table widgets(
			id integer not null primary key,
			name text not null
		);
	`); err != nil {
		t.Fatal(err)
	}

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
		if _, err := sess.Exec(widget, `insert widgets`); err != nil {
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

	builder := NewRowFunc(sess, &Widget{}, TableName("widgets"))
	builder.MustMakeQuery(&dao.get, &dao.getMany, &dao.selectRow, &dao.selectRows, &dao.load)

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

	{
		widgets, err := dao.selectRows("select {} from widgets order by id")
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
		w, err := dao.selectRow(`select {} from widgets where name like ?`, pattern)
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
				t.Logf("not implemented yet: want=no error, got=%v", err)
				continue
			}
			if got, want := w.Name, fmt.Sprintf("Widget %d", i); got != want {
				t.Errorf("got=%v, want=%v", got, want)
			}
		}
	}
}
