Prior Art
=========

There are many packages for Go that enhance the functionality found in 
the standard library `database/sql` package. While not an exhaustive list,
here are a few packages that have some similarity to package `sqlr`.

These packages pre-date package `sqlr` and are popular. If you are considering
using this package, you might like to investigate these other packages, to see
if they suit your needs better.

If you are aware of a package that you think should be included in this list,
please create an issue or submit a pull request.

This list does not include any of the numerous ORMs available for Go. The reason
is that package `sqlr` does not qualify as an ORM. It is a package that
provides assistance with using the Go standard library, but not a fully featured
ORM. For  a comprehensive list of Go ORMs refer to the following:

* http://libs.club/golang/data-storage/orms
* https://awesome-go.com/#orm
* https://golanglibs.com/search?q=ORM

github.com/jmoiron/sqlx
-----------------------

Package `sqlx <https://github.com/jmoiron/sqlx>`_ is a popular package 
that provides extensions to the Go standard library `database/sql` package.
It provides a number of useful features, including the ability to marshal
rows into structs, maps and slices.

github.com/kisielk/sqlstruct
----------------------------

Package `sqlstruct <https://github.com/kisielk/sqlstruct>`_ provides 
convenience functions for using structs with the Go library standard 
`database/sql` package. It is similar to package `sqlr` in the sense 
that it formats SQL statements with column names derived from Go 
struct fields.
