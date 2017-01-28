package column

import (
	"reflect"
	"strings"

	"github.com/jjeffery/sqlrow/private/scanner"
)

// Info contains information about a database
// column that has been extracted from a struct field
// using reflection.
type Info struct {
	Field         reflect.StructField
	Index         Index
	Path          Path
	PrimaryKey    bool
	AutoIncrement bool
	Version       bool
	JSON          bool
}

func newInfo(field reflect.StructField) *Info {
	info := &Info{
		Field: field,
	}
	info.updateOptsFromTag()
	return info
}

func (info *Info) updateOptsFromTag() {
	scan := newScanner(info.Field.Tag)
	if scan == nil {
		return
	}
	for scan.Scan() {
		tok, lit := scan.Token(), scan.Text()
		switch tok {
		case scanner.KEYWORD:
			switch strings.ToLower(lit) {
			case "pk", "primary_key":
				info.PrimaryKey = true
			case "autoincrement", "autoincr":
				info.AutoIncrement = true
			case "primary":
				if scan.Scan(); strings.ToLower(scan.Text()) == "key" {
					info.PrimaryKey = true
				}
			case "auto":
				if scan.Scan(); strings.ToLower(scan.Text()) == "increment" {
					info.AutoIncrement = true
				}
			case "identity":
				info.AutoIncrement = true
			case "version":
				info.Version = true
			case "json", "jsonb":
				info.JSON = true
			}
		}
	}
}

// columnNameFromTag returns the column name from the field tag,
// or the empty string if none specified.
func columnNameFromTag(tags reflect.StructTag) string {
	scan := newScanner(tags)
	if scan == nil {
		return ""
	}
	for scan.Scan() {
		tok, lit := scan.Token(), scan.Text()
		switch tok {
		case scanner.KEYWORD:
			// exit on first keyword, no column specified
			return ""
		case scanner.IDENT:
			// first identifier indicates the column name, and
			// may be quoted
			return scanner.Unquote(lit)
		case scanner.LITERAL:
			if scanner.IsQuoted(lit) {
				// a string literal is accepted as the column name
				return scanner.Unquote(lit)
			}
		case scanner.OP:
			if lit == "-" {
				// indicates should not be a column
				return lit
			}
		}
	}
	return ""
}

// newScanner returns a scanner for reading the contents of the struct tag.
// Returns nil if there is no appropriate struct tag to read.
func newScanner(tag reflect.StructTag) *scanner.Scanner {
	for _, key := range []string{"sqlrow", "sql"} {
		str := strings.TrimSpace(tag.Get(key))
		if str != "" {
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
				"natural_key")
			return scan
		}
	}
	return nil
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
