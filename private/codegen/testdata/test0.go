package testdata

//go:generate sqlr-gen

import "github.com/jjeffery/sqlr"

type Row0 struct {
	ID   string `sql:"primary key"`
	Name string `sql:"natural key"`
}

type Row0Query struct {
	querier sqlr.Querier `methods:"Get,Select,SelectRow,Insert,Update,Delete,Upsert"`
	schema  *sqlr.Schema
	rowType *Row0 `table:"xyz.rows" singular:"document" plural:"documents"`
}
