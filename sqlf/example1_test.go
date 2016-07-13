package sqlf_test

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jjeffery/sqlf/sqlf"
	_ "github.com/mattn/go-sqlite3"
)

// The UserRow struct represents a single row in the users table.
// Note that the sqlf package becomes more useful when tables
// have many more columns than shown in this example.
type UserRow struct {
	ID         int64 `sql:",primary key auto increment"`
	GivenName  string
	FamilyName string
}

func Example() {
	db, err := sql.Open("sqlite3", ":memory:")
	exitIfError(err)

	insertRowStmt := sqlf.PrepareInsertRow(UserRow{}, `
		insert into users({columns})
		values({values})
	`)
	updateRowStmt := sqlf.PrepareUpdateRow(UserRow{}, `
		update users
		set {set}
		where {where}
	`)
	// A statement for deleting one row is prepared using the
	// same function as a statement updating one row.
	deleteRowStmt := sqlf.PrepareUpdateRow(UserRow{}, `
		delete from users
		where {where}
	`)
	getRowStmt := sqlf.PrepareGetRow(UserRow{}, `
		select {columns}
		from users
		where {where}
	`)
	selectAllRowsStmt := sqlf.PrepareSelectRows(UserRow{}, `
		select {columns}
		from users
		order by id
	`)

	tx, err := db.Begin()
	exitIfError(err)
	defer tx.Rollback()

	// insert three rows, IDs are automatically generated (1, 2, 3)
	for _, givenName := range []string{"John", "Joan", "Jane"} {
		u := &UserRow{
			GivenName:  givenName,
			FamilyName: "Citizen",
		}
		err = insertRowStmt.Exec(tx, u)
		exitIfError(err)
	}

	// get user with ID of 3 and then delete it
	{
		u := &UserRow{ID: 3}
		_, err = getRowStmt.Get(tx, u)
		exitIfError(err)

		_, err = deleteRowStmt.Exec(tx, u)
		exitIfError(err)
	}

	// update family name for user with ID of 2
	{
		u := &UserRow{ID: 2}
		_, err = getRowStmt.Get(tx, u)
		exitIfError(err)

		u.FamilyName = "Doe"
		_, err = updateRowStmt.Exec(tx, u)
		exitIfError(err)
	}

	// select rows from table and print
	{
		var users []*UserRow
		err = selectAllRowsStmt.Select(tx, &users)
		exitIfError(err)
		for _, u := range users {
			fmt.Printf("User %d: %s, %s", u.ID, u.FamilyName, u.GivenName)
		}
	}

	// Output:
	// User 1: Citizen, John
	// User 2: Doe, Jane
}

func exitIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	log.SetFlags(log.Lshortfile)
}
