Performing Queries
==================

The `sqlr` package provides assistance in the more tedious aspects
of writing SQL queries, particularly queries that involve rows with
a large number of columns.

Having said that, in the interest of keeping the examples concise, 
the following examples do not have very complex table structures, 
or very many columns. Keep in mind, however, that the `sqlr` package 
becomes quite useful when the tables have a large number of columns.

Consider the following simple table:

.. code-block:: mysql

	create table users(
		id            integer primary key autoincrement not null,
		given_name    text,
		family_name   text,
		email_address text
	);

A corresponding Go struct for representing a row in the `users` table is::

	type User struct {
		ID           int `sql:"primary key autoincrement"`
		GivenName    string
		FamilyName   string
		EmailAddress string
	}

Note the use of struct tags to include information about the primary key
and auto-increment behaviour.

The following examples assume that a database has been opened, the database
table has been created, and the `*sql.DB` is stored in variable ``db``::

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(`
		create table users(
			id            integer primary key autoincrement not null,
			given_name    text,
			family_name   text,
			email_address text
		);`); err != nil {
		log.Fatal(err)
	}		

Creating the Schema
-------------------

The `Schema` type keeps track of the information required to map a Go struct field name 
into a corresponding column name. To prepare SQL statements, first create a `Schema` object::

	schema := sqlr.NewSchema(
		sqlr.WithDialect(sqlr.SQLite),
		sqlr.WithNamingConvention(sqlr.SnakeCase),
	)

The example above creates a schema that will generate SQL using a dialect compatible
with SQLite, where columns follow a 
`snake_case <https://en.wikipedia.org/wiki/Snake_case>`_ naming convention.

There is more detailed information on :ref:`the Schema type <schema_type>`, 
:ref:`dialects <dialects>`, and :ref:`naming conventions <naming_conventions>`, 
but for now we will move onto the more immediate topic of performing queries.

Inserting a row
---------------

.. code-block:: go

	// create the row object and populate with data
	userRow := &User{
		GivenName:    "Jane",
		FamilyName:   "Citizen",
		EmailAddress: "jane@citizen.com",
	}

	// insert the row into the `users` table using the db connection opened earlier,
	// and the schema created in the previous code sample
	err := schema.Exec(db, userRow, "insert into users({}) values({})")

	if err != nil {
		log.Fatal(err)
	}

	// userRow.ID contains the autoincrement value assigned by the DB server
	fmt.Println("User ID:", userRow.ID)

	// Output: User ID: 1

Note the non-standard ``{}`` in the SQL query above. The `sqlr` package
knows to substitute in column names in the appropriate quoted format that
is acceptable for the SQL dialect. In the example above, the SQL generated 
will look like the following:

.. code-block:: mysql

	insert into users(`given_name`,`family_name`,`email_address`)
	values(?,?,?)


The format of this "extended" SQL syntax is 
:ref:`covered in more detail later <sql_format>`, but for now take it as a given that
the schema knows how to expand the ``{}`` symbol into a column list that is 
appropriate for the SQL clause in which it appears.

Because this is an insert statement, and the ``id`` column is an auto-increment
column, the value of ``userRow.ID`` will contain the auto-generated value after 
the insert row statement has been executed.

(Note that the Postgres driver `github.com/lib/pq` does not support
the `Result.LastInsertId` method, and so this feature does not work for that
driver. See the `pq` package `GoDoc <https://godoc.org/github.com/lib/pq>`_ for
a work-around).

Updating a row
--------------

Continuing from the previous example::

	// change user details
	userRow.EmailAddress = "jane.citizen.314159@gmail.com"

	// update the row in the `users` table
	n, err = schema.Exec(db, userRow, "update users set {} where {}")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Number of rows updated:", n)

	// Output: Number of rows updated: 1


Once again the "extended" SQL syntax has been expanded into something more like:

.. code-block:: mysql

	update users set `given_name`=?,`family_name`=?,`email_address`=? where id=`?`

The value of the fields in the ``userRow`` instance have been supplied as arguments
for the placeholders in the update query.

Deleting a row
--------------

Continuing from the previous example::

	// delete the row in the `users` table
	n, err = schema.Exec(db, userRow, "delete from users where {}")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Number of rows deleted:", n)

	// Output: Number of rows deleted: 1

Selecting a single row
----------------------

Pretending that we have not deleted the row in the previous example::

	var userRow User 

	n, err := schema.Select(db, &userRow, "select {} from users where {}", 1)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Rows returned:", n)
	fmt.Println("User email:", u.EmailAddress)

	// Output:
	// Rows returned: 1
	// User email: jane.citizen.314159@gmail.com

The SQL generated in the above example would look like:

.. code-block:: mysql

	select `id`,`given_name`,`family_name`,`email_address` from users where `id`=?

The ``{}`` in the where clause always expands to the column or columns for the 
primary key. It possible (and common) to specify other criteria in the where
clause, something like:

.. code-block:: postgres

	select {} from users where email_address = ?

Selecting multiple rows
-----------------------

Performing a query that returns multiple rows is similar to returning a single
row. The only difference is that instead of passing a pointer to a struct, pass
a pointer to a slice of structs, or a pointer to a slice of struct pointers::

	// declare a slice of users for receiving the result of the query
	var users []*User

	// perform the query, specifying an argument for each of the
	// placeholders in the SQL query
	_,  err = schema.Select(db, &users, `
		select {}
		from users
		where family_name = ?`, "Citizen")
	if err != nil {
		log.Fatal(err)
	}

	// at this point, the users slice will contain one object for each
	// row returned by the SQL query
	for _, u := range users {
		doSomethingWith(u)
	}

Note, once again, the non-standard ``{}`` in the SQL query above. The `sqlr` 
package knows to substitute in column names in the appropriate format. In the 
example above, the SQL generated will look like the following:

.. code-block:: mysql

	select `id`,`family_name`,`given_name`,`email_address`
	from users
	where family_name = ?

For queries that involve multiple tables, it is always a good idea to
use table aliases::

	// declare a slice of users for receiving the result of the query
	var users []*User

	// perform the query, specifying an argument for each of the
	// placeholders in the SQL query
	_, err = schema.Select(db, &users, `
		select {alias u}
		from users u
		inner join user_search_terms t on t.user_id = u.id
		where u.term like ?`, "cit%")
	if err != nil {
		log.Fatal(err)
	}

	for _, u := range users {
		doSomethingWith(u)
	}

The SQL generated in this example looks like the following:

.. code-block:: mysql

	select u.`id`,u.`family_name`,u.`given_name`,u.`email_address`
	from users u
	inner join user_search_terms t on t.user_id = u.id
	where u.term like ?

WHERE IN Clauses
----------------

While most SQL queries accept a fixed number of parameters, if the SQL query
contains a `WHERE IN` clause, it requires additional string manipulation to match
the number of placeholders in the query with args.

This package simplifies queries with a variable number of arguments. When processing
an SQL query, it detects if any of the arguments are slices::

	// GetWidgets returns all the widgets associated with the supplied IDs.
	func GetWidgets(db *sql.DB, ids ...int) ([]*Widget, error) {
		var rows []*Widget
		_, err := schema.Select(db, &rows, `select {} from widgets where id in (?)`, ids)
		if err != nil {
			return nil, err
		}
		return widgets, nil
	}

In the above example, the number of placeholders (``?``) in the query will be increased to
match the number of values in the ``ids`` slice. The expansion logic can handle any mix of
slice and scalar arguments.
