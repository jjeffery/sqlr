# The Schema Type

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Creating the Schema](#creating-the-schema)
- [Mapping Individual Field/Column Names](#mapping-individual-fieldcolumn-names)
- [Cloning a Schema](#cloning-a-schema)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

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

## Creating the Schema

A schema is created using the `NewSchema` function, which accepts
a variable number of `SchemaOption` options.

```go
schema := sqlr.NewSchema(
    sqlr.WithDialect(sqlr.Postgres),
    sqlr.WithNamingConvention(sqlr.SnakeCase),
)
```

In most cases it is enough to create a single schema instance
at program initialization, probably at the same point as the
`*sql.DB` instance is being created. In fact it is possible to
specify the dialect directly from the `*sql.DB` handle:

```go
db, err := sql.Open("postgres", "postgres://user:pwd@localhost/mydb")
if err != nil {
    log.Fatal(err)
}

// create a new schema whose dialect matches the db, with default
// naming convention
schema := sqlr.NewSchema(
    WithDB(db), // will choose the correct dialect for db
    WithNamingConvention(sqlr.LowerCase),
)
```

*NOTE* It not all that expensive to create a schema, but it a better 
idea to create a schema at program initialization and re-use it rather 
than create one whenever necessary. The reason for this is that the 
schema caches information about Go structures and SQL statements created to 
improve performance.

## Mapping Individual Field/Column Names

Sometimes it is necessary to provide a special naming convention for
creating a column name from a Go struct field. When this is necessary,
it can be achieved using the `WithField` schema option. Take the following
(simplified) table:

```sql
create table contact_details(
    id int not null primary key,
    work_email text,
    work_fax text,
    home_email text,
    home_fax_number text
)
```
There is some inconsistency in the column names, but we want to use a common
structure to handle the contact details:

```go
type LocationContact struct {
    Email     string
    Facsimile string
}

type ContactDetail struct {
    ID   int              `sql:"primary key"`
    Home LocationContact
    Work LocationContact
}
```

When it comes to wanting to map the names for the fax numbers, it turns out
not to be possible to use the Go struct tag: it is not possible to specify
the names of two different columns on the same Go struct field.

The solution is to specify the individual column names when creating the
schema.

```go
schema := NewSchema(
    WithDialect(sqlr.SQLite),
    WithNamingConvention(sqlr.Snake),
    WithField("Work.Facsimile", "work_fax"),
    WithField("Home.Facsimile", "home_fax_number"),
)
```

## Cloning a Schema

Once created a schema is immutable. The reason for this is that schemas
have an internal cache to help with performance, and changing the dialect
or naming convention would render this cache invalid.

It is not anticipated that this is a common requirement, but in some cases
it is useful to be able to clone a schema, which provides a deep copy of
the configuration and allows for further configuration:

```go

// ... schema has already been created elsewhere ...

schemaForContactDetails := schema.Clone(
    WithField("Work.Facsimile", "work_fax"),
    WithField("Home.Facsimile", "home_fax_number"),
)
```

Looking at more unusual cases, it would be possible to keep separate schemas 
for a database whose tables have evolved with different naming conventions:

```go
var schema struct {
    Base   *sqlr.Schema
    Table1 *sqlr.Schema
    Table2 *sqlr.Schema
}

func init() {
    // The original database was created using `SameCase`, as many MS SQL DBs are.
    schema.Base := NewSchema(
        WithDialect(sqlr.MSSQL),
        WithNamingConvention(sqlr.SameCase),
    )

    // Table1 got added later by someone who preferred `snake_case` and thought
    // it would be better to use it than be consistent. (This can happen...)
    schema.Table1 := schema.Base.Clone(
        WithNamingConvention(sqlr.SnakeCase),
    )

    // Table2 was added, but there are some exceptions to the naming rules that
    // are not universal across the database.
    schema.Table2 := schema.Base.Clone(
        WithField("Home.Locality", "HomeSuburb"),
        WithField("ID", "table2_id"),
    )
}
```