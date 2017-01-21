// Package wherein expands SQL statements that have placeholders
// that can accept slices of arguments. This is most commonly
// useful for SQL statements that might look something like
//  SELECT * FROM table_name WHERE column_name IN (?)
// If the argument associated with the placeholder is a slice
// containing (say) three values, then the SQL would be expanded
// to
//  SELECT * FROM table_name where column_name in (?,?,?)
// See the example for more details.
package wherein
