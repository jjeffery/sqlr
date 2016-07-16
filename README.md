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
database tables that have many columns and hence many placeholders (ie "?")
in the SQL -- it can be error-prone constructing and maintaining the API 
calls to have the correct number of arguments in the correct order.

Package `sqlstmt` attempts to make the generation of SQL statements easier
by using an API that builds SQL statements based on the contents of Go language 
structures.

## Example

Coming soon.

