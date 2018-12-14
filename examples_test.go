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
	schemas := []*Schema{
		NewSchema(
			WithDialect(MSSQL),
			WithNamingConvention(SameCase),
		),
		NewSchema(
			WithDialect(MySQL),
			WithNamingConvention(LowerCase),
		),
		NewSchema(
			WithDialect(Postgres),
			WithNamingConvention(SnakeCase),
		),
	}

	// for each schema, print the SQL generated for each statement
	for _, schema := range schemas {
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

func newSession() *Session {
	return nil
}

func ExamplNewSchema() {
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

	schema := NewSchema(
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

	schema := NewSchema(
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

	schema := NewSchema(
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

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleSession_Row() {
	var (
		session *Session
	)

	// start a session
	session = newSession()

	// ExampleRow is an example of a row structure.
	type ExampleRow struct {
		ID    int `sql:"primary key"`
		Name  string
		Value int
	}

	var row = &ExampleRow{
		ID:    1,
		Name:  "first row",
		Value: 10,
	}

	// Insert a row
	result, err := session.Row(row).Exec("insert into examples({}) values({})")
	checkError(err)
	count, err := result.RowsAffected()
	checkError(err)
	log.Printf("row inserted, count=%d", count)

	// Delete the row
	result, err = session.Row(row).Exec("delete from examples where {}")
	checkError(err)
	count, err = result.RowsAffected()
	checkError(err)
	log.Printf("row deleted, count=%d", count)
}

func ExampleSessionRow_Exec() {
	var (
		session *Session
	)

	// start a session
	session = newSession()

	// ExampleRow is an example of a row structure.
	type ExampleRow struct {
		ID    int `sql:"primary key"`
		Name  string
		Value int
	}

	var row = &ExampleRow{
		ID:    1,
		Name:  "first row",
		Value: 10,
	}

	// Performs an insert, bypassing any of the cleverness of the
	// Insert() method.
	//
	// SQL looks something like:
	//  insert into the_table("id", "name", "value") values ($1, $2, $3)
	// Arguments are
	//  [ 1, "first row", 10 ]
	n, err := session.Row(row).Exec("insert into the_table({}) values({})")
	checkError(err)
	log.Printf("row inserted, count=%d", n)

	// Perform an update with an additional test
	//
	// SQL looks like:
	//  update the_table set name = $1, value = $2 where id =$3 and value = $4
	// Arguments are:
	//  [ "first row", 10, 1, 10 ]
	n, err = session.Row(row).Exec("update the_table set {} where {} and value > ?", 10)
	checkError(err)
	log.Printf("row inserted, count=%d", n)
}
