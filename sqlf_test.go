package sqlf_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/cznic/ql"
	"github.com/jjeffery/sqlf"
	"github.com/jjeffery/sqlf/private/dialect"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func checkError(t *testing.T, err error, msg string) {
	if err != nil {
		t.Fatal(msg+": ", err)
	}
}

func setup(t *testing.T) {
	var err error
	ql.RegisterMemDriver()
	db, err = sql.Open("sqlite3", ":memory:")
	sqlf.DefaultSchema.Dialect = dialect.New("sqlite3")
	checkError(t, err, "cannot open db")

	tx, err := db.Begin()
	checkError(t, err, "cannot begin tx")

	_, err = tx.Exec(`
	create table users(
		id int,
		name string,
		updated_at datetime
	)`)
	checkError(t, err, "cannot create table")
	checkError(t, tx.Commit(), "cannot commit")
}

type User struct {
	ID        int `sql:",autoincr"`
	Name      string
	UpdatedAt time.Time
}

func Test1(t *testing.T) {
	setup(t)
	table := sqlf.Table("users", User{})
	stmt := sqlf.InsertRowPrintf(`
	insert into %s(%s) 
	values(%s)`,
		table.Insert.TableName,
		table.Insert.Columns,
		table.Insert.Values)
	t.Logf("insert query: %s", stmt.Query())
	tx, err := db.Begin()
	checkError(t, err, "cannot begin tx")

	u := &User{Name: "Name", UpdatedAt: time.Now()}

	err = stmt.Exec(tx, u)
	checkError(t, err, "cannot insert")
	checkError(t, tx.Commit(), "cannot commit")

	if u.ID != 1 {
		t.Errorf("expected=1, actual=%d", u.ID)
	}
}
