# sqlr: SQL statements

Package sqlr is designed to reduce the effort required to implement
common operations performed with SQL databases.

[![GoDoc](https://godoc.org/github.com/jjeffery/sqlr?status.svg)](https://godoc.org/github.com/jjeffery/sqlr)
[![License](http://img.shields.io/badge/license-MIT-green.svg?style=flat)](https://raw.githubusercontent.com/jjeffery/sqlr/master/LICENSE.md)
[![Build Status (Linux)](https://travis-ci.org/jjeffery/sqlr.svg?branch=master)](https://travis-ci.org/jjeffery/sqlr)
[![Coverage Status](https://coveralls.io/repos/github/jjeffery/sqlr/badge.svg?branch=master)](https://coveralls.io/github/jjeffery/sqlr?branch=master)
[![GoReportCard](https://goreportcard.com/badge/github.com/jjeffery/sqlr)](https://goreportcard.com/report/github.com/jjeffery/sqlr)

**NOTE:** This package is still a work in progress. There is no backwards compatibility guarantee
at this time.

`import "github.com/jjeffery/sqlr"`

Package sqlr is designed to reduce the effort required to implement
common operations performed with SQL databases. It is intended for programmers
who are comfortable with writing SQL, but would like assistance with the
sometimes tedious process of preparing SQL queries for tables that have a
large number of columns, or have a variable number of input parameters.

This package is designed to work seamlessly with the standard library
"database/sql" package. It does not provide any layer on top of *sql.DB
or *sql.Tx. If the calling program has a need to execute queries independently
of this package, it can use "database/sql" directly, or make use of any other
third party package that uses "database/sql".

For more information, see the [GoDoc documentation](https://godoc.org/github.com/jjeffery/sqlr).