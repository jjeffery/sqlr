// Package colname contains functions for deriving
// a database table column name from a Go struct
// field name.
//
// Each function accepts a field name and an optional
// prefix. The prefix will be non-blank when naming
// a field within an embedded structure.
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
		columnName: toSnakeCase,
		join:       joinSnake,
	}
}

// Same is an alternative convention, where the database column
// name is identical to the field name.
var Same Convention

func init() {
	Same = Convention{
		columnName: columnNameSame,
		join:       joinSame,
	}
}

// A Convention provides a naming convention for
// inferring database column names from Go struct field names.
type Convention struct {
	columnName func(string) string
	join       func(string, string) string
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
