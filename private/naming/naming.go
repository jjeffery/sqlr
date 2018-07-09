// Package naming provides naming conventions used to convert
// Go struct field names to database columns.
package naming

import (
	"bytes"
	"strings"
	"unicode"
)

// Instances of the different naming conventions
var (
	SnakeCase SnakeCaseConvention
	LowerCase LowerCaseConvention
	SameCase  SameCaseConvention
)

// SnakeCaseConvention converts Go struct fields into "snake_case".
// So the field name "UserID" would be converted to "user_id".
type SnakeCaseConvention struct{}

// Convert converts fieldName into snake_case.
func (sc SnakeCaseConvention) Convert(name string) string {
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

// Join joins together the names with underscores.
func (sc SnakeCaseConvention) Join(names []string) string {
	return strings.Join(names, "_")
}

// TableName converts a type name into a table name.
func (sc SnakeCaseConvention) TableName(typeName string) string {
	return sc.Convert(typeName)
}

// LowerCaseConvention implements NamingConvention, by converting to lower case.
type LowerCaseConvention struct{}

// Convert converts the field name to lower case.
func (lc LowerCaseConvention) Convert(fieldName string) string {
	return strings.ToLower(fieldName)
}

// Join joins together the names with no separating characters between them.
func (lc LowerCaseConvention) Join(names []string) string {
	return strings.Join(names, "")
}

// TableName converts a type name into a table name.
func (lc LowerCaseConvention) TableName(typeName string) string {
	return lc.Convert(typeName)
}

// SameCaseConvention implements NamingConvention. It does not alter field names.
type SameCaseConvention struct{}

// Convert returns fieldName unchanged.
func (sc SameCaseConvention) Convert(fieldName string) string {
	return fieldName
}

// Join joins together the names with no separating characters between them.
func (sc SameCaseConvention) Join(names []string) string {
	return strings.Join(names, "")
}

// TableName converts a type name into a table name.
func (sc SameCaseConvention) TableName(typeName string) string {
	return sc.Convert(typeName)
}
