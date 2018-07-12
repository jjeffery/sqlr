package column

import (
	"reflect"
	"strings"

	"github.com/jjeffery/sqlr/private/scanner"
)

// Info contains information about a database
// column that has been extracted from a struct field
// using reflection.
type Info struct {
	Field      reflect.StructField
	Index      Index
	Path       Path
	FieldNames string  // one or more field names, joined by periods
	Tag        TagInfo // meta data from the struct field tag
}

func newInfo(field reflect.StructField) *Info {
	info := &Info{
		Field: field,
		Tag:   ParseTag(field.Tag),
	}
	return info
}

// newScanner returns a scanner for reading the contents of the struct tag.
// Returns nil if there is no appropriate struct tag to read.
func newScanner(tag reflect.StructTag) *scanner.Scanner {
	for _, key := range structTagKeys {
		str := strings.TrimSpace(tag.Get(key))
		if str != "" {
			return newScannerForString(str)
		}
	}
	return nil
}

func newScannerForString(str string) *scanner.Scanner {
	scan := scanner.New(strings.NewReader(str))
	scan.IgnoreWhiteSpace = true
	scan.AddKeywords(
		"pk",
		"primary_key",
		"primary",
		"autoincrement",
		"autoincr",
		"auto",
		"identity",
		"version",
		"json",
		"jsonb",
		"natural",
		"natural_key",
		"null",
		"omitempty",
		"emptynull")
	return scan
}

// TagInfo is information obtained about a column from the
// struct tags of its corresponding field.
type TagInfo struct {
	Ignore        bool
	Name          string
	PrimaryKey    bool
	AutoIncrement bool
	Version       bool
	JSON          bool
	NaturalKey    bool
	EmptyNull     bool
}

// ParseTag returns a TagInfo containing information obtained from the
// StructTag of the field associated with the column.
func ParseTag(tag reflect.StructTag) TagInfo {
	var tagInfo TagInfo

	scan := newScanner(tag)
	if scan == nil {
		return tagInfo
	}
	var hadKeyword bool
	for scan.Scan() {
		tok, lit := scan.Token(), scan.Text()
		switch tok {
		case scanner.KEYWORD:
			hadKeyword = true
			switch strings.ToLower(lit) {
			case "pk", "primary_key":
				tagInfo.PrimaryKey = true
			case "autoincrement", "autoincr":
				tagInfo.AutoIncrement = true
			case "primary":
				if scan.Scan(); strings.ToLower(scan.Text()) == "key" {
					tagInfo.PrimaryKey = true
				}
			case "auto":
				if scan.Scan(); strings.ToLower(scan.Text()) == "increment" {
					tagInfo.AutoIncrement = true
				}
			case "identity":
				tagInfo.AutoIncrement = true
			case "version":
				tagInfo.Version = true
			case "json", "jsonb":
				tagInfo.JSON = true
			case "natural_key":
				tagInfo.NaturalKey = true
			case "natural":
				if scan.Scan(); strings.ToLower(scan.Text()) == "key" {
					tagInfo.NaturalKey = true
				}
			case "null", "omitempty", "emptynull":
				tagInfo.EmptyNull = true
			}
		case scanner.IDENT:
			if !hadKeyword && tagInfo.Name == "" {
				tagInfo.Name = scanner.Unquote(lit)
			}
		case scanner.LITERAL:
			if !hadKeyword && tagInfo.Name == "" && scanner.IsQuoted(lit) {
				// a string literal is accepted as the column name
				tagInfo.Name = scanner.Unquote(lit)
			}
		case scanner.OP:
			if !hadKeyword && tagInfo.Name == "" && lit == "-" {
				// indicates should not be a column
				tagInfo.Ignore = true
			}
		}
	}
	return tagInfo
}
