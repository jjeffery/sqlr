# sqlf: Formatting SQL statements

Package `sqlstmt` provides assistance in creating SQL statements. 

[![GoDoc](https://godoc.org/github.com/jjeffery/sqlstmt?status.svg)](https://godoc.org/github.com/jjeffery/sqlstmt)
[![License](http://img.shields.io/badge/license-MIT-green.svg?style=flat)](https://raw.githubusercontent.com/jjeffery/sqlstmt/master/LICENSE.md)
[![GoReportCard](https://goreportcard.com/badge/github.com/jjeffery/sqlstmt)](https://goreportcard.com/report/github.com/jjeffery/sqlstmt)

**NOTE:** This package is a work in progress. There is 
no backwards compatibility guarantee at this time.

## Rationale

There are a number of ORM packages for the Go language, with varying
sets of features. There are times, however, when an ORM may not be 
appropriate, and an SQL-based approach might provide the desired simplicity,
control, and performance.

Using an SQL-based API such as `database/sql`, however, can be a little tedious
and error prone. There are some popular packages available that make working
with `database/sql` easier &mdash; a good example  is `github.com/jmoiron/sqlx`.

While packages such as `sqlx` go a long way towards handling the results
of SQL queries, it can still be quite tedious to construct the SQL for a
query in the first place. This is particularly so for queries against
database tables that have many columns and hence many placeholders (ie `?`)
in the SQL -- it can be error-prone constructing and maintaining the API 
calls to have the correct number of arguments in the correct order.

Package `sqlstmt` attempts to solve this problem by enabling construction of
SQL statements using an API based on the contents of Go language structures.
The philosophy behind the design if the `sqlstmt` API is:

1. Simple CRUD operations should be very easy to construct.
2. Slightly more complex operations should be possible with a little more effort.
3. Fallback to using `database/sql` for queries that are too complex to be handled easily by this package.

## Obtaining the package

```bash
go get github.com/jjeffery/sqlstmt
```

Note that if you are interested in running the unit tests, you will need
package `github.com/mattn/sqlite3`, which requires cgo and a C compiler
setup to compile correctly.

## Example

Note that there are more examples in the 
[GoDoc](https://godoc.org/github.com/jjeffery/sqlstmt) documentation.

The following examples use a fairly simple database schema. Note that
this package becomes much more useful for database schemas where tables
have many columns (and hence the row structs have many fields).

```sql
create table users(
	id integer primary key autoincrement,
	given_name text
	family_name text
	email_address text
)
```

A corresponding Go struct for representing a row in the `users` table is:

```go
type User struct {
	ID           int `sql:primary key autoincrement`
	GivenName    string
	FamilyName   string
	EmailAddress string
}
```

Note the use of struct tags to include information about the primary key
and auto-increment behaviour.

The following examples assume that a database has been opened and the 
`*sql.DB` is stored in variable `db`:

```go
db, err := sql.Open("sqlite3", ":memory:")
if err != nil {
	log.Fatal(err)
}
```

This example code is using SQLite, but the `sqlstmt` package supports any
SQL dialect via a very simple `Dialect` interface, and the following SQL
dialects are supported out of the box:

* SQLite
* PostgreSQL
* MySQL
* MS SQL

### Inserting a row

```go

// create the statement -- this only needs to be done once, at
// program initialization time
insertRow := sqlstmt.NewInsertRowStmt(User{}, "users")

// create the row object and populate with data
u := &User{
	GivenName: "Jane",
	FamilyName: "Citizen",
	EmailAddress: "jane@citizen.com",
}

// execute the insert statement using a db connection opened earlier
err := insertRow.Exec(db, u)

if err != nil {
	log.Fatal(err)
}

fmt.Println("User ID:", u.ID)

// Output: User ID: 1
```

Because the `id` column is an auto-increment column, the value of `u.ID` will
contain the auto-generated value after the insert row statement has been executed.

### Updating a row

Continuing from the previous example:

```go
// create the statement -- this only needs to be done once, at
// program initialization time
updateRow := sqlstmt.NewUpdateRowStmt(User{}, "users")

// change user details
u.EmailAddress = "jane.citizen.314159@gmail.com"

// execute the update statement
n, err = updateRow.Exec(db, u)

if err != nil {
	log.Fatal(err)
}

fmt.Println("Number of rows updated:", n)

// Output: Number of rows updated: 1
```

### Deleting a row

Continuing from the previous example:

```go
// create the statement -- this only needs to be done once, at
// program initialization time
deleteRow := sqlstmt.NewDeleteRowStmt(User{}, "users")

// execute the delete statement
n, err = updateRow.Exec(db, u)

if err != nil {
	log.Fatal(err)
}

fmt.Println("Number of rows deleted:", n)

// Output: Number of rows deleted: 1
```

### Getting a row by primary key

Pretending that we have not deleted the row in the previous example:

```go
getRow := sqlstmt.NewGetRowStmt(User{}, "users")

// create a row variable and populate with the primary key of the row
// that we are after
u := &User{ID: 1}

n, err := getRow.Get(db, u)

if err != nil {
	log.Fatal(err)
}

fmt.Println("Rows returned:", n)
fmt.Println("User email:", u.EmailAddress)

// Output:
// Rows returned: 1
// User email: jane.citizen.314159@gmail.com
```

### Performing queries

Performing a query that returns zero, one or more rows usually involves
writing some SQL, and this is where it becomes necessary to write some
SQL. The `sqlstmt` package provides an extended syntax that is shorthand
for having to explicitly list all columns and SQL placeholders.

```go
familyNameQuery := sqlstmt.NewSelectStmt(User{}, `
	select {}
	from users
	where family_name = ?
`)

// declare a slice of users for receiving the result of the query
var users []User

// perform the query, specifying an argument for each of the
// placeholders in the SQL query
err = familyNameQuery.Select(db, &users, "Citizen")
if err != nil {
	log.Fatal(err)
}

// at this point, the users slice will contain one object for each
// row returned by the SQL query
for _, u := range users {
	doSomethingWith(u)
}
```

Note the non-standard `{}` in the SQL query above. The `sqlstmt` statement
knows to substitute in column names in the appropriate format. In the 
example above, the SQL generated will look like the following:

```sql
select `id`,`family_name`,`given_name`,`email_address`
from users
where family_name = ?
```

For queries that involve multiple tables, it is always a good idea to
use table aliases when specifying tables:

```go
searchTermQuery := sqlstmt.NewSelectStmt(User{}, `
	select {alias u}
	from users u
	inner join user_search_terms t
	  on t.user_id = u.id
	where u.term like ?
`)

// declare a slice of users for receiving the result of the query
var users []User

// perform the query, specifying an argument for each of the
// placeholders in the SQL query
err = searchTermQuery.Select(db, &users, "Cit%")
if err != nil {
	log.Fatal(err)
}

for _, u := range users {
	doSomethingWith(u)
}
```

The SQL generated in this example looks like the following:
```sql
select u.`id`,u.`family_name`,u.`given_name`,u.`email_address`
from users u
inner join user_search_terms t
  on t.user_id = u.id
where u.term like ?
```
