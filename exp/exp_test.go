package exp_test

import (
	"database/sql"
	"strings"
	"time"

	"github.com/jjeffery/sqlf/exp"
)

func Example() {
	type User struct {
		ID   int64 `db:",pk autoincr"`
		Name string
	}

	schema := &exp.Schema{}
	schema.DefineTable("users", User{})

	var db *sql.DB
	var user User

	if err := schema.InsertRow(db, user); err != nil {
		panic(err)
	}
	if err := schema.SelectRow(db, &user); err != nil {
		panic(err)
	}
	if _, err := schema.UpdateRow(db, user); err != nil {
		panic(err)
	}
	if _, err := schema.DeleteRow(db, user); err != nil {
		panic(err)
	}

	query1 := schema.MustPrepareQuery(`
select distinct {u}
from users u
inner join user_search_terms t
  on t.user_id = u.id
where t.search_term like ?
order by u.family_name, u.given_name, u.email, u.id`)

	var users []User

	if err := query1.Select(db, &users, "lollypop%"); err != nil {
		panic(err)
	}

	// Output:
}

func ExampleSchema_1() {
	// Many programs use a single database driver and
	// a single database. In this case the DefaultSchema
	// can be used.

	// Set a non-default naming convention for the default schema.
	exp.DefaultSchema.ColumnNameFor = func(field string) string {
		return strings.ToUpper(field)
	}

	type BlogRow struct {
		ID   string
		Name string
	}

	type PostRow struct {
		ID       int64
		BlogID   string
		PostedAt time.Time
	}

	exp.DefineTable("blogs", BlogRow{})
	exp.DefineTable("posts", PostRow{})
}

func ExampleSchema_2() {
	// Example of a less-common scenario, when multiple database
	// schemas are used in the same program.
	//
	// This might be because there are different DB backends in
	// use, or perhaps when the same database has a number of different
	// naming conventions and it is easier to handle via different
	// schemas.

	// create a schema for mysql
	mysql := &exp.Schema{
		Dialect: exp.NewDialect("mysql"),
	}

	// schema for a sqlite3 database
	sqlite := &exp.Schema{
		Dialect: exp.NewDialect("sqlite3"),

		// naming convention: column name is identical to field name
		ColumnNameFor: func(field string) string { return field },
	}

	type Row1 struct {
		ID   string
		Name string
	}

	type Row2 struct {
		ID        int64
		Name      string
		UpdatedAt time.Time
	}

	mysql.DefineTable("table1", Row1{})
	sqlite.DefineTable("table1", Row1{})
	sqlite.DefineTable("table2", Row2{})
}
