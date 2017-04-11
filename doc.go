/*
Package sqlrow is designed to reduce the effort required to implement
common operations performed with SQL databases. It is intended for programmers
who are comfortable with writing SQL, but would like assistance with the
sometimes tedious process of preparing SQL queries for tables that have a
large number of columns, or have a variable number of input parameters.

This package is designed to work seamlessly with the standard library
"database/sql" package. It does not provide any layer on top of *sql.DB
or *sql.Tx. If the calling program has a need to execute queries independently
of this package, it can use "database/sql" directly, or make use of any other
third party package that uses "database/sql".

The following features are provided to simplify writing SQL database queries:
 - Prepare SQL from row structures
 - Autoincrement column values
 - Null columns
 - JSON columns
 - WHERE IN Clauses with multiple values
 - Code generation

Prepare SQL from row structures

Preparing SQL queries with many placeholder arguments is tedious and error-prone. The following
insert query has a dozen placeholders, and it is difficult to match up the columns with the
placeholders. It is not uncommon to have tables with many dozens of columns, at which point the
process of preparing SQL queries using the standard library becomes extremely tiresome.
 insert into users(id,given_name,family_name,dob,ssn,street,locality,postcode,country,phone,mobile,fax)
 values(?,?,?,?,?,?,?,?,?,?,?,?)
This package uses reflection to simplify the construction of SQL statements for insert, update, delete
and select queries. Supplementary information about each database column is stored as a structure tag
in the associated field.
 type User struct {
     ID          int       `sql:"primary key"`
     GivenName   string
     FamilyName  string
     DOB         time.Time `sql:"null"`
     SSN         string
     Street      string
     Locality    string
     Postcode    string
     Country     string
     Phone       string    `sql:"null"`
     Mobile      string    `sql:"null"`
     Facsimile   string    `sql:"fax null"` // "fax" overrides the column name
 }
The calling program creates a schema, which describes rules for generating SQL statements. These
rules include specifying the SQL dialect (eg MySQL, Posgres, SQLite) and the naming convention
used to convert Go struct field names into column names (eg "GivenName" => "given_name"). The schema
is usually created during program initialization.
 schema := NewSchema(
   WithDialect(MySQL),
   WithNamingConvention(SnakeCase),
 )
Once the schema has been defined and a database handle is available (eg *sql.DB, *sql.Tx), it is possible
to create simple row insert/update/delete statements with minimal effort.
 var row User
 // ... populate row with data here and then ...

 // generates the correct SQL to insert a row into the users table
 rowsAffected, err := schema.Exec(db, row, "insert into users({}) values({})")

 // ... and then later on ...

 // generates the correct SQL to update a the matching row in the users table
 rowsAffected, err := schema.Exec(db, row, "update users set {} where {}")
The Exec method parses the SQL query and replaces occurrances of "{}" with the column names
or placeholders that make sense for the SQL clause in which they occur. In the example above,
the insert and update statements would look like:
 insert into users(`id`,`given_name`,`family_name`,`dob`,`ssn`,`street`,`locality`,`postcode`,
 `country`,`phone`,`mobile`,`fax`) values(?,?,?,?,?,?,?,?,?,?,?,?)

 update users set `given_name`=?,`family_name`=?,`dob`=?,`ssn`=?,`street`=?,`locality`=?,
 `postcode`=?,`country`=?,`phone`=?,`mobile`=?,`fax`=? where `id`=?
If the schema is created with a different dialect then the generated SQL will be different.
For example if the Postgres dialect was used the insert and update queries would look more like:
 insert into users("id","given_name","family_name","dob","ssn","street","locality","postcode",
 "country","phone","mobile","fax"") values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)

 update users set "given_name"=$1,"family_name"=$2,"dob"=$3,"ssn"=$4,"street"=$5,"locality"=$6,
 "postcode"=$7,"country"=$8,"phone"=$9,"mobile"=$10,"fax"=$11 where "id"=$12
Select queries are handled in a similar fashion:
 var rows []*User

 // will populate rows slice with the results of the query
 rowCount, err := schema.Select(db, &rows, "select {} from users where postcode = ?", postcode)

 var row User

 // will populate row with the first row returned by the query
 rowCount, err = schema.Select(db, &row, "select {} from users where {}", userID)

 // more complex query involving joins and aliases
 rowCount, err = schema.Select(db, &rows, `
     select {alias u}
     from users u
     inner join user_search_terms ust on ust.user_id = u.id
     where ust.search_term like ?
     order by {alias u}`, searchTermText + "%")
The SQL queries prepared in the above example would look like the following:
 select `id`,`given_name`,`family_name`,`dob`,`ssn`,`street`,`locality`,`postcode`,
 `country`,`phone`,`mobile`,`fax` from users where postcode=?

 select `id`,`given_name`,`family_name`,`dob`,`ssn`,`street`,`locality`,`postcode`,`country`,
 `phone`,`mobile`,`fax` from users where id=?

 select u.`id`,u.`given_name`,u.`family_name`,u.`dob`,u.`ssn`,u.`street`,u.`locality`,
 u.`postcode`,u.`country`,u.`phone`,u.`mobile`,u.`fax` from users u inner join
 user_search_terms ust on ust.user_id = u.id where ust.search_term_like ? order by u.`id`
The examples are using a MySQL dialect. If the schema had been setup for, say, a Postgres
dialect, a generated query would look more like
 select "id","given_name","family_name","dob","ssn","street","locality","postcode","country",
 "phone","mobile","fax" from users where postcode=$1
It is an important point to note that this feature is not about writing the SQL for the programmer.
Rather it is "filling in the blanks" and allowing the programmer to specify as much of the
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
 _, err := schema.Exec(db, row, "insert into table_name({}) values({})")
 if err != nil {
   log.Fatal(err)
 }

 // row.ID will contain the auto-generated value
 fmt.Println(row.ID)
This feature only works with database drivers that support autoincrement columns. The Postgres
driver, in particular, does not support this feature.

Null Columns

Most SQL database tables have columns that are nullable, and it can be tiresome to always
map to pointer types of special nullable types such as sql.NullString. In many cases it is
acceptable to map a database NULL value to the empty value for the corresponding Go struct
field. (NOTE: It is not always acceptable, but experience has shown that it is a common
enough situation).

Where it is acceptable to map a NULL value to an empty value and vice-versa, the Go struct
field can be marked with the "null" keyword in the field's struct tag.
 type User struct {
     ID       int     `sql:"primary key"`
     Name     string
     SpouseID int     `sql:"null"`
     Phone    string  `sql:"null"`
 }
In the above example the `spouse_id` column can be null, but because all IDs are non-zero,
it is unambiguous to map a database NULL to the zero value. Similarly, if the `phone` column
is null it will be mapped to an empty string. An empty string in the Go struct field will
be mapped to NULL in the database.

Care should be taken, because there are cases where an empty value and a database NULL are not
the same thing. There are many cases, however, where this feature can be applied, and result
is simpler code that is easier to read.

JSON Columns

It is not uncommon to serialize complex objects as JSON text for storage in an SQL database.
Native support for JSON is available in some database servers: in partcular Postgres has
excellent support for JSON.

It is straightforward to use this package to serialize a structure field to JSON.
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
 func GetWidgets(db *sql.DB, ids ...int) ([]*Widget, error) {
     var rows []*Widget
     _, err := schema.Select(db, &rows, `select {} from widgets where id in (?)`, ids)
     if err != nil {
       return nil, err
     }
     return widgets, nil
 }
In the above example, the number of placeholders ("?") in the query will be increased to
match the number of values in the `ids` slice. The expansion logic can handle any mix of
slice and scalar arguments.

Code Generation

This package contains a code generation tool in the "./cmd/sqlrow-gen" directory. It can
be quite useful to reduce the amount of code even further.

Performance and Caching

This package makes use of reflection in order to build the SQL that is sent
to the database server, and this imposes a performance penalty. In order
to reduce this overhead each schema instance caches queries generated.
The end result is that the performance of this package is close to the
performance of code that uses hand-constructed SQL queries to call
package "database/sql" directly.

Source Code

More information about this package can be found at https://github.com/jjeffery/sqlrow.
*/
package sqlrow
