/*
OLD DESCRIPTION

Package sqlrow makes it easy to construct and execute SQL
queries for common, row-based scenarios. Supported scenarios include:

 (a) Insert, update or delete a single row based on the contents of a Go struct;
 (b) Select a single row into a Go struct; and
 (c) Select zero, one or more rows into a slice of Go structs.

This package is intended for programmers who are comfortable with
writing SQL, but would like assistance with the sometimes tedious
process of preparing SELECT, INSERT, UPDATE and DELETE statements
for tables that have a large number of columns.

This package is designed to work seamlessly with the standard library
"database/sql" package. It does not provide any layer on top of *sql.DB
or *sql.Tx. If the calling program has a need to execute queries independently
of this package, it can use "database/sql" directly, or make use of any other
third party package that uses "database/sql".

SQL INSERT

Package sqlrow uses reflection on the supplied row to provide assistance
with creating SQL. This assistance is particularly useful for tables with
many columns, but the following examples use this simple structure:

 type UserRow struct {
	ID         int64 `sql:"primary key autoincrement"`
	GivenName  string
	FamilyName string
 }

For a row of type UserRow the following INSERT query:

 sqlrow.Insert(db, row, `insert into users({}) values({})`)

will be translated into the following, depending on the SQL dialect:
 insert into users(`given_name`,`family_name`) values(?,?)   -- MySQL, SQLite
 insert into users("given_name","family_name") values($1,$2) -- PostgreSQL
 insert into users([given_name],[family_name]) values(?,?)   -- MSSQL

In the above example note that the "id" column is not inserted. This is
because it is defined as an auto-increment column. If it were not an
auto-increment column it would be included in the column list.

This pattern is so common for inserting individual rows that,
for convenience, providing just the table name has the same result:

 sqlrow.Insert(db, row, `users`)

SQL UPDATE

The following UPDATE query:

 sqlrow.Update(db, row, `update users set {} where {}`)

will be translated into the following:
 update users set `given_name`=?,`family_name`=? where `id`=?    -- MySQL, SQLite
 update users set "given_name"=$1,"family_name"=$2 where "id"=$3 -- PostgreSQL
 update users set [given_name]=?,[family_name]=? where [id]=?    -- MSSQL

This pattern is so common for inserting individual rows that,
for convenience, providing just the table name has the same result:

 sqlrow.Update(db, row, `users`)

It is possible to construct more complex UPDATE statements. The following
example can be useful for rows that make use of optimistic locking:
 sqlrow.Update(db, row, `update users set {} where {} and version = ?', oldVersion)

SQL DELETE

DELETE queries are similar to UPDATE queries:

 sqlrow.Delete(db, row, `delete from users where {}`)
and
 sqlrow.Delete(db, row, `users`)

are both translated as (for MySQL, SQLite):
 delete from users where `id`=?

SQL SELECT

SQL SELECT queries can be constructed easily

 var rows []UserRow
 sql.Select(db, &rows, `select {} from users where given_name=?`, "Smith")

is translated as (for MySQL, SQLite):
 select `id`,`given_name`,`family_name` from users where given_name=?

More complex queries involving joins and table aliases are possible:

 sql.Select(db, &rows, `
   select {alias u}
   from users u
   inner join user_search_terms t on t.user_id = u.id
   where t.search_term like ?`, "Jon%")

is translated as (for MySQL, SQLite):
 select u.`id`,u.`given_name`,u.`family_name`
 from users u inner join user_search_terms t
 on t.user_id = u.id
 where t.search_term like ?

Performance and Caching

Package sqlrow makes use of reflection in order to build the SQL that is sent
to the database server, and this imposes a performance penalty. In order
to reduce this overhead the package caches queries generated. The end result
is that the performance of this package is close to the performance of
code that uses hand-constructed SQL queries to call package "database/sql"
directly.

Source Code

More information about this package can be found at https://github.com/jjeffery/sqlrow.
*/

