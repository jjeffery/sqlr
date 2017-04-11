package testdata

// Test case: row type without primary key does not generate methods that would require one

//go:generate sqlr-gen

import (
	"time"

	"github.com/jjeffery/sqlr"
)

type Row2 struct {
	ID   string
	Name string
	DOB  time.Time
}

type Row2Query struct {
	db      sqlr.DB
	schema  *sqlr.Schema
	rowType *Row2
}
