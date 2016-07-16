# sqlf: Formatting SQL statements

Package `sqlstmt` provides assistance in creating SQL statements. 

[![GoDoc](https://godoc.org/github.com/jjeffery/sqlstmt?status.svg)](https://godoc.org/github.com/jjeffery/sqlstmt)
[![License](http://img.shields.io/badge/license-MIT-green.svg?style=flat)](https://raw.githubusercontent.com/jjeffery/sqlstmt/master/LICENSE.md)
[![GoReportCard](https://goreportcard.com/badge/github.com/jjeffery/sqlstmt)](https://goreportcard.com/report/github.com/jjeffery/sqlstmt)

**NOTE:** This package is still a work in progress. The API is reasonably stable, but there is 
no backwards compatibility guarantee at this time.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Rationale](#rationale)
- [Obtaining the package](#obtaining-the-package)
- [Examples](#examples)
  - [Inserting a row](#inserting-a-row)
  - [Updating a row](#updating-a-row)
  - [Deleting a row](#deleting-a-row)
  - [Getting a row by primary key](#getting-a-row-by-primary-key)
  - [Performing queries](#performing-queries)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


## Rationale

Package `sqlstmt` aims to make it easy to construct and execute SQL
statements for common scenarios. Supported scenarios include:

* Insert a single row based on a Go struct
* Update a single row based on a Go struct
* Delete a single row based on a Go struct
* Select a single row into a Go struct
* Select zero, one or more rows int a a slice of Go structs

This package is intended for programmers who are comfortable with
writing SQL, but would like assistance with the sometimes tedious
process of preparing SELECT, INSERT, UPDATE and DELETE statements
for tables that have a large number of columns.

This package is designed to work seamlessly with the standard library
"database/sql" package. It does not provide any layer on top of *sql.DB
or *sql.Tx. If the calling program has a need to execute queries independently
of this package, it can use "database/sql" directly, or make use of any other
third party package that uses "database/sql".

The philosophy behind the design if the `sqlstmt` API is:

* Simple, single-row CRUD operations should be easy to construct
* Slightly more complex operations should be possible with only a little more effort
* Handle different SQL dialects and naming conventions using simple interfaces
* Support popular SQL dialects out of the box with minimal setup
* Easily fallback to using `database/sql` and other third-party packages for any functionality that is not handled by this package

## Obtaining the package

```bash
go get github.com/jjeffery/sqlstmt
```

Note that if you are interested in running the unit tests, you will need
package `github.com/mattn/sqlite3`, which requires cgo and a C compiler
setup to compile correctly.

## Examples

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

## SQL dialects

The `sqlstmt` package is designed to be as SQL-agnostic as possible, but 
when it is generating SQL it does need to know the following:

* How to quote column names to ensure they are not interpreted as an SQL keyword
  * PostgreSQL uses double quotes: `"column_name"`
  * MySQL uses back ticks: `\`column_name\``
  * MS SQL Server uses square braces: `[column_name]`
* How to write placeholders for arguments
  * PostgreSQL uses numbered placeholders: `$1`, `$2`, etc
  * Almost everyone else uses question marks: `?`

Most programs use only one SQL driver, and in these circumstances `sqlstmt`
will do the right thing.

If a program is using PostgreSQL, it will load the appropriate driver somewhere,
probably in the `main` package:

```go
import _ "github.com/lib/pq"
```

By default `sqlstmt` will check the list of loaded SQL drivers and pick the
first one to decide on the SQL dialect to use. In this example it will
automatically choose the "postgres" dialect.

### Specifying the SQL dialect

If your program references multiple SQL drivers, it may be necesary to 
specify which dialect is in use. This can be done when opening the 
database connection:

```go
// open the database
db, err := sql.Open("postgres", "user=test dbname=test sslmode=disable")
if err != nil {
	log.Fatal(err)
}

// specify the dialect in use
sqlstmt.DefaultSchema.Dialect = sqlstmt.NewDialect("postgres")
```

### Using multiple dialects

If your program makes use of multiple database backends and you want
to use `sqlstmt` for both of them, the best thing to do is to specify
a `sqlstmt.Schema` for each of the database backends.

```go
var pgSchema = &sqlstmt.Schema{
	Dialect: sqlstmt.NewDialect("postgres"),
}

var mysqlSchema = &sqlstmt.Schema{
	Dialect: sqlstmt.NewDialect("mysql"),
}
```

When the time comes to create statements, use the appropriate schema:

```go
// insert statement for widgets in postgres database
var insertWidget = pgSchema.NewInsertRowStmt(Widget{}, "widgets")

// update statement for gadgets in mysql database
var updateGadget = mysqlSchema.NewUpdateRowStmt(Gadget{}, "gadgets")
```