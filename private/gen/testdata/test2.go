package testdata

// Test case: row type without primary key does not generate methods that would require one

//go:generate sqlrow-gen

import (
	"time"

	"github.com/jjeffery/sqlrow"
)

type Row2 struct {
	ID   string
	Name string
	DOB  time.Time
}

type Row2Query struct {
	db      sqlrow.DB
	schema  sqlrow.Schema
	rowType *Row2
}
