# OBSOLETE: keep for now to include relevant parts in new documention.

# sqlrow: SQL statements

Package `sqlrow` provides assistance in creating SQL statements. 

[![GoDoc](https://godoc.org/github.com/jjeffery/sqlrow?status.svg)](https://godoc.org/github.com/jjeffery/sqlrow)
[![License](http://img.shields.io/badge/license-MIT-green.svg?style=flat)](https://raw.githubusercontent.com/jjeffery/sqlrow/master/LICENSE.md)
[![Build Status (Linux)](https://travis-ci.org/jjeffery/sqlrow.svg?branch=master)](https://travis-ci.org/jjeffery/sqlrow)
[![Coverage Status](https://coveralls.io/repos/github/jjeffery/sqlrow/badge.svg?branch=master)](https://coveralls.io/github/jjeffery/sqlrow?branch=master)
[![GoReportCard](https://goreportcard.com/badge/github.com/jjeffery/sqlrow)](https://goreportcard.com/report/github.com/jjeffery/sqlrow)

**NOTE:** This package is still a work in progress. The API is reasonably stable, but there is 
no backwards compatibility guarantee at this time.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Overview](#overview)
- [Obtaining the package](#obtaining-the-package)
- [Examples](#examples)
  - [Inserting a row](#inserting-a-row)
  - [Updating a row](#updating-a-row)
  - [Deleting a row](#deleting-a-row)
  - [Getting a row by primary key](#getting-a-row-by-primary-key)
  - [Performing queries](#performing-queries)
- [SQL dialects](#sql-dialects)
  - [The default dialect](#the-default-dialect)
  - [Specifying the SQL dialect](#specifying-the-sql-dialect)
  - [Using multiple dialects](#using-multiple-dialects)
- [Column mapping](#column-mapping)
  - [Simple structs](#simple-structs)
  - [Anonymous structs](#anonymous-structs)
  - [Embedded structs](#embedded-structs)
- [Column naming conventions](#column-naming-conventions)
- [Contributing](#contributing)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


## Overview

Package `sqlrow` aims to make it easy to construct and execute SQL
statements for common scenarios. Supported scenarios include:

* Insert, update and delete a single row based on the contents of a Go struct
* Select a single row into a Go struct
* Select zero, one or more rows into a slice of Go structs

This package is intended for programmers who are comfortable with
writing SQL, but would like assistance with the sometimes tedious
process of preparing SELECT, INSERT, UPDATE and DELETE statements
for tables that have a large number of columns. It is designed to 
work seamlessly with the standard library `database/sql` package
in that it does not provide any layer on top of `*sql.DB` or `*sql.Tx`. 
If the calling program has a need to execute queries independently
of this package, it can use `database/sql` directly, or make use of 
any other third party package.

The philosophy behind the design of the `sqlrow` API is:

* Simple, single-row CRUD operations should be easy to construct
* Slightly more complex operations should be possible with only a little more effort
* Support popular SQL dialects out of the box; provide for further customization through simple interfaces
* Easily fallback to using `database/sql` and other third-party packages for any functionality that is not handled by this package

## Obtaining the package

```bash
go get github.com/jjeffery/sqlrow
```

Note that if you are interested in running the unit tests, you will need
package `github.com/mattn/sqlite3`, which requires cgo and a C compiler
setup to compile correctly.

## Examples

Note that there are more examples in the 
[GoDoc](https://godoc.org/github.com/jjeffery/sqlrow) documentation.

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
	ID           int `sql:"primary key autoincrement"`
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

This example code is using SQLite, but the `sqlrow` package supports any
SQL dialect via a very simple `Dialect` interface, and the following SQL
dialects are supported out of the box:

* SQLite
* PostgreSQL
* MySQL
* MS SQL

### Inserting a row

```go

// create the row object and populate with data
u := &User{
	GivenName:    "Jane",
	FamilyName:   "Citizen",
	EmailAddress: "jane@citizen.com",
}

// insert the row into the `users` table using the db connection opened earlier
err := sqlrow.Insert(db, u, "users")

if err != nil {
	log.Fatal(err)
}

fmt.Println("User ID:", u.ID)

// Output: User ID: 1
```

Because the `id` column is an auto-increment column, the value of `u.ID` will
contain the auto-generated value after the insert row statement has been executed.

> Note that the Postgres driver `github.com/lib/pq` does not support
> the `Result.LastInsertId` method, and so this feature does not work for that
> driver. See the `pq` package [GoDoc](http://godoc.org/github.com/lib/pq) for
> a work-around.

### Updating a row

Continuing from the previous example:

```go
// change user details
u.EmailAddress = "jane.citizen.314159@gmail.com"

// update the row in the `users` table
n, err = sqlrow.Update(db, u, "users")

if err != nil {
	log.Fatal(err)
}

fmt.Println("Number of rows updated:", n)

// Output: Number of rows updated: 1
```

### Deleting a row

Continuing from the previous example:

```go
// execute the row in the `users` table
n, err = sqlrow.Delete(db, u, "users")

if err != nil {
	log.Fatal(err)
}

fmt.Println("Number of rows deleted:", n)

// Output: Number of rows deleted: 1
```

### Getting a row by primary key

Pretending that we have not deleted the row in the previous example:

```go
var user User 

n, err := sqlrow.Select(db, &user, "users", 1)

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

Performing a query that returns zero, one or more rows involves
writing some SQL. The `sqlrow` package provides an extended syntax that 
provides a shorthand alternative to explicitly listing all columns and 
parameter placeholders.

```go
// declare a slice of users for receiving the result of the query
var users []*User

// perform the query, specifying an argument for each of the
// placeholders in the SQL query
_,  err = sqlrow.Select(db, &users, `
        select {}
		from users
		where family_name = ?`, "Citizen")
if err != nil {
	log.Fatal(err)
}

// at this point, the users slice will contain one object for each
// row returned by the SQL query
for _, u := range users {
	doSomethingWith(u)
}
```

Note the non-standard `{}` in the SQL query above. The `sqlrow` statement
knows to substitute in column names in the appropriate format. In the 
example above, the SQL generated will look like the following:

```sql
select `id`,`family_name`,`given_name`,`email_address`
from users
where family_name = ?
```

For queries that involve multiple tables, it is always a good idea to
use table aliases:

```go
// declare a slice of users for receiving the result of the query
var users []*User

// perform the query, specifying an argument for each of the
// placeholders in the SQL query
_, err = sqlrow.Select(db, &users, `
      	select {alias u}
		from users u
	   	inner join user_search_terms t
			on t.user_id = u.id
		where u.term like ?`, `Cit%`)
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

The `sqlrow` package is designed to be as SQL-agnostic as possible, but 
when it is generating SQL it does need to know the following:

* How to quote column names to ensure they are not interpreted as an SQL keyword
  * PostgreSQL uses double quotes: "column_name"
  * MySQL uses back ticks: \`column_name\`
  * MS SQL Server uses square braces: [column_name]
* How to write placeholders for arguments
  * PostgreSQL uses numbered placeholders: `$1`, `$2`, etc
  * Almost everyone else uses question marks: `?`

### The default dialect

Most programs use only one SQL driver, and in these circumstances `sqlrow`
will do the right thing.

For example, if a program is using Postgres, it will need to load the appropriate driver,
probably in the `main` package:

```go
import _ "github.com/lib/pq"
```

By default `sqlrow` will check the list of loaded SQL drivers and pick the
first one to decide on the SQL dialect to use. If only one SQL driver has been
loaded, it will choose correctly. In this example it will automatically choose 
the "postgres" dialect.

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
sqlrow.Default.Dialect = sqlrow.DialectFor("postgres")
```

### Using multiple dialects

If your program makes use of multiple database connections with different
types of server, the best thing to do is to specify a `sqlrow.Schema` 
for each of the databases.

```go
var postgresDB = &sqlrow.Schema{
	Dialect: sqlrow.NewDialect("postgres"),
}

var mysqlDB = &sqlrow.Schema{
	Dialect: sqlrow.NewDialect("mysql"),
}
```

When the time comes to create statements, use the appropriate schema:

```go
// insert widgets in postgres database
err = postgresDB.Insert(db1, &widget, "widgets")

// update statement for gadgets in mysql database
_, err = mysqlDB.Update(db2, &gadget, "gadgets")
```

## Column mapping

For each statement, the `sqlrow` package generates column names based on
a Go struct.

### Simple structs

```go
type User struct {
	ID        int64 `sql:"primary key"`
	Name      string
	UpdatedAt time.Time
	CreatedAt time.Time
}

// Column names generated by sqlrow:
// * id
// * name
// * updated_at
// * created_at
```

Note the use of the struct tag to specify the primary key. The struct tag
can also be used to override the column name:

```go
type User struct {
	ID        int64     `sql:"user_id primary key"`
	Name      string
	UpdatedAt time.Time
	CreatedAt time.Time
	DOB       time.Time `sql:"date_of_birth"`
}

// Column names generated by sqlrow:
// * user_id
// * name
// * updated_at
// * created_at
// * date_of_birth
```

If you need to override the column name to be an SQL keyword, (which is
rarely a good idea), you can use quotes to specify the column name.

```go
// Not recommended
type User struct {
	ID int64 `sql:"'primary' primary key"` // setting column name to SQL keyword
	// ... rest of struct here
}
```

### Anonymous structs

Sometimes there are a set of common columns, used by each table.
Anonymous structs are a convenient way to ensure consistency across
the Go structs:

```go
type Entity struct {
	ID        int64 `sql:"primary key autoincrement"`
	UpdatedAt time.Time
	CreatedAt time.Time
}

type User struct {
	Entity
	Name  string
	Email string
}

// Column names generated by sqlrow:
// * id
// * updated_at
// * created_at
// * name
// * email

type Vehicle struct {
	Entity
	Make string
	Model string
}

// Column names generated by sqlrow:
// * id
// * updated_at
// * created_at
// * make
// * model

```

### Embedded structs

In some cases it is useful to use embedded structures when representing
components in a structure.

```go

type Address struct {
	Street   string
	Locality string
	City     string
	Postcode string
	Country  string
}

type CustomerContact struct {
	CustomerID    int64 `sql:"primary key"`
	HomeAddress   Address
	PostalAddress Address
}

// Column names generated by sqlrow:
// * id
// * home_address_street
// * home_address_locality
// * home_address_city
// * home_address_postcode
// * home_address_country
// * postal_address_street
// * postal_address_locality
// * postal_address_city
// * postal_address_postcode
// * postal_address_country
```

## Column naming conventions

The `sqlrow` package has a default naming convention which will convert
a Go field name like `HomeAddress` into it's "snake case" equivalent:
`home_address`. This is a popular common naming convention and is supported
by default by Active Record and other popular ORM frameworks.

If this naming convention does not suit, you can override by providing an
implementation of the `Convention` interface:

```go
// Convention provides naming convention methods for
// inferring a database column name from Go struct field names.
type Convention interface {
	// The name of the convention. This can be used as
	// a key for caching, so if If two conventions have
	// the same name, then they should be identical.
	Name() string
	
	// ColumnName returns the name of a database column based
	// on the name of a Go struct field.
	ColumnName(fieldName string) string

	// Join joins a prefix with a name to form a column name.
	// Used for naming columns based on fields within embedded
	// structures. The column name will be based on the name of
	// the Go struct field and its enclosing embedded struct fields.
	Join(prefix, name string) string
}
```

The `ColumnName` method accepts a field name (eg "HomeAddress") and
returns the associated column name.

The `Join` method is used for embedded structures. It joins a
prefix (for example "home_address") with a name (eg "street") to 
produce a joined name (eg "home_address_street").

The `sqlrow` package comes with two naming conventions out of the box:

* `ConventionSnake`: the default, "snake_case" convention; and
* `ConventionSame`: a convention where the column name is identical to the Go field name.

To set a convention other than the default, set the `Schema.Convention` property:

```go
// set the default naming convention so that column names are
// the same as Go struct field names
sqlrow.Default.Convention = sqlrow.ConventionSame

// create a new schema with it's own naming convention
mySchema := &sqlrow.Schema{
	Convention: newMyCustomNamingConvention(),
}

// This will use the default convention (which is now sqlrow.ConventionSame)
err := sqlrow.Insert(db1, widget, "widgets")

// This will use the custom convention associated with mySchema
_, err = mySchema.Update(db2, gadget, "gadgets")
```

## Contributing

Pull requests are welcome. Please include tests providing test coverage 
for your changes.

If you are raising an issue that describes a bug, please include a minimal
example that reproduces the bug.
