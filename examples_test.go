package sqlr

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
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
	// insert into users([GivenName], [FamilyName]) values(?, ?)
	// insert into users(`givenname`, `familyname`) values(?, ?)
	// insert into users("given_name", "family_name") values($1, $2)
}

func ExampleStmt_Exec_insert() {
	type UserRow struct {
		ID         int `sql:"primary key autoincrement"`
		GivenName  string
		FamilyName string
	}

	schema := NewSchema()
	ctx := context.Background()

	stmt, err := schema.Prepare(UserRow{}, `insert into users({}) values({})`)
	if err != nil {
		log.Fatal(err)
	}

	// ... later on ...

	row := UserRow{
		GivenName:  "John",
		FamilyName: "Citizen",
	}

	_, err = stmt.Exec(ctx, db, row)

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
	ctx := context.Background()

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

	_, err = stmt.Exec(ctx, db, row)

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
	ctx := context.Background()

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

	_, err = stmt.Exec(ctx, db, row)

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
	ctx := context.Background()

	stmt, err := schema.Prepare(UserRow{}, `select {} from users where {}`)
	if err != nil {
		log.Fatal(err)
	}

	// ... later on ...

	// find user with ID=42
	var row UserRow
	n, err := stmt.Select(ctx, db, &row, 42)
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
	ctx := context.Background()

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
	n, err := stmt.Select(ctx, db, &rows, "smith%", 0, 100)
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

func ExampleWithIdentifier() {
	// Take an example of a program that operates against an SQL Server
	// database where a table is named "[User]", but the same table is
	// named "users" in the Postgres schema.
	mssql := NewSchema(
		WithDialect(MSSQL),
		WithNamingConvention(SameCase),
		WithIdentifier("[User]", "users"),
		WithIdentifier("UserId", "user_id"),
		WithIdentifier("[Name]", "name"),
	)
	postgres := NewSchema(
		WithDialect(Postgres),
		WithNamingConvention(SnakeCase),
	)

	type User struct {
		UserId int `sql:"primary key"`
		Name   string
	}

	// If a statement is prepared and executed for both
	const query = "select {} from users where user_id = ?"

	mssqlStmt, err := mssql.Prepare(User{}, query)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mssqlStmt)
	postgresStmt, err := postgres.Prepare(User{}, query)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(postgresStmt)

	// Output:
	// select [UserId], [Name] from [User] where UserId = ?
	// select "user_id", "name" from users where user_id = $1
}

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
		limit ? offset ?`, "smith%", 100, 0)
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

func ExampleSession_MakeQuery() {
	var schema Schema
	ctx := context.Background()
	tx := beginTransaction() // get a DB transaction, assumes no errors
	defer tx.Commit()        // WARNING: no error handling here: example code only

	type Row struct {
		ID            int64 `sql:"primary key"`
		Name          string
		FavoriteColor string
	}

	// begin a request-scoped database session
	sess := NewSession(ctx, tx, &schema)

	// data access object
	var dao struct {
		Get    func(id int64) (*Row, error)
		Select func(query string, args ...interface{}) ([]*Row, error)
	}

	sess.MakeQuery(&dao.Get, &dao.Select)

	// can now use the type-safe data access functions
	row42, err := dao.Get(42)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Row 42:", row42)

	redRows, err := dao.Select("select {} from rows where favorite_color = ?", "red")
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range redRows {
		log.Println("Likes red:", row)
	}
}

func beginTransaction() *sql.Tx {
	return nil
}

func ExampleMustCreateSchema() {
	type UserRow struct {
		ID   int64 `sql:"primary key"`
		Name string
	}

	type PostRow struct {
		ID        int64 `sql:"primary key"`
		UserID    int64
		CreatedAt time.Time
		Title     string
		Content   string
	}

	schema := MustCreateSchema(
		WithNamingConvention(SnakeCase),
		WithTables(TablesConfig{
			(*UserRow)(nil): {
				TableName: "users_table", // override naming convention for table
			},
			(*PostRow)(nil): {
				Columns: ColumnsConfig{
					"CreatedAt": {
						ColumnName: "create_timestamp", // override naming convention
						EmptyNull:  true,               // store empty string as null
					},
				},
			},
		}),
	)

	doSomethingWith(schema)
}

func ExampleColumnConfig() {
	type Address struct {
		Street   string
		Locality string
		City     string
		State    string
		Country  string
	}

	type PersonRow struct {
		ID      int64 `sql:"primary key"`
		Name    string
		DOB     time.Time
		Address Address
	}

	schema := MustCreateSchema(
		WithTables(TablesConfig{
			(*PersonRow)(nil): {
				Columns: ColumnsConfig{
					"Address.Locality": {
						ColumnName: "address_suburb",
					},
					"DOB": {
						ColumnName: "date_of_birth",
						EmptyNull:  true,
					},
				},
			},
		}),
	)

	doSomethingWith(schema)
}

func ExampleColumnsConfig() {
	type Address struct {
		Street   string
		Locality string
		City     string
		State    string
		Country  string
	}

	type PersonRow struct {
		ID      int64 `sql:"primary key"`
		Name    string
		DOB     time.Time
		Address Address
	}

	schema := MustCreateSchema(
		WithTables(TablesConfig{
			(*PersonRow)(nil): {
				Columns: ColumnsConfig{
					"Address.Locality": {
						ColumnName: "address_suburb",
					},
					"DOB": {
						ColumnName: "date_of_birth",
						EmptyNull:  true,
					},
				},
			},
		}),
	)

	doSomethingWith(schema)
}

func doSomethingWith(v interface{}) {

}
