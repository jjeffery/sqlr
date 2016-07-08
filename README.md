# sqlf: Formatting SQL statements

Package `sqlf` provides assistance in creating SQL statements. 

[![GoDoc](https://godoc.org/github.com/jjeffery/sqlf?status.svg)](https://godoc.org/github.com/jjeffery/sqlf)
[![License](http://img.shields.io/badge/license-MIT-green.svg?style=flat)](https://raw.githubusercontent.com/jjeffery/sqlf/master/LICENSE.md)
[![GoReportCard](https://goreportcard.com/badge/github.com/jjeffery/sqlf)](https://goreportcard.com/report/github.com/jjeffery/sqlf)

**NOTE:** This package is a work in progress. There is 
no backwards compatibility guarantee at this time.

## Rationale

There are a number of good ORM packages for the Go language, with varying
sets of features. There are times, however, when an ORM may not be 
appropriate, and an SQL-based API might provide the desired simplicity,
control, and/or performance.

Using an SQL-based API such as `database/sql`, however, can be a little tedious
and error prone. There are some good packages available that make working
with `database/sql` easier: a good example  is `github.com/jmoiron/sqlx`.

While packages such as `sqlx` go a long way towards handling the results
of SQL queries, it can still be quite tedious to construct the SQL for a
query in the first place. This is particularly so for queries against
database tables that have many columns and hence many placeholders (ie "?")
in the SQL -- it can be error-prone constructing and maintaining the API 
calls to have the correct number of arguments in the correct order.

Package `sqlf` attempts to make the generation of SQL statements easier
by using a Printf-style API that will build SQL statements based on the
contents of Go language structures.

## Example

Coming soon.





