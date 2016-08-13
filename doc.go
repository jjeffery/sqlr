/*
Package sqlrow to make it easy to construct and execute SQL
queries for common, row-based scenarios. Supported scenarios include:

 (a) Insert, update or delete a single row based on the contents of a Go struct;
 (b) Select zero, one or more rows int a a slice of Go structs; and
 (c) Select a single row into a Go struct.

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
 sqlrow.Update(db, row, `update users set {} where {} and version = ?', row.version)

SQL DELETE

DELETE queries are similar to UPDATE queries:

 sqlrow.Delete(db, row, `users`)

is translated as (for MySQL, SQLite):
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
package sqlrow
