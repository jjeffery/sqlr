.. _naming_conventions:

Column Naming Conventions
=========================

The `sqlr` package creates SQL column lists from the information present
in Go struct fields. In order to do this, it needs to know the naming
convention needed to convert a Go struct field name into its corresponding
database column name.

The following naming conventions are supported out of the box:

* `Snake case <https://en.wikipedia.org/wiki/Snake_case>`_ (eg 
  ``HomePhone`` converts to ``home_phone``)
* Same case (same as the Go struct field, eg ``HomePhone`` => ``HomePhone``)
* Lower case (convert to lower case, eg ``HomePhone`` => ``homephone``)

Preparing a new naming convention is possible, by implementing 
the `sqlr.NamingConvention <https://godoc.org/github.com/jjeffery/sqlr#NamingConvention>`_
interface.

The naming convention can be specified when creating the schema::

	schema := sqlr.NewSchema(
		sqlr.WithDialect(sqlr.Postgres),
		sqlr.WithNamingConvention(sqlr.SnakeCase),
	)

If the naming convention is not specified, it defaults to snake case.
