package testdata

// Test case: row type without primary key does not generate methods that would require one

//go:generate sqlr-gen

import (
	"github.com/jjeffery/sqlr"
	"github.com/jjeffery/sqlr/private/codegen/testdata/rowtype"
)

type Row3Query struct {
	db      sqlr.Querier
	schema  *sqlr.Schema
	rowType *rowtype.Row3
}
