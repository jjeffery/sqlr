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

// SameCaseConvention implements NamingConvention. It does not alter field names.
type SameCaseConvention struct{}

// Convert returns fieldName unchanged.
func (lc SameCaseConvention) Convert(fieldName string) string {
	return fieldName
}

// Join joins together the names with no separating characters between them.
func (lc SameCaseConvention) Join(names []string) string {
	return strings.Join(names, "")
}
