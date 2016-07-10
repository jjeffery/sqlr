package column

import (
	"reflect"
	"regexp"
	"strings"
)

// used for parsing tag names
var (
	tagNames = []string{"sql", "db"}
	splitRE  = regexp.MustCompile("[ ,]+")
)

// Info contains information about a database
// column that has been extracted from a struct field
// using reflection.
type Info struct {
	Field         reflect.StructField
	Index         Index
	Path          string
	ColumnName    string
	PrimaryKey    bool
	AutoIncrement bool
	Version       bool
}

func (info *Info) update() {
	for _, key := range tagNames {
		str := info.Field.Tag.Get(key)
		parts := strings.SplitN(str, ",", 2)
		if len(parts) > 1 {
			opts := splitRE.Split(strings.ToLower(parts[1]), -1)
			for i, word := range opts {
				word = strings.TrimSpace(word)
				switch word {
				case "pk", "primary_key":
					info.PrimaryKey = true
				case "primary":
					if i+1 < len(opts) && opts[i+1] == "key" {
						info.PrimaryKey = true
					}
				case "version":
					info.Version = true
				case "autoincr", "identity", "auto_increment":
					info.AutoIncrement = true
				}
			}
		}
	}
}

// columnNameFromTag returns the column name from the field tag,
// or the empty string if none specified.
func columnNameFromTag(tags reflect.StructTag) string {
	for _, key := range tagNames {
		str := tags.Get(key)
		name := strings.TrimSpace(strings.Split(str, ",")[0])
		if name != "" {
			return name
		}
	}
	return ""
}
