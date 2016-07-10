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
		id integer primary key autoincrement,
		name string,
		updated_at datetime
	)`)
	checkError(t, err, "cannot create table")
	checkError(t, tx.Commit(), "cannot commit")
}

type User struct {
	ID        int `sql:",pk autoincr"`
	Name      string
	UpdatedAt time.Time
}

func TestInsertUpdate(t *testing.T) {
	setup(t)
	table := sqlf.Table("users", User{})
	insertStmt := sqlf.InsertRowPrintf(`insert into %s(%s) values(%s)`,
		table.Insert.TableName,
		table.Insert.Columns,
		table.Insert.Values)
	t.Logf("insert query: %s", insertStmt.Query())
	tx, err := db.Begin()
	defer tx.Rollback()
	checkError(t, err, "cannot begin tx")

	u1 := &User{Name: "Name", UpdatedAt: time.Now()}

	err = insertStmt.Exec(tx, u1)
	checkError(t, err, "cannot insert")

	if u1.ID != 1 {
		t.Errorf("expected=1, actual=%d", u1.ID)
	}

	updateStmt := sqlf.UpdateRowPrintf(`update %s set %s where %s`,
		table.Update.TableName,
		table.Update.SetColumns,
		table.Update.WhereColumns)
	t.Logf("update query: %s", updateStmt.Query())

	u1.Name = "Another name"
	u1.UpdatedAt = time.Now()
	var rowsAffected int
	rowsAffected, err = updateStmt.Exec(tx, u1)
	checkError(t, err, "cannot update")
	if rowsAffected != 1 {
		t.Errorf("rows updated: expected=1, actual=%d", rowsAffected)
	}

	u2 := &User{Name: "User2", UpdatedAt: time.Now()}
	err = insertStmt.Exec(tx, u2)
	checkError(t, err, "cannot insert")
	if u2.ID != 2 {
		t.Errorf("expected=2, actual=%d", u2.ID)
	}

	selectStmt := sqlf.SelectRowsPrintf("select %s from %s order by %s",
		table.Select.Columns,
		table.Select.TableName,
		table.Select.OrderBy)

	var users []*User
	err = selectStmt.Select(tx, &users)
	checkError(t, err, "cannot select")

	if len(users) != 2 {
		t.Errorf("select: expected rows=2, actual=%d", len(users))
	}
	t.Log("users: ", users)
	for i, u := range users {
		t.Logf("user %d: %v", i+1, u)
	}

	deleteStmt := sqlf.UpdateRowPrintf(`delete from %s where %s`,
		table.Delete.TableName,
		table.Delete.WhereColumns)
	t.Logf("deltete query: %s", deleteStmt.Query())

	rowsAffected, err = deleteStmt.Exec(tx, u1)
	checkError(t, err, "cannot delete")
	if rowsAffected != 1 {
		t.Errorf("rows deleted: expected=1, actual=%d", rowsAffected)
	}

	checkError(t, tx.Commit(), "cannot commit")
}
