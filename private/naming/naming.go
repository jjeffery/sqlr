// Package naming is concerned with naming database tables
// and columns.
package naming

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
		key:     "snake",
		convert: toSnakeCase,
		join:    joinSnake,
	}
}

// Same is an alternative convention, where the database column
// name is identical to the field name.
var Same Convention

func init() {
	Same = Convention{
		key:     "same",
		convert: convertToSame,
		join:    joinSame,
	}
}

// Lower is an alternative convention, where the database column name
// is the lower case of the field name. This convention is useful with
// PostgreSQL.
var Lower Convention

func init() {
	Lower = Convention{
		key:     "lower",
		convert: convertToLower,
		join:    joinSame,
	}
}

// A Convention provides a naming convention for
// inferring database column names from Go struct field names.
type Convention struct {
	key     string
	convert func(string) string
	join    func([]string) string
}

// Key returns the key that describes the convention.
// This key is used to locate the struct tag key value.
func (c Convention) Key() string {
	return c.key
}

// Convert converts a Go struct field name according to the naming convention.
func (c Convention) Convert(fieldName string) string {
	return c.convert(fieldName)
}

// Join the prefix string to the name to form a column name.
// Used for inferring the database column name for fields
// within embedded structs.
func (c Convention) Join(names []string) string {
	return c.join(names)
}

func joinSnake(names []string) string {
	return strings.Join(names, "_")
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

func convertToSame(fieldName string) string {
	return fieldName
}

func joinSame(names []string) string {
	return strings.Join(names, "")
}

func convertToLower(fieldName string) string {
	return strings.ToLower(fieldName)
}
