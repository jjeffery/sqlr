package sqlstmt_test

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/jjeffery/sqlstmt"
	_ "github.com/mattn/go-sqlite3"
)

// The UserRow struct represents a single row in the users table.
// Note that the sqlstmt package becomes more useful when tables
// have many more columns than shown in this example.
type UserRow struct {
	ID         int64 `sql:",primary key autoincrement"`
	GivenName  string
	FamilyName string
}

func Example() {
	db, err := sql.Open("sqlite3", ":memory:")
	exitIfError(err)
	setupSchema(db)

	insertRow := sqlstmt.NewInsertRowStmt(UserRow{}, `users`)
	updateRow := sqlstmt.NewUpdateRowStmt(UserRow{}, `users`)
	deleteRow := sqlstmt.NewDeleteRowStmt(UserRow{}, `users`)
	getRow := sqlstmt.NewGetRowStmt(UserRow{}, `users`)
	selectAllRows := sqlstmt.NewSelectStmt(UserRow{}, `
		select {}
		from users
		order by id
	`)

	tx, err := db.Begin()
	exitIfError(err)
	defer tx.Rollback()

	// insert three rows, IDs are automatically generated (1, 2, 3)
	for _, givenName := range []string{"John", "Jane", "Joan"} {
		u := &UserRow{
			GivenName:  givenName,
			FamilyName: "Citizen",
		}
		err = insertRow.Exec(tx, u)
		exitIfError(err)
	}

	// get user with ID of 3 and then delete it
	{
		u := &UserRow{ID: 3}
		_, err = getRow.Get(tx, u)
		exitIfError(err)

		_, err = deleteRow.Exec(tx, u)
		exitIfError(err)
	}

	// update family name for user with ID of 2
	{
		u := &UserRow{ID: 2}
		_, err = getRow.Get(tx, u)
		exitIfError(err)

		u.FamilyName = "Doe"
		_, err = updateRow.Exec(tx, u)
		exitIfError(err)
	}

	// select rows from table and print
	{
		var users []*UserRow
		err = selectAllRows.Select(tx, &users)
		exitIfError(err)
		for _, u := range users {
			fmt.Printf("User %d: %s, %s\n", u.ID, u.FamilyName, u.GivenName)
		}
	}

	// Output:
	// User 1: Citizen, John
	// User 2: Doe, Jane
}

func exitIfError(err error) {
	if err != nil {
		log.Output(2, err.Error())
		os.Exit(1)
	}
}

func init() {
	log.SetFlags(0)

	// uncomment to log SQL statements
	//sqlstmt.DefaultSchema.Logger = log.New(os.Stderr, "sqlstmt: ", log.Flags())
}

func setupSchema(db *sql.DB) {
	_, err := db.Exec(`
		create table users(
			id integer primary key autoincrement,
			given_name text,
			family_name text
		)
	`)
	exitIfError(err)
}
