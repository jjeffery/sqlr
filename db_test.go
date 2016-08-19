package sqlrow

import (
	"database/sql"
	"fmt"
	"testing"

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
