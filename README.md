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

### Inserting a row

```go

// create the statement -- this only needs to be done once
insertRow := sqlstmt.NewInsertStmt(User{}, "users")

// perform the insertStmt using a db connection opened earlier
err := insertRow.Exec(db, &User{
	GivenName: "Jane",
	FamilyName: "Citizen",
	EmailAddress: "jane@citizen.com",
})
```

