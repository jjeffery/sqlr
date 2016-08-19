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
				"jsonb")
			return scan
		}
	}
	return nil
}