/*
Package sqlrow is designed to reduce the effort required to implement
common operations against SQL databases. It is intended for programmers
who are comfortable with writing SQL, but would like assistance with the
sometimes tedious process of preparing SQL queries for tables that have a
large number of columns, or have a variable number of input parameters.

Prepare SQL from Row structures

Preparing SQL queries with many placeholder arguments is tedious and error-prone. The following
insert query has a dozen placeholders, and it is difficult enough to match the columns with the
placeholder. Many tables have multiple dozens of columns, some have more.
 insert into users(id,given_name,family_name,dob,ssn,street,locality,postcode,country,phone,mobile,fax)
 values(?,?,?,?,?,?,?,?,?,?,?,?)
This package uses reflection to simplify the construction of SQL statements for insert, update, delete
and select queries. Supplementary information about each database column is stored as a structure tag
in the associated field.
 type User struct {
     ID          int       `sql:"primary key"`
     GivenName   string
     FamilyName  string
     DOB         time.Time `sql:"null"`
     SSN         string
     Street      string
     Locality    string
     Postcode    string
     Country     string
     Phone       string    `sql:"null"`
     Mobile      string    `sql:"null"`
     Facsimile   string    `sql:"fax null"`
 }
The calling program creates a schema, which describes rules for generating SQL statements. These
rules include specifying the SQL dialect (eg MySQL, Posgres, SQLite) and the naming convention
used to convert Go struct field names into column names (eg "GivenName" => "given_name"). The schema
is usually created during program initialization.
 schema := NewSchema(
   WithDialect(MySQL),
   WithNamingConvention(SnakeCase),
 )
Once the schema has been defined and a database handle is available (eg *sql.DB, *sql.Tx), it is possible
to create simple row insert/update/delete statements with minimal effort.
 var row User
 // ... populate row with data here and then ...

 // generates the correct SQL to insert a row into the users table
 rowsAffected, err := schema.Exec(db, row, "insert into users({}) values({})")

 // ... and then later on ...

 // generates the correct SQL to update a the matching row in the users table
 rowsAffected, err := schema.Exec(db, row, "update users set {} where {}")
The Exec method parses the SQL query and replaces occurrances of "{}" with the column names
or placeholders that make sense for the SQL clause in which they occur. In the example above,
the insert and update statements would look like:
 insert into users(`id`,`given_name`,`family_name`,`dob`,`ssn`,`street`,`locality`,`postcode`,
 `country`,`phone`,`mobile`,`fax`) values(?,?,?,?,?,?,?,?,?,?,?,?)

 update users set `given_name`=?,`family_name`=?,`dob`=?,`ssn`=?,`street`=?,`locality`=?,
 `postcode`=?,`country`=?,`phone`=?,`mobile`=?,`fax`=? where `id`=?
If the schema is created with a different dialect then the generated SQL will be different.
For example if the Postgres dialect was used the insert and update queries would look more like:
 insert into users("id","given_name","family_name","dob","ssn","street","locality","postcode",
 "country","phone","mobile","fax"") values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)

 update users set "given_name"=$1,"family_name"=$2,"dob"=$3,"ssn"=$4,"street"=$5,"locality"=$6,
 "postcode"=$7,"country"=$8,"phone"=$9,"mobile"=$10,"fax"=$11 where "id"=$12
Select queries are handled in a similar fashion:
 var rows []*User

 // will populate rows slice with the results of the query
 rowCount, err := schema.Select(db, &rows, "select {} from users where postcode = ?", postcode)

 var row User

 // will populate row with the first row returned by the query
 rowCount, err = schema.Select(db, &row, "select {} from users where {}", userID)

 // more complex query involving joins and aliases
 rowCount, err = schema.Select(db, &rows, `
     select {alias u}
     from users u
     inner join user_search_terms ust on ust.user_id = u.id
     where ust.search_term like ?
     order by {alias u}`, searchTermText + "%")
The SQL queries prepared in the above example would look like the following:
 select `id`,`given_name`,`family_name`,`dob`,`ssn`,`street`,`locality`,`postcode`,
 `country`,`phone`,`mobile`,`fax` from users where postcode=?

 select `id`,`given_name`,`family_name`,`dob`,`ssn`,`street`,`locality`,`postcode`,`country`,
 `phone`,`mobile`,`fax` from users where id=?

 select u.`id`,u.`given_name`,u.`family_name`,u.`dob`,u.`ssn`,u.`street`,u.`locality`,
 u.`postcode`,u.`country`,u.`phone`,u.`mobile`,u.`fax` from users u inner join
 user_search_terms ust on ust.user_id = u.id where ust.search_term_like ? order by u.`id`
The examples are using a MySQL dialect. If the schema had been setup for, say, a Postgres
dialect, a generated query would look more like
 select "id","given_name","family_name","dob","ssn","street","locality","postcode","country",
 "phone","mobile","fax" from users where postcode=$1

Autoincrement Primary Keys

When inserting rows, if a column is defined as an autoincrement column, then the generated
value will be retrieved and the corresponding field in the row structure will be updated.
 type Row {
   ID   int    `sql:"primary key autoincrement"`
   Name string
 }

 row := &Row{ Name: "some name"}
 _, err := schema.Exec(db, row, "insert into table_name({}) values({})")
 if err != nil {
   log.Fatal(err)
 }

 // row.ID will contain the auto-generated value
 fmt.Println(row.ID)
*/
package sqlrow
