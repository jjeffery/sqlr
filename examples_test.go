package sqlrow

import (
	"database/sql"
	"fmt"
	"log"
)

var db *sql.DB

func ExampleSchema_Prepare() {
	type UserRow struct {
		ID         int `sql:"primary key autoincrement"`
		GivenName  string
		FamilyName string
	}

	// Define different schemas for different dialects and naming conventions

	mssql := NewSchema(
		WithDialect(MSSQL),
		WithNamingConvention(SameCase),
	)

	mysql := NewSchema(
		WithDialect(MySQL),
		WithNamingConvention(LowerCase),
	)

	postgres := NewSchema(
		WithDialect(Postgres),
		WithNamingConvention(SnakeCase),
	)

	// for each schema, print the SQL generated for each statement
	for _, schema := range []*Schema{mssql, mysql, postgres} {
		stmt, err := schema.Prepare(UserRow{}, `insert into users({}) values({})`)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(stmt)
	}

	// Output:
	// insert into users([GivenName],[FamilyName]) values(?,?)
	// insert into users(`givenname`,`familyname`) values(?,?)
	// insert into users("given_name","family_name") values($1,$2)
}

func ExampleStmt_Exec_insert() {
	type UserRow struct {
		ID         int `sql:"primary key autoincrement"`
		GivenName  string
		FamilyName string
	}

	schema := NewSchema()

	stmt, err := schema.Prepare(UserRow{}, `insert into users({}) values({})`)
	if err != nil {
		log.Fatal(err)
	}

	// ... later on ...

	row := UserRow{
		GivenName:  "John",
		FamilyName: "Citizen",
	}

	_, err = stmt.Exec(db, row)

	if err != nil {
		log.Fatal(err)
	}
}

func ExampleStmt_Exec_update() {
	type UserRow struct {
		ID         int `sql:"primary key autoincrement"`
		GivenName  string
		FamilyName string
	}

	schema := NewSchema()

	stmt, err := schema.Prepare(UserRow{}, `update users set {} where {}`)
	if err != nil {
		log.Fatal(err)
	}

	// ... later on ...

	row := UserRow{
		ID:         42,
		GivenName:  "John",
		FamilyName: "Citizen",
	}

	_, err = stmt.Exec(db, row)

	if err != nil {
		log.Fatal(err)
	}
}

func ExampleStmt_Exec_delete() {
	type UserRow struct {
		ID         int `sql:"primary key autoincrement"`
		GivenName  string
		FamilyName string
	}

	schema := NewSchema()

	stmt, err := schema.Prepare(UserRow{}, `delete from users where {}`)
	if err != nil {
		log.Fatal(err)
	}

	// ... later on ...

	row := UserRow{
		ID:         42,
		GivenName:  "John",
		FamilyName: "Citizen",
	}

	_, err = stmt.Exec(db, row)

	if err != nil {
		log.Fatal(err)
	}
}

func ExampleStmt_Select_oneRow() {
	type UserRow struct {
		ID         int `sql:"primary key autoincrement"`
		GivenName  string
		FamilyName string
	}

	schema := NewSchema()

	stmt, err := schema.Prepare(UserRow{}, `select {} from users where {}`)
	if err != nil {
		log.Fatal(err)
	}

	// ... later on ...

	// find user with ID=42
	var row UserRow
	n, err := stmt.Select(db, &row, 42)
	if err != nil {
		log.Fatal(err)
	}
	if n > 0 {
		log.Printf("found: %v", row)
	} else {
		log.Printf("not found")
	}
}

func ExampleStmt_Select_multipleRows() {
	type UserRow struct {
		ID         int `sql:"primary key autoincrement"`
		GivenName  string
		FamilyName string
	}

	schema := NewSchema()

	stmt, err := schema.Prepare(UserRow{}, `
		select {alias u}
		from users u
		inner join user_search_terms t on t.user_id = u.id
		where t.search_term like ?
		limit ? offset ?`)
	if err != nil {
		log.Fatal(err)
	}

	// ... later on ...

	// find users with search terms
	var rows []UserRow
	n, err := stmt.Select(db, &rows, "smith%", 0, 100)
	if err != nil {
		log.Fatal(err)
	}
	if n > 0 {
		for i, row := range rows {
			log.Printf("row %d: %v", i, row)
		}
	} else {
		log.Printf("not found")
	}
}

