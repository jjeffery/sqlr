# sqlr: SQL API for Go

[![GoDoc](https://godoc.org/github.com/jjeffery/sqlr?status.svg)](https://godoc.org/github.com/jjeffery/sqlr)
[![Documentation](https://img.shields.io/badge/documentation-reference-blue.svg)](https://jjeffery.github.io/sqlr)
[![License](http://img.shields.io/badge/license-MIT-green.svg?style=flat)](https://raw.githubusercontent.com/jjeffery/sqlr/master/LICENSE.md)
[![Build Status (Linux)](https://travis-ci.org/jjeffery/sqlr.svg?branch=master)](https://travis-ci.org/jjeffery/sqlr)
[![Coverage Status](https://codecov.io/github/jjeffery/sqlr/badge.svg?branch=master)](https://codecov.io/github/jjeffery/sqlr?branch=master)
[![GoReportCard](https://goreportcard.com/badge/github.com/jjeffery/sqlr)](https://goreportcard.com/report/github.com/jjeffery/sqlr)

**This package is deprecated.** Use the excellent [sqlc](https://github.com/kyleconroy/sqlc) package instead.

Package sqlr is designed to reduce the effort required to work with SQL databases.
It is intended for programmers who are comfortable with writing SQL, but would like 
assistance with the sometimes tedious process of preparing SQL queries for tables 
that have a large number of columns, or have a variable number of input parameters.

This package is designed to work seamlessly with the standard library
"database/sql" package. It does not provide any layer on top of *sql.DB
or *sql.Tx. If the calling program has a need to execute queries independently
of this package, it can use "database/sql" directly, or make use of any other
third party package that uses "database/sql".

This README provides an overview of how to use this package. For
more detailed documentation, see https://jjeffery.github.io/sqlr, or consult
the [GoDoc documentation](https://godoc.org/github.com/jjeffery/sqlr).

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Obtaining the package](#obtaining-the-package)
- [Prepare SQL queries based on row structures](#prepare-sql-queries-based-on-row-structures)
- [Autoincrement Column Values](#autoincrement-column-values)
- [Null Columns](#null-columns)
- [JSON Columns](#json-columns)
- [WHERE IN Clauses with Multiple Values](#where-in-clauses-with-multiple-values)
- [Type-Safe Query Functions](#type-safe-query-functions)
- [Performance and Caching](#performance-and-caching)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Obtaining the package

```bash
go get github.com/jjeffery/sqlr
```

Note that if you are interested in running the tests, you will need to
get additional database driver packages and setup a test database. See
the [detailed documentation](https://jjeffery.github.io/sqlr) for more
information.

## Prepare SQL queries based on row structures

Preparing SQL queries with many placeholder arguments is tedious and error-prone. The following
insert query has a dozen placeholders, and it is difficult to match up the columns with the
placeholders. It is not uncommon to have tables with many more columns than this example, and the
level of difficulty increases with the number of columns in the table.

```sql
insert into users(id,given_name,family_name,dob,ssn,street,locality,postcode,country,phone,mobile,fax)
values(?,?,?,?,?,?,?,?,?,?,?,?)
```

This package uses reflection to simplify the construction of SQL queries. Supplementary information
about each database column is stored in the structure tag of the associated field.

```go
type User struct {
    ID          int       `sql:"primary key"`
    GivenName   string
    FamilyName  string
    DOB         time.Time
    SSN         string
    Street      string
    Locality    string
    Postcode    string
    Country     string
    Phone       string
    Mobile      string
    Facsimile   string    `sql:"fax"` // "fax" overrides the column name
}
```

The calling program creates a schema, which describes rules for generating SQL statements. These
rules include specifying the SQL dialect (eg MySQL, Postgres, SQLite) and the naming convention
used to convert Go struct field names into column names (eg "GivenName" => "given_name"). The schema
is usually created during program initialization. Once created, a schema is immutable and can be
called concurrently from multiple goroutines.

```go
schema := NewSchema(
  WithDialect(MySQL),
  WithNamingConvention(SnakeCase),
)
```

A session is created using a context, a database connection (eg `*sql.DB`, `*sql.Tx`, `*sql.Conn`),
and a schema. A session is inexpensive to create, and is intended to last no longer than a single
request (which might be a HTTP request, in the case of a HTTP server). A session is bounded by the
lifetime of its context. The most common pattern is to create a new session for each database transaction.

```go
sess := NewSession(ctx, tx, schema)
```

With a session, it is possible to create simple CRUD statements with minimal effort.

```go
 var row User
 // ... populate row with data here and then ...

 // generates the correct SQL to insert a row into the users table
 result, err := sess.InsertRow(row)

 // ... and then later on ...

 // generates the correct SQL to update a the matching row in the users table
 result, err := sess.UpdateRow(row)
```

In the example above, the generated insert and update statements would look like:

```sql
 insert into users(`id`,`given_name`,`family_name`,`dob`,`ssn`,`street`,`locality`,`postcode`,
 `country`,`phone`,`mobile`,`fax`) values(?,?,?,?,?,?,?,?,?,?,?,?)

 update users set `given_name`=?,`family_name`=?,`dob`=?,`ssn`=?,`street`=?,`locality`=?,
 `postcode`=?,`country`=?,`phone`=?,`mobile`=?,`fax`=? where `id`=?
```

If the schema is created with a different dialect then the generated SQL will be different.
For example if the Postgres dialect was used the insert and update queries would look more like:

```sql
 insert into users("id","given_name","family_name","dob","ssn","street","locality","postcode",
 "country","phone","mobile","fax") values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)

 update users set "given_name"=$1,"family_name"=$2,"dob"=$3,"ssn"=$4,"street"=$5,"locality"=$6,
 "postcode"=$7,"country"=$8,"phone"=$9,"mobile"=$10,"fax"=$11 where "id"=$12
```

More complex update queries are handled by the [Session.Exec](https://godoc.org/github.com/jjeffery/sqlr#Session.Exec) method.

Select queries are handled by the [Session.Select](https://godoc.org/github.com/jjeffery/sqlr#Session.Select) method:

```go
 var rows []*User

 // will populate rows slice with the results of the query
 rowCount, err := sess.Select(&rows, "select {} from users where postcode = ?", postcode)

 var row User

 // will populate row with the first row returned by the query
 rowCount, err = sess.Select(&row, "select {} from users where {}", userID)

 // more complex query involving joins and aliases
 rowCount, err = sess.Select(&rows, `
     select {alias u}
     from users u
     inner join user_search_terms ust on ust.user_id = u.id
     where ust.search_term like ?
     order by {alias u}`, searchTermText)
```

The SQL queries prepared in the above example would look like the following:

```go
 select `id`,`given_name`,`family_name`,`dob`,`ssn`,`street`,`locality`,`postcode`,
 `country`,`phone`,`mobile`,`fax` from users where postcode=?

 select `id`,`given_name`,`family_name`,`dob`,`ssn`,`street`,`locality`,`postcode`,`country`,
 `phone`,`mobile`,`fax` from users where id=?

 select u.`id`,u.`given_name`,u.`family_name`,u.`dob`,u.`ssn`,u.`street`,u.`locality`,
 u.`postcode`,u.`country`,u.`phone`,u.`mobile`,u.`fax` from users u inner join
 user_search_terms ust on ust.user_id = u.id where ust.search_term_like ? order by u.`id`
```

The examples are using a MySQL dialect. If the schema had been setup for, say, a Postgres
dialect, a generated query would look more like:

```go
 select "id","given_name","family_name","dob","ssn","street","locality","postcode","country",
 "phone","mobile","fax" from users where postcode=$1
```

It is an important point to note that this feature is not about writing the SQL for the programmer.
Rather it is about "filling in the blanks": allowing the programmer to specify as much of the
SQL query as they want without having to write the tiresome bits.

For more information on preparing queries, see [the detailed documentation](https://jjeffery.github.io/sqlr).

## Autoincrement Column Values

When inserting rows using [InsertRow](https://godoc.org/github.com/jjeffery/sqlr#Session.InsertRow),
if a column is defined as an autoincrement column, then the generated value will be retrieved from
the database server, and the corresponding field in the row structure will be updated.

```go
 type Row {
   ID   int    `sql:"primary key autoincrement"`
   Name string
 }

 row := &Row{Name: "some name"}
 _, err := sess.InsertRow(row)
 if err != nil {
   log.Fatal(err)
 }

 // row.ID will contain the auto-generated value
 fmt.Println(row.ID)
```

## Null Columns

Most SQL database tables have columns that are nullable, and it can be tiresome to 
always map to pointer types or special nullable types such as `sql.NullString`. In 
many cases it is acceptable to map the zero value for the field a database NULL 
in the corresponding database column.

Where it is acceptable to map a zero value to a NULL database column, the Go struct
field can be marked with the "null" keyword in the field's struct tag.

```go
 type Employee struct {
     ID        int     `sql:"primary key"`
     Name      string
     ManagerID int     `sql:"null"`
     Phone     string  `sql:"null"`
 }
```

In the above example the `manager_id` column can be null, but if all valid IDs are 
non-zero, it is unambiguous to map the zero value to a database NULL. Similarly, if 
the `phone` column an empty string it will be stored as a NULL in the database.

Care should be taken, because there are cases where an empty value and a database NULL do not
represent the same thing. There are many cases, however, where this feature can be applied,
and the result is simpler code that is easier to read.

## JSON Columns

It is not uncommon to serialize complex objects as JSON text for storage in an SQL database.
Native support for JSON is available in some database servers: in partcular Postgres has
excellent support for JSON.

It is straightforward to use this package to serialize a structure field to JSON:

```go
 type SomethingComplex struct {
     Name       string
     Values     []int
     MoreValues map[string]float64
     // ... and more fields here ...
 }

 type Row struct {
     ID    int                `sql:"primary key"`
     Name  string
     Cmplx *SomethingComplex  `sql:"json"`
 }
```

In the example above the `Cmplx` field will be marshaled as JSON text when
writing to the database, and unmarshaled into the struct when reading from
the database.

## WHERE IN Clauses with Multiple Values

While most SQL queries accept a fixed number of parameters, if the SQL query
contains a `WHERE IN` clause, it requires additional string manipulation to match
the number of placeholders in the query with args.

This package simplifies queries with a variable number of arguments. When processing
an SQL query, it detects if any of the arguments are slices:

```go
 // GetWidgets returns all the widgets associated with the supplied IDs.
 func GetWidgets(sess *sqlr.Session, ids ...int) ([]*Widget, error) {
     var rows []*Widget
     _, err := sess.Select(&rows, `select {} from widgets where id in (?)`, ids)
     if err != nil {
       return nil, err
     }
     return widgets, nil
 }
```

In the above example, the number of placeholders ("?") in the query will be increased to
match the number of values in the `ids` slice. The expansion logic can handle any mix of
slice and scalar arguments.

## Type-Safe Query Functions

A session can create type-safe query functions. This is a very powerful feature and makes
it very easy to create type-safe data access objects.

```go
var getWidget func(id int64) (*Widget, error)

// a session can make a typesafe function to retrieve an individual widget
sess.MakeQuery(&getWidget)

// now use the created function
widget, err := getWidget(42)
if err != nil {
    return err
}

// ... now use the widget ...
```

See [Session.MakeQuery](https://godoc.org/github.com/jjeffery/sqlr/#Session.MakeQuery)
in the [GoDoc](https://godoc.org/github.com/jjeffery/sqlr) for examples.
