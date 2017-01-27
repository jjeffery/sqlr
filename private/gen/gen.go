package gen

import (
	"errors"
	"fmt"
)

// Model contains all of the information required by the template
// to generate code.
type Model struct {
	CommandLine string
	Package     string
	Imports     []*Import
	QueryTypes  []*QueryType
}

// Import describes a single import line required for the generated file.
type Import struct {
	Name string // Local name, or blank
	Path string
}

func (imp *Import) String() string {
	if imp.Name != "" {
		return fmt.Sprintf("%s %s", imp.Name, imp.Path)
	}
	return imp.Path
}

// QueryType contains all the information the template needs
// about a struct type for which methods are generated for
// DB queries.
type QueryType struct {
	TypeName        string
	QuotedTableName string // Table name in quotes
	Singular        string // Describes one instance in error msg
	Plural          string // Describes multiple instances in error msg
	DBField         string // Name of the field of type sqlrow.DB (probably db)
	RowType         *RowType
	Method          struct {
		Get       bool
		Select    bool
		SelectOne bool
		Insert    bool
		Update    bool
		Delete    bool
		Upsert    bool
	}
}

// RowType contains all the information the template needs about
// a struct type that is used to represent a single DB table row.
type RowType struct {
	Name     string
	IDArgs   string // for function arguments specifying primary key ID field(s)
	IDParams string // for function parameters specifying primary key ID field(s)
	Keyvals  string // for error messages
}

// Parse the file, and any other related files and build the
// model, which can be used to generate the code.
func Parse(filename string) (*Model, error) {
	return nil, errors.New("not implemented")
}
