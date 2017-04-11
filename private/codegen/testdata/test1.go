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
	schema  *sqlrow.Schema
	rowType *Document
}

// FindModifiedAfter is an example of how to build a method with a custom
// query using the code-generated Select method.
func (q DocumentQuery) FindModifiedAfter(t time.Time) ([]*Document, error) {
	return q.selectRows("select {} from documents where updated_at > ?", t)
}
