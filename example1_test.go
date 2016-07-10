package sqlf_test

import (
	"database/sql"
	"log"
	"time"

	"github.com/jjeffery/sqlf"
	_ "github.com/lib/pq"
)

// The UserRow struct represents a single row in the users table.
// Note that the sqlf package becomes more useful with row structs
// which have many more columns than this example.
type UserRow struct {
	ID        int64 `sql:",primary key autoincr"`
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Data access functions for accessing the users table.
// See the setupStmts function below to see how these
// functions are created using the sqlf package.
var (
	insertUserRow     func(db sqlf.Execer, u *UserRow) error
	updateUserRow     func(db sqlf.Execer, u *UserRow) (n int, err error)
	deleteUserRow     func(db sqlf.Execer, u *UserRow) (n int, err error)
	getUserRow        func(db sqlf.Queryer, id int) (*UserRow, error)
	selectAllUserRows func(db sqlf.Queryer) ([]*UserRow, error)
)

func Example() {
	db, err := sql.Open("postgres", "user=test dbname=test sslmode=disable")
	checkForError(err)

	// connected to database: setup statements using sqlf functions
	setupStmts()

	tx, err := db.Begin()
	checkForError(err)
	defer tx.Rollback()

	u1 := &UserRow{
		Name: "John Doe",
	}
	if err := insertUserRow(tx, u1); err != nil {
		log.Fatal(err)
	}

	u2 := &UserRow{
		Name: "Jane Doe",
	}
	if err := insertUserRow(tx, u2); err != nil {
		log.Fatal(err)
	}

	users, err := selectAllUserRows(tx)
	checkForError(err)
	for i, u := range users {
		log.Printf("user %d: %v", i+1, u)
	}
}

// setupStmts is an example of how to create type-safe data access functions
func setupStmts() {
	table := sqlf.Table("users", UserRow{})

	insertStmt := sqlf.InsertRowPrintf("insert into %s(%s) values(%s)",
		table.Insert.TableName,
		table.Insert.Columns,
		table.Insert.Values)
	insertUserRow = func(db sqlf.Execer, u *UserRow) error {
		u.CreatedAt = time.Now()
		u.UpdatedAt = u.CreatedAt
		return insertStmt.Exec(db, u)
	}

	updateStmt := sqlf.UpdateRowPrintf("update %s set %s where %s",
		table.Update.TableName,
		table.Update.SetColumns,
		table.Update.WhereColumns)
	updateUserRow = func(db sqlf.Execer, u *UserRow) (int, error) {
		u.UpdatedAt = time.Now()
		return updateStmt.Exec(db, u)
	}

	deleteStmt := sqlf.UpdateRowPrintf("delete from %s where %s",
		table.Delete.TableName,
		table.Delete.WhereColumns)
	deleteUserRow = func(db sqlf.Execer, u *UserRow) (int, error) {
		return deleteStmt.Exec(db, u)
	}

	getStmt := sqlf.SelectRowsPrintf(`
		select %s
		from %s
		where id = ?`,
		table.Select.Columns,
		table.Select.TableName)
	getUserRow = func(db sqlf.Queryer, id int) (*UserRow, error) {
		var users []*UserRow
		if err := getStmt.Select(db, &users, id); err != nil {
			return nil, err
		}
		if len(users) == 0 {
			// not found
			return nil, nil
		}
		return users[0], nil
	}

	selectAllStmt := sqlf.SelectRowsPrintf(`
		select %s
		from %s
		order by %s`, table.Select.Columns,
		table.Select.TableName,
		table.Select.OrderBy)
	selectAllUserRows = func(db sqlf.Queryer) ([]*UserRow, error) {
		var users []*UserRow
		if err := selectAllStmt.Select(db, &users); err != nil {
			return nil, err
		}
		return users, nil
	}
}

func checkForError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
