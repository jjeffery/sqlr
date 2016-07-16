/*
Package sqlstmt aims to make it easy to construct and execute SQL
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

More information about this package can be found at https://github.com/jjeffery/sqlstmt.
*/
package sqlstmt
