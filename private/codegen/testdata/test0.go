package testdata

//go:generate sqlrow-gen

import "github.com/jjeffery/sqlrow"

type Row0 struct {
	ID   string `sql:"primary key"`
	Name string `sql:"natural key"`
}

type Row0Query struct {
	db      sqlrow.DB `methods:"Get,Select,SelectOne,Insert,Update,Delete,Upsert"`
	schema  *sqlrow.Schema
	rowType *Row0 `table:"xyz.rows" singular:"document" plural:"documents"`
}
