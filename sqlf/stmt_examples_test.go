package sqlf_test

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jjeffery/sqlf/sqlf"
)

func openTestDB() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
		create table users(
			id integer primary key autoincrement,
			given_name text,
			family_name text
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
		insert into users(given_name, family_name)
		values('John', 'Citizen')
	`)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func ExampleInsertRowStmt() {
	type User struct {
		ID         int64 `sql:",primary key auto increment"`
		GivenName  string
		FamilyName string
	}

	stmt := sqlf.NewInsertRowStmt(User{}, `users`)
	fmt.Println(stmt.String())

	var db *sql.DB = openTestDB()

	// Get user with specified primary key
	u := &User{GivenName: "Jane", FamilyName: "Doe"}
	err := stmt.Exec(db, u)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Inserted row: ID=%d\n", u.ID)

	// Output:
	// insert into users(`given_name`,`family_name`) values(?,?)
	// Inserted row: ID=2
}

func ExampleGetRowStmt() {
	type User struct {
		ID         int64 `sql:",primary key auto increment"`
		GivenName  string
		FamilyName string
	}

	stmt := sqlf.NewGetRowStmt(User{}, `users`)
	fmt.Println(stmt.String())

	var db *sql.DB = openTestDB()

	// Get user with specified primary key
	u := &User{ID: 1}
	_, err := stmt.Get(db, u)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ID=%d, GivenName=%q, FamilyName=%q\n",
		u.ID, u.GivenName, u.FamilyName)

	// Output:
	// select `id`,`given_name`,`family_name` from users where `id`=?
	// ID=1, GivenName="John", FamilyName="Citizen"
}

func ExampleExecRowStmt() {
	type User struct {
		ID         int64 `sql:",primary key auto increment"`
		GivenName  string
		FamilyName string
	}

	updateStmt := sqlf.NewUpdateRowStmt(User{}, `users`)
	deleteStmt := sqlf.NewDeleteRowStmt(User{}, `users`)
	fmt.Println(updateStmt.String())
	fmt.Println(deleteStmt.String())

	var db *sql.DB = openTestDB()

	// Get user with specified primary key
	u := &User{ID: 1}
	_, err := sqlf.NewGetRowStmt(User{}, `users`).Get(db, u)
	if err != nil {
		log.Fatal(err)
	}

	// Update row
	u.GivenName = "Donald"
	n, err := updateStmt.Exec(db, u)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("number of rows updated:", n)

	// Delete row
	n, err = deleteStmt.Exec(db, u)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("number of rows deleted:", n)

	// Output:
	// update users set `given_name`=?,`family_name`=? where `id`=?
	// delete from users where `id`=?
	// number of rows updated: 1
	// number of rows deleted: 1
}

func ExampleNewDeleteRowStmt() {
	type User struct {
		ID         int64 `sql:",primary key auto increment"`
		GivenName  string
		FamilyName string
	}

	stmt := sqlf.NewDeleteRowStmt(User{}, `users`)
	fmt.Println(stmt.String())

	// creates a row with ID=1
	var db *sql.DB = openTestDB()

	// Delete user with specified primary key
	u := &User{ID: 1}
	n, err := stmt.Exec(db, u)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("number of rows deleted:", n)

	// Output:
	// delete from users where `id`=?
	// number of rows deleted: 1
}

func ExampleNewUpdateRowStmt() {
	type User struct {
		ID         int64 `sql:",primary key auto increment"`
		GivenName  string
		FamilyName string
	}

	updateStmt := sqlf.NewUpdateRowStmt(User{}, `users`)
	fmt.Println(updateStmt.String())

	var db *sql.DB = openTestDB()

	// Get user with specified primary key
	u := &User{ID: 1}
	_, err := sqlf.NewGetRowStmt(User{}, `users`).Get(db, u)
	if err != nil {
		log.Fatal(err)
	}

	// Update row
	u.GivenName = "Donald"
	n, err := updateStmt.Exec(db, u)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("number of rows updated:", n)

	// Output:
	// update users set `given_name`=?,`family_name`=? where `id`=?
	// number of rows updated: 1
}

func ExampleSelectStmt() {
	type User struct {
		ID      int64 `sql:",primary key auto increment"`
		Login   string
		HashPwd string
		Name    string
	}

	stmt := sqlf.NewSelectStmt(User{}, `
		select distinct {alias u} 
		from users u
		inner join user_search_terms t on t.user_id = u.id
		where t.search_term like ?
	`)
	fmt.Println(stmt.String())

	// Output:
	// select distinct u.`id`,u.`login`,u.`hash_pwd`,u.`name` from users u inner join user_search_terms t on t.user_id = u.id where t.search_term like ?
}
