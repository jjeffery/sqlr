package testdata

// Test case: row type without primary key does not generate methods that would require one

//go:generate sqlrow-gen

import (
	"github.com/jjeffery/sqlrow"
	"github.com/jjeffery/sqlrow/private/codegen/testdata/rowtype"
)

type Row3Query struct {
	db      sqlrow.DB
	schema  sqlrow.Schema
	rowType *rowtype.Row3
}
