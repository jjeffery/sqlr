.. highlight: go
.. _session_type:

Sessions
========

The `Session` type describes a short-lived object that combines:

* a context_;
* a database querier (ie ``*sql..DB``, ``*sql.Tx``, ``*sql.Conn``); and
* a database :ref:`schema <schema_type>`

.. _context: https://golang.org/pkg/context#Context

The context defines the lifetime of the session, the database querier is
used to send queries to the database, and the schema provides information
about the database schema to assist with preparing queries.

Sessions are inexpensive to create and should be short-lived. The typical
use-case is to create one session for each database transaction::

    var (
	    ctx    context.Context
	    tx     *sql.Tx // or could be a *sql.DB or *sql.Conn
	    schema *sqlr.Schema
    )

    // ... initialize ctx, tx and schema and then ...

    session := sqlr.NewSession(ctx, tx, schema)


Query rows using Select
-----------------------

Execute commands using Exec
---------------------------



TODO

* sessions do all the work
* querying directly using Select, Exec
* row operations using session.Row(row).Exec
* create query functions using MakeQuery
* intro to using DAOs
