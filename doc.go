/*
Package sqlr is designed to reduce the effort required to work with SQL databases.
It is intended for programmers who are comfortable with writing SQL, but would
like assistance with the sometimes tedious process of preparing SQL queries for
tables that have a large number of columns, or have a variable number of input parameters.

This GoDoc summary provides an overview of how to use this package. For
more detailed documentation, see https://jjeffery.github.io/sqlr.

Prepare SQL queries based on row structures

Preparing SQL queries with many placeholder arguments is tedious and error-prone. The following
insert query has a dozen placeholders, and it is difficult to match up the columns with the
placeholders. It is not uncommon to have tables with many more columns than this example, and the
level of difficulty increases with the number of columns in the table.
 insert into users(id,given_name,family_name,dob,ssn,street,locality,postcode,
 country,phone,mobile,fax) values(?,?,?,?,?,?,?,?,?,?,?,?)
This package uses reflection to simplify the construction of SQL queries. Supplementary information
about each database column is stored in the structure tag of the associated field.
 type User struct {
     ID          int       `sql:"primary key"`
     GivenName   string
     FamilyName  string
     DOB         time.Time
     SSN         string
     Street      string
     Locality    string
     Postcode    string
     Country     string
     Phone       string
     Mobile      string
     Facsimile   string    `sql:"fax"` // "fax" overrides the column name
 }
The calling program creates a schema, which describes rules for generating SQL statements. These
rules include specifying the SQL dialect (eg MySQL, Postgres, SQLite) and the naming convention
used to convert Go struct field names into column names (eg "GivenName" => "given_name"). The schema
is usually created during program initialization. Once created, a schema is immutable and can be
called concurrently from multiple goroutines.
 schema := NewSchema(
   WithDialect(MySQL),
   WithNamingConvention(SnakeCase),
 )
A session is created using a context, a database connection (eg *sql.DB, *sql.Tx, *sql.Conn), and a
schema. A session is inexpensive to create, and is intended to last no longer than a single request
(which might be a HTTP request, in the case of a HTTP server). A session is bounded by the lifetime
of its context.
 session := NewSession(ctx, tx, schema)
Once a session has been created, it is possible to create simple row insert/update statements with
minimal effort.
 var row User
 // ... populate row with data here and then ...

 // generates the correct SQL to insert a row into the users table
 result, err := session.Row(row).Exec("insert into users({}) values({})")

 // ... and then later on ...

 // generates the correct SQL to update a the matching row in the users table
 result, err := session.Row(row).Exec("update users set {} where {}")
The Exec method parses the SQL query and replaces occurrences of "{}" with the column names
or placeholders that make sense for the SQL clause in which they occur. In the example above,
the insert and update statements would look like:
 insert into users(`id`,`given_name`,`family_name`,`dob`,`ssn`,`street`,
 `locality`,`postcode`,`country`,`phone`,`mobile`,`fax`) values(?,?,?,?,
 ?,?,?,?,?,?,?,?)

 update users set `given_name`=?,`family_name`=?,`dob`=?,`ssn`=?,`street`=?,
 `locality`=?,`postcode`=?,`country`=?,`phone`=?,`mobile`=?,`fax`=? where `id`=?
If the schema is created with a different dialect then the generated SQL will be different.
For example if the Postgres dialect was used the insert and update queries would look more like:
 insert into users("id","given_name","family_name","dob","ssn","street","locality",
 "postcode","country","phone","mobile","fax") values($1,$2,$3,$4,$5,$6,$7,$8,$9,
 $10,$11,$12)

 update users set "given_name"=$1,"family_name"=$2,"dob"=$3,"ssn"=$4,"street"=$5,
 "locality"=$6,"postcode"=$7,"country"=$8,"phone"=$9,"mobile"=$10,"fax"=$11
 where "id"=$12
Inserting and updating a single row are common enough operations that the session has methods
that make it very simple:
 session.InsertRow(row)
 session.UpdateRow(row)
Select queries can be performed using the session's Select method:
 var rows []*User

 // will populate rows slice with the results of the query
 rowCount, err := session.Select(&rows, "select {} from users where postcode = ?", postcode)

 var row User

 // will populate row with the first row returned by the query
 rowCount, err = session.Select(&row, "select {} from users where {}", userID)

 // more complex query involving joins and aliases
 rowCount, err = session.Select(&rows, `
     select {alias u}
     from users u
     inner join user_search_terms ust on ust.user_id = u.id
     where ust.search_term like ?
     order by {alias u}`, searchTermText + "%")
The SQL queries prepared in the above example would look like the following:
 select `id`,`given_name`,`family_name`,`dob`,`ssn`,`street`,`locality`,
 `postcode`,`country`,`phone`,`mobile`,`fax` from users where postcode=?

 select `id`,`given_name`,`family_name`,`dob`,`ssn`,`street`,`locality`,
 `postcode`,`country`,`phone`,`mobile`,`fax` from users where id=?

 select u.`id`,u.`given_name`,u.`family_name`,u.`dob`,u.`ssn`,u.`street`,
 u.`locality`,u.`postcode`,u.`country`,u.`phone`,u.`mobile`,u.`fax`
 from users u inner join user_search_terms ust on ust.user_id = u.id where
 ust.search_term_like ? order by u.`id`
The examples are using a MySQL dialect. If the schema had been setup for, say, a Postgres
dialect, a generated query would look more like:
 select "id","given_name","family_name","dob","ssn","street","locality",
 "postcode","country","phone","mobile","fax" from users where postcode=$1
It is an important point to note that this feature is not about writing the SQL for the programmer.
Rather it is about "filling in the blanks": allowing the programmer to specify as much of the
SQL query as they want without having to write the tiresome bits.

Autoincrement Column Values

When inserting rows, if a column is defined as an autoincrement column, then the generated
value will be retrieved from the database server, and the corresponding field in the row
structure will be updated.
 type Row {
   ID   int    `sql:"primary key autoincrement"`
   Name string
 }

 row := &Row{Name: "some name"}
 _, err := session.InsertRow(row)
 if err != nil {
   log.Fatal(err)
 }

 // row.ID will contain the auto-generated value
 fmt.Println(row.ID)
Autoincrement column values work for all supported databases (PostgreSQL, MySQL,
Microsoft SQL Server and SQLite).

Null Columns

Most SQL database tables have columns that are nullable, and it can be tiresome to always
map to pointer types or special nullable types such as sql.NullString. In many cases it is
acceptable to map the zero value for the field a database NULL in the corresponding database
column.

Where it is acceptable to map a zero value to a NULL database column, the Go struct field can
be marked with the "null" keyword in the field's struct tag.
 type Employee struct {
     ID        int     `sql:"primary key"`
     Name      string
     ManagerID int     `sql:"null"`
     Phone     string  `sql:"null"`
 }
In the above example the `manager_id` column can be null, but if all valid IDs are non-zero,
it is unambiguous to map the zero value to a database NULL. Similarly, if the `phone` column
an empty string it will be stored as a NULL in the database.

Care should be taken, because there are cases where a zero value and a database NULL do not
represent the same thing. There are many cases, however, where this feature can be applied,
and the result is simpler code that is easier to read.

JSON Columns

It is not uncommon to serialize complex objects as JSON text for storage in an SQL database.
Native support for JSON is available in some database servers: in partcular Postgres has
excellent support for JSON.

It is straightforward to use this package to serialize a structure field to JSON:
 type SomethingComplex struct {
     Name       string
     Values     []int
     MoreValues map[string]float64
     // ... and more fields here ...
 }

 type Row struct {
     ID    int                `sql:"primary key"`
     Name  string
     Cmplx *SomethingComplex  `sql:"json"`
 }
In the example above the `Cmplx` field will be marshaled as JSON text when
writing to the database, and unmarshaled into the struct when reading from
the database.

WHERE IN Clauses with Multiple Values

While most SQL queries accept a fixed number of parameters, if the SQL query
contains a `WHERE IN` clause, it requires additional string manipulation to match
the number of placeholders in the query with args.

This package simplifies queries with a variable number of arguments. When processing
an SQL query, it detects if any of the arguments are slices:
 // GetWidgets returns all the widgets associated with the supplied IDs.
 func GetWidgets(session *sqlr.Session, ids ...int) ([]*Widget, error) {
     var rows []*Widget
     _, err := session.Select(&rows, `select {} from widgets where id in (?)`, ids)
     if err != nil {
       return nil, err
     }
     return widgets, nil
 }
In the above example, the number of placeholders ("?") in the query will be increased to
match the number of values in the `ids` slice. The expansion logic can handle any mix of
slice and scalar arguments.

Type-Safe Query Functions

A session can create type-safe query functions. This is a very powerful feature and makes
it very easy to create type-safe data access. See the Session.MakeQuery function for examples.
*/
package sqlr
