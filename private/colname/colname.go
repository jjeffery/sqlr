// Package colname is concerned with inferring database table
// column names from the names of the associated Go struct fields.
package colname

import (
	"bytes"
	"strings"
	"unicode"
)

// Snake is the default convention, converting Go struct field
// names into "snake case". So the field name "UserID" would be
// converted to "user_id".
var Snake Convention

func init() {
	Snake = Convention{
		name:       "snake",
		columnName: toSnakeCase,
		join:       joinSnake,
	}
}

// Same is an alternative convention, where the database column
// name is identical to the field name.
var Same Convention

func init() {
	Same = Convention{
		name:       "same",
		columnName: columnNameSame,
		join:       joinSame,
	}
}

// Lower is an alternative convention, where the database column name
// is the lower case of the field name. This convention is useful with
// PostgreSQL.
var Lower Convention

func init() {
	Lower = Convention{
		name:       "lower",
		columnName: columnNameLower,
		join:       joinSame,
	}
}

// A Convention provides a naming convention for
// inferring database column names from Go struct field names.
type Convention struct {
	name       string
	columnName func(string) string
	join       func(string, string) string
}

// Name returns the name of the convention.
func (c Convention) Name() string {
	return c.name
}

// ColumnName converts a Go struct field name to a database
// column name.
func (c Convention) ColumnName(fieldName string) string {
	return c.columnName(fieldName)
}

// Join the prefix string to the name to form a column name.
// Used for inferring the database column name for fields
// within embedded structs.
func (c Convention) Join(prefix, name string) string {
	return c.join(prefix, name)
}

func joinSnake(prefix, name string) string {
	if prefix == "" {
		return name
	}
	if name == "" {
		return prefix
	}
	return strings.TrimSuffix(prefix, "_") +
		"_" +
		strings.TrimPrefix(name, "_")
}

func toSnakeCase(name string) string {
	runes := []rune(name)
	n := len(runes)
	var buf bytes.Buffer

	for i := 0; i < n; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < n && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			buf.WriteRune('_')
		}
		buf.WriteRune(unicode.ToLower(runes[i]))
	}

	return buf.String()
}

func columnNameSame(fieldName string) string {
	return fieldName
}

func joinSame(prefix, name string) string {
	return prefix + name
}

func columnNameLower(fieldName string) string {
	return strings.ToLower(fieldName)
}
