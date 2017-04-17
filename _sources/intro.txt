Introduction
============

Package `sqlr` is designed to reduce the effort required to implement
common operations performed with SQL databases. It is intended for programmers
who are comfortable with writing SQL, but would like assistance with the
sometimes tedious process of preparing SQL queries for tables that have a
large number of columns, or have a variable number of input parameters.

This package is designed to work seamlessly with the standard library
"database/sql" package. It does not provide any layer on top of `*sql.DB`
or `*sql.Tx`. If the calling program has a need to execute queries independently
of this package, it can use "database/sql" directly, or make use of any other
third party package that uses "database/sql".

The following features are provided to simplify writing SQL database queries:

- Prepare SQL from row structures
- Autoincrement column values
- Null columns
- JSON columns
- WHERE IN Clauses with multiple values
- Code generation


Installing
----------

To obtain a copy of the `sqlr` package, use ``go get``:

.. code-block:: sh

    go get github.com/jjeffery/sqlr

Note that additional setup is required if you wish to run the tests
against database servers. The setup required is discussed in :ref:`tests`.