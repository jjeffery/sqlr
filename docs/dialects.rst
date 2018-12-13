.. _dialects:

SQL Dialects
============

The `sqlr` package is designed to be as SQL-agnostic as possible, but 
when it is generating SQL it does need to know the following:

* How to quote column names to ensure they are not interpreted as an SQL keyword

  * PostgreSQL uses double quotes: ``"column_name"``
  * MySQL uses back ticks: ```column_name```
  * MS SQL Server uses square braces: ``[column_name]``

* How to write placeholders for arguments

  * PostgreSQL uses numbered placeholders: ``$1``, ``$2``, etc
  * Almost everyone else uses question marks: ``?``

The following dialects are available out of the box:

* Postgres *(aka PostgreSQL)*
* MySQL
* SQLite
* MS SQL Server
* ANSI SQL

Preparing a new dialect is possible, by implementing the 
`sqlr.Dialect <https://godoc.org/github.com/jjeffery/sqlr#Dialect>`_
interface.

The default dialect
-------------------

Most programs use only one SQL driver, and in these circumstances `sqlr`
will do the right thing.

For example, if a program is using Postgres, it will need to load the appropriate driver,
probably in the `main` package:

.. code-block:: go

  import _ "github.com/lib/pq"

By default `sqlr` will check the list of loaded SQL drivers and pick the
first one to decide on the SQL dialect to use. If only one SQL driver has been
loaded, it will choose correctly. In this example it will automatically choose 
the "postgres" dialect.

Specifying the SQL dialect
--------------------------

If your program references multiple SQL drivers, it is necesary to 
specify which dialect is in use. This can be done when opening the 
database connection::

  // open the database
  db, err := sql.Open("postgres", "user=test dbname=test sslmode=disable")
  if err != nil {
    log.Fatal(err)
  }

  // create the schema
  schema := sqlr.NewSchema(
    sqlr.WithDialect(sqlr.Postgres),
  )

Using multiple dialects
-----------------------

If your program makes use of multiple database connections with different
types of server, the best thing to do is to specify a ``sqlr.Schema`` 
for each of the databases::

  var (
    pgSchema = sqlr.NewSchema(
      sqlr.WithDialect(sqlr.Postgres),
    )

    mysqlSchema = sqlr.NewSchema(
      sqlr.WithDialect(sqlr.MySQL),
    )
  )

When the time comes to create sessions, use the appropriate schema::

  // create a session for the postgres database
  pgSession := sqlr.NewSession(ctx, pgDB, pgSchema)

  // create a session for the mysql database
  mysqlSession := sqlr.NewSession(ctx, mysqlDB, mysqlSchema)

