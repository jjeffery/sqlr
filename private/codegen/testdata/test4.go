package testdata

// Test case: row type with a custom primary key defined in the same package as the row

//go:generate sqlr-gen

import (
	"github.com/jjeffery/sqlr"
	"github.com/jjeffery/sqlr/private/codegen/testdata/rowtype"
)

type Row4Query struct {
	db      sqlr.DB
	schema  *sqlr.Schema
	rowType *rowtype.Row4
}
