package testdata

//go:generate sqlrow-gen

import (
	"time"

	"github.com/jjeffery/sqlrow"
)

type Document struct {
	ID        string `sql:"primary key"`
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DocumentQuery struct {
	db      sqlrow.DB
	schema  sqlrow.Schema
	rowType *Document
}
