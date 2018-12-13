.. highlight: go
.. _schema_type:

The Schema Type
===============

The `Schema` type provides the logic necessary to prepare SQL
statements that will be accepted by the database server. To
function, the Schema requires the following information:

* Information about the SQL dialect supported by the target
  database server;
* The naming convention used for mapping Go struct field
  names into database column names; and
* Any special exceptions for database column names that
  do not fit the naming convention, and cannot be specified
  in the Go struct tag.

Creating the Schema
-------------------

A schema is created using the `NewSchema` function, which accepts
a variable number of `SchemaOption` options::

    schema := sqlr.NewSchema(
        sqlr.WithDialect(sqlr.Postgres),
        sqlr.WithNamingConvention(sqlr.SnakeCase),
    )

In most cases it is enough to create a single schema instance
at program initialization, probably at the same point as the
``*sql.DB`` instance is being created. In fact it is possible to
specify the dialect directly from the ``*sql.DB`` handle::

    db, err := sql.Open("postgres", "postgres://user:pwd@localhost/mydb")
    if err != nil {
        log.Fatal(err)
    }

    schema := sqlr.NewSchema(
        ForDB(db), // will choose the correct dialect for db
    )

It is recommended practice to create a schema at program initialization and 
re-use it rather 
than create one whenever necessary. The reason for this is that the 
schema caches information about Go structures and SQL statements in order
to improve performance.

Mapping Individual Field/Column Names
-------------------------------------

Sometimes it is necessary to provide a special naming convention for
creating a column name from a Go struct field. Although it is probably
more common to use Go struct tags to specify a column name for a field,
there are times when this is not possible. When this happens, it is 
possible to specify individual naming rules using the ``WithField`` schema 
option. Take the following (simplified) table::

    create table contact_details(
        id              integer primary key not null,
        work_email      text,
        work_fax        text,
        home_email      text,
        home_fax_number text
    )

There is some inconsistency in the column names, but we want to use a common
structure to handle the location-specific contact details::

    type LocationContact struct {
        Email     string
        Facsimile string // cannot specify a struct tag here for two different column names
    }

    type ContactDetail struct {
        ID   int              `sql:"primary key"`
        Home LocationContact
        Work LocationContact
    }

When it comes to wanting to map the names for the fax numbers, it turns out
not to be possible to use the Go struct tag: it is not possible to specify
the names of two different columns on the same Go struct field.

The solution is to specify the individual column names when creating the
schema::

    schema := NewSchema(
        WithDialect(sqlr.SQLite),
        WithNamingConvention(sqlr.Snake),
        WithField("Work.Facsimile", "work_fax"),
        WithField("Home.Facsimile", "home_fax_number"),
    )

Specifying a struct tag key for column names
--------------------------------------------

One of the schema options allows for a struct tag key to be associated with 
the schema::

    schema := NewSchema(
        WithDialect(sqlr.Postgres),
        WithNamingConvention(sqlr.SnakeCase),
        WithKey("pg"),
    )

Where does this come in useful? Perhaps it is best to describe the scenario
that occured that brought about this feature being added to the package.

Take the example of a existing system that makes use of (say) an MS SQL Server
database. Like many SQL Server databases, it uses a "Same" naming convention, 
where the name of the column is the same as its equivalent Go struct field.

Then consider that this system is being migrated over to operate with (say) a 
PostgreSQL database. The "Same" naming convention does not work well with
PostgreSQL: it is far more idiomatic to use the "snake_case" naming convention.

So the decision is made to change the database naming convention as part of
the migration project. To make the scheduling of the cutover more flexible it
would be good if the Go program could work with both the MS SQL Server database
as well as with the PostgreSQL database. For queries made to the database using
the ``sqlr`` package this should not be a significant problem; all that is needed
are two different schemas to handle the different dialects and the different
naming conventions::

	mssqlSchema := sqlr.NewSchema(
		WithDialect(sqlr.MSSQL),
		WithNamingConvention(sqlr.SameCase),
	)

	pgSchema := sqlr.NewSchema(
		WithDialect(sqlr.Postgres),
		WithNamingConvention(sqlr.SnakeCase),
	)

The problem occurs when it is necessary to specify a column naming exception.
Just say that there is a Go struct with a field called `Max`, and that we
have our reasons for wanting it to be called that, but that the MS SQL Server
column name is called `MaximumValue` to avoid conflict with the SQL reserved word. ::

    type MyRow struct {
        ID  int   `sql:"primary key"`
        Max int   `sql:"MaximumValue"`
        // ... other fields go here ...
    }

We would like to be able to specify a different column name for the Postgres
database (which also treats ``max`` as a reserved word), but there is no way to
specify two different column names in the Go struct tag.

While it is possible to get around this problem using the ``WithField`` schema
option, there is some benefit visibility-wise if the two column names can
appear in the Go struct field. This is where the ``WithKey`` schema option
becomes relevant. If the two schemas each specify a different struct tag key, then
the ``sqlr`` package will look in the struct tag key for column names::

    mssqlSchema := sqlr.NewSchema(
        WithDialect(sqlr.MSSQL),
        WithNamingConvention(sqlr.SameCase),
        WithKey("mssql"),
    )

    pgSchema := sqlr.NewSchema(
        WithDialect(sqlr.Postgres),
        WithNamingConvention(sqlr.SnakeCase),
        WithKey("pg"),
    )

So now the column name exceptions can be included in the struct tag, and will only
apply to the schema with the matching tag key::

    type MyRow struct {
        ID  int   `sql:"primary key"`
        Max int   `mssql:"MaximumValue" pg:"maximum_value"`
        // ... other fields go here ...
    }

This is another example of a schema configuration option that is not likely to
be used very often, but it can occasionally come in useful.

Replacing SQL identifiers
-------------------------

The ``WithIdentifer`` option is another schema option that helps with portability SQL
queries across different SQL database schemas that have different naming conventions.
Once again, it is a feature that is probably not commonly used, but can come in handy
with the sort of scenario described in the previous section (eg migrating from one
database server to another, with a change of naming convention along the way).

In the scenario described, we are attempting to make our code portable across two
different database schemas, where the structure of the data is the same but the
columns have different naming conventions. The column names are mapped from the
corresponding Go structures, but there is no reason why they cannot appear in the
SQL text as well::

    rowsAffected, err := schema.Exec(db, row, `
        update widgets 
        set {} where {} 
        and version = ?`, row.Version,
    )

In the example above the schema will handle the different naming conventions for
the column names specified by the special ``{}`` markers, but the table name ``widgets``
and the column name ``version`` has been hard-coded into the SQL.

The solution, while not ideal, is to specify individual identifer replacements 
when creating the schema::

    mssqlSchema := sqlr.NewSchema(
        WithDialect(sqlr.MSSQL),
        WithNamingConvention(sqlr.SameCase),
        WithIdentifer("widgets", "Widgets"),
        WithIdentifer("version", "Version"),
    )

So the SQL that would be produced for the MSSQL schema will now substitute all
identifers named ``widgets`` with ``Widgets``, and all identifiers named 
``version`` with ``Version``. 

If you think this is untidy, you are probably correct. It is a seldom-used feature
designed to help with an uncommon situation.
