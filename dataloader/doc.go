/*
Package dataloader provides an implementation of the
data loader pattern, which is useful for batching up requests
to a database rather than making a large number of small queries.
This can result in significant performance improvements.
The dataloader pattern is particularly useful when implementing
GraphQL servers.

This package makes use of reflection in order to make things as
simple as possible for the calling program.

To create a loader function, you need a query function and a key
function. These are described in detail below, but assuming you
have a query function and a key function, creating a loader function
is as simple as:

 var loader func(key int) func() (*Row, error)
 Make(&loader, queryFunc, keyFunc)

Read on for more detail about the loader function, and query and
key functions.

Loader Function

A loader function looks something like the following:
 var loader func(key int) func() (*Row, error)
Where key is a key identifying the row to be loaded. The
returned value is a function that, when called, will return
the row associated with the key (or an error). This returned
function is known as a "thunk". It is a common practice to
define a custom type for the thunk, in which case the loader
function would looks like:
 type RowThunk func() (*Row, error)

 var loader func(key int) RowThunk
Loader functions can use any string or integral type as a key,
or any custom type that is based on a string or integral type:
 type RowID string
 type RowThunk func() (*Row, error)

 var loader func(id RowID) RowThunk

Query Functions and Key Functions

A loader function is created using the "Make" function, and it
needs two functions: a query function and a key function.

The query function accepts a slice of keys and returns a slice
of rows. The order of the returned rows is arbitrary. A query
function for the loader function above would look like:
 func performQuery(id []RowID) ([]*Row, error) {
	 // ... perform database query here ...
 }

Because the query function can return rows in any order, it is
necessary to provide a function that, given a row, will return
the key associated with the row. The key function is usually
very simple and looks like:
 func getKey(row *Row) RowID {
	 return row.ID
 }
*/
package dataloader
