package sqlrow

import (
	"database/sql"
	"fmt"
	"sync"
	"testing"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func TestDB1(t *testing.T) {
	// clear the stmt cache so we can check at the end of the test
	clearStmtCache()

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

	// insert three rows, IDs are automatically generated (1, 2, 3)
	for i, s := range []string{"AAAA", "BBBB", "CCCC"} {
		row := Row{
			String: s,
			Number: i,
		}
		err = Insert(db, &row, `test_table`)
		if err != nil {
			t.Fatal("insert: ", err)
		}
	}

	{
		var rows []Row
		n, err := Select(db, &rows, "select {} from test_table order by {}")
		if err != nil {
			t.Fatal("sqlrow.Select:", err)
		}
		if want := 3; n != want {
			t.Errorf("expected %d, got %d", want, n)
		}
	}

	{
		var row Row
		n, err := Select(db, &row, "select {} from test_table order by {}")
		if err != nil {
			t.Fatal("sqlrow.Select:", err)
		}
		if want := 3; n != want {
			t.Errorf("expected %d, got %d", want, n)
		}
		if want := "AAAA"; row.String != want {
			t.Errorf("want %q, got %q", want, row.String)
		}
		n, err = Update(db, &row, "update test_table set {} where {} and int_column = ?", 0)
		if err != nil {
			t.Fatal("sqlrow.Update:", err)
		}
		if want := 1; n != want {
			t.Errorf("expected %d, got %d", want, n)
		}
	}

	{
		expected := 3
		actual := len(stmtCache.stmts)
		if actual != expected {
			t.Errorf("statement cache: expected = %d, actual = %d", expected, actual)
		}
		for k, stmt := range stmtCache.stmts {
			t.Logf("%s=%v", k, stmt)
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

	err = Insert(db, &row, "test_table")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	{
		var row2 Row
		n, err := Select(db, &row2, "test_table", 1)
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
		n, err := Select(db, &rows, "test_table", 1)
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
	// clear the stmt cache so we can check at the end of the test
	clearStmtCache()

	db, err := sql.Open("postgres", "postgres://sqlrow_test:sqlrow_test@localhost/sqlrow_test?sslmode=disable")
	if err != nil {
		t.Fatal("sql.Open:", err)
	}
	defer db.Close()

	Default.Dialect = DialectFor("postgres")
	defer func() { Default.Dialect = nil }()

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
				if err := Insert(db, row, "t1"); err != nil {
					t.Errorf("cannot insert row %d: %v", id, err)
					return
				}

				var rows []Row1
				if _, err := Select(db, &rows, "select {} from t1 order by id desc limit ?", id); err != nil {
					t.Errorf("%d: cannot query rows: %v", id, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	{
		expected := 2
		actual := len(stmtCache.stmts)
		if actual != expected {
			t.Errorf("statement cache: expected = %d, actual = %d", expected, actual)
		}
		for k, stmt := range stmtCache.stmts {
			t.Logf("%s=%v", k, stmt)
		}
	}
}