/**** obsolete

func ExampleInsert() {
	type UserRow struct {
		ID         int `sql:"primary key autoincrement"`
		GivenName  string
		FamilyName string
	}

	row := UserRow{
		GivenName:  "John",
		FamilyName: "Citizen",
	}
	err := sqlrow.Insert(db, &row, "users")
	if err != nil {
		log.Fatal(err)
	}

	// row.ID will contain the new ID for the row
	log.Printf("Row inserted, ID=%d", row.ID)
}

func ExampleSchema_Insert() {
	type UserRow struct {
		ID         int `sql:"primary key autoincrement"`
		GivenName  string
		FamilyName string
	}

	// Schema for an MSSQL database, where column names
	// are the same as the Go struct field names.
	mssql := sqlrow.Schema{
		Dialect:    sqlrow.DialectFor("mssql"),
		Convention: sqlrow.ConventionSame,
	}

	row := UserRow{
		GivenName:  "John",
		FamilyName: "Citizen",
	}
	err := mssql.Insert(db, &row, "users")
	if err != nil {
		log.Fatal(err)
	}

	// row.ID will contain the new ID for the row
	log.Printf("Row inserted, ID=%d", row.ID)
}

func ExampleUpdate() {
	type UserRow struct {
		ID         int `sql:"primary key autoincrement"`
		GivenName  string
		FamilyName string
	}

	row := UserRow{
		ID:         43,
		GivenName:  "John",
		FamilyName: "Citizen",
	}
	n, err := sqlrow.Update(db, &row, "users")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Number of rows updated = %d", n)
}

func ExampleSchema_Update() {
	type UserRow struct {
		ID         int `sql:"primary key autoincrement"`
		GivenName  string
		FamilyName string
	}

	// Schema for an MSSQL database, where column names
	// are the same as the Go struct field names.
	mssql := sqlrow.Schema{
		Dialect:    sqlrow.DialectFor("mssql"),
		Convention: sqlrow.ConventionSame,
	}

	row := UserRow{
		ID:         43,
		GivenName:  "John",
		FamilyName: "Citizen",
	}
	n, err := mssql.Update(db, &row, "users")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Number of rows updated = %d", n)
}

func ExampleDelete() {
	type UserRow struct {
		ID         int `sql:"primary key autoincrement"`
		GivenName  string
		FamilyName string
	}

	row := UserRow{
		ID:         43,
		GivenName:  "John",
		FamilyName: "Citizen",
	}
	n, err := sqlrow.Delete(db, &row, "users")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Number of rows deleted = %d", n)
}

func ExampleSchema_Delete() {
	type UserRow struct {
		ID         int `sql:"primary key autoincrement"`
		GivenName  string
		FamilyName string
	}

	// Schema for an MSSQL database, where column names
	// are the same as the Go struct field names.
	mssql := sqlrow.Schema{
		Dialect:    sqlrow.DialectFor("mssql"),
		Convention: sqlrow.ConventionSame,
	}

	row := UserRow{
		ID:         43,
		GivenName:  "John",
		FamilyName: "Citizen",
	}
	n, err := mssql.Delete(db, &row, "users")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Number of rows deleted = %d", n)
}

func ExampleDialectFor() {
	// Set the default dialect for PostgreSQL.
	sqlrow.Default.Dialect = sqlrow.DialectFor("postgres")
}

func ExampleSelect_oneRow() {
	type UserRow struct {
		ID         int `sql:"primary key autoincrement"`
		GivenName  string
		FamilyName string
	}

	// find user with ID=42
	var row UserRow
	n, err := sqlrow.Select(db, &row, `select {} from users where ID=?`, 42)
	if err != nil {
		log.Fatal(err)
	}

	if n > 0 {
		log.Printf("found: %v", row)
	} else {
		log.Printf("not found")
	}
}
****/

func ExampleSchema_Select_oneRow() {
	type UserRow struct {
		ID         int `sql:"primary key autoincrement"`
		GivenName  string
		FamilyName string
	}

	// Schema for an MSSQL database, where column names
	// are the same as the Go struct field names.
	mssql := NewSchema(
		WithDialect(MSSQL),
		WithNamingConvention(SameCase),
	)

	// find user with ID=42
	var row UserRow
	n, err := mssql.Select(db, &row, `select {} from [Users] where ID=?`, 42)
	if err != nil {
		log.Fatal(err)
	}

	if n > 0 {
		log.Printf("found: %v", row)
	} else {
		log.Printf("not found")
	}
}

func ExampleSchema_Select_multipleRows() {
	type UserRow struct {
		ID         int `sql:"primary key autoincrement"`
		GivenName  string
		FamilyName string
	}

	// Schema for an MSSQL database, where column names
	// are the same as the Go struct field names.
	mssql := NewSchema(
		WithDialect(MSSQL),
		WithNamingConvention(SameCase),
	)

	// find users with search terms
	var rows []UserRow
	n, err := mssql.Select(db, &rows, `
		select {alias u}
		from [Users] u
		inner join [UserSearchTerms] t on t.UserID = u.ID
		where t.SearchTerm like ?
		limit ? offset ?`, "smith%", 0, 100)
	if err != nil {
		log.Fatal(err)
	}

	if n > 0 {
		for i, row := range rows {
			log.Printf("row %d: %v", i, row)
		}
	} else {
		log.Printf("not found")
	}
}
