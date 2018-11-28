package sqlr

import (
	"reflect"
	"testing"
)

func TestInferRowType(t *testing.T) {
	type Row struct {
		ID int
	}

	tests := []struct {
		row     interface{}
		rowType reflect.Type
		errText string
	}{
		{
			row:     Row{},
			rowType: reflect.TypeOf(Row{}),
		},
		{
			row:     &Row{},
			rowType: reflect.TypeOf(Row{}),
		},
		{
			row:     []Row{},
			rowType: reflect.TypeOf(Row{}),
		},
		{
			row:     []*Row{},
			rowType: reflect.TypeOf(Row{}),
		},
	}

	for i, tt := range tests {
		rowType, err := getRowType(tt.row)
		if err != nil {
			if got, want := err.Error(), tt.errText; got != want {
				t.Errorf("%d: want=%q, got=%q", i, want, got)
			}
			continue
		}
		if got, want := rowType, tt.rowType; got != want {
			t.Errorf("%d: want=%v, got=%v", i, want, got)
		}
	}
}

func TestPrepare(t *testing.T) {
	dialects := map[string]Dialect{
		"mysql":    MySQL,
		"postgres": Postgres,
	}
	tests := []struct {
		row     interface{}
		sql     string
		queries map[string]string
	}{
		{
			row: struct {
				ID   string `sql:"primary key auto increment"`
				Name string
			}{},
			sql: "insert into tbl",
			queries: map[string]string{
				"mysql":    "insert into tbl(`name`) values(?)",
				"postgres": `insert into tbl("name") values($1)`,
			},
		},
		{
			row: struct {
				ID   string `sql:"primary key auto increment"`
				Name string
			}{},
			sql: "insert into tbl({all}) values({})",
			queries: map[string]string{
				"mysql":    "insert into tbl(`id`, `name`) values(?, ?)",
				"postgres": `insert into tbl("id", "name") values($1, $2)`,
			},
		},
		{
			row: struct {
				ID   string `sql:"primary key"`
				Name string
			}{},
			sql: "insert tbl",
			queries: map[string]string{
				"mysql":    "insert into tbl(`id`, `name`) values(?, ?)",
				"postgres": `insert into tbl("id", "name") values($1, $2)`,
			},
		},
		{
			row: struct {
				ID   string `sql:"primary key auto increment"`
				Name string
			}{},
			sql: "update tbl set {} where {}",
			queries: map[string]string{
				"mysql":    "update tbl set `name` = ? where `id` = ?",
				"postgres": `update tbl set "name" = $1 where "id" = $2`,
			},
		},
		{
			row: struct {
				ID   string `sql:"primary key auto increment"`
				Name string
			}{},
			sql: "update tbl",
			queries: map[string]string{
				"mysql":    "update tbl set `name` = ? where `id` = ?",
				"postgres": `update tbl set "name" = $1 where "id" = $2`,
			},
		},
		{
			row: struct {
				ID    string `sql:"primary key auto increment"`
				Hash  string `sql:"pk"`
				Name  string
				Count int
			}{},
			sql: "update [xxx]\nset\n{}\nwhere {}",
			queries: map[string]string{
				"mysql":    "update `xxx` set `name` = ?, `count` = ? where `id` = ? and `hash` = ?",
				"postgres": `update "xxx" set "name" = $1, "count" = $2 where "id" = $3 and "hash" = $4`,
			},
		},
		{
			row: struct {
				ID   string `sql:"primary key auto increment"`
				Name string
			}{},
			sql: "delete from tbl where {}",
			queries: map[string]string{
				"mysql":    "delete from tbl where `id` = ?",
				"postgres": `delete from tbl where "id" = $1`,
			},
		},
		{
			row: struct {
				ID    string `sql:"primary key auto increment"`
				Hash  string `sql:"pk"`
				Name  string
				Count int
			}{},
			sql: "delete from `xxx`\n-- this is a comment\nwhere {}",
			queries: map[string]string{
				"mysql":    "delete from `xxx` where `id` = ? and `hash` = ?",
				"postgres": `delete from "xxx" where "id" = $1 and "hash" = $2`,
			},
		},
		{
			row: struct {
				ID   string `sql:"primary key auto increment"`
				Name string
			}{},
			sql: "select {} from tbl where {}",
			queries: map[string]string{
				"mysql":    "select `id`, `name` from tbl where `id` = ?",
				"postgres": `select "id", "name" from tbl where "id" = $1`,
			},
		},
		{
			row: struct {
				ID   string `sql:"primary key auto increment"`
				Name string
			}{},
			sql: "select {alias t} from tbl t where {pk,alias t}",
			queries: map[string]string{
				"mysql":    "select t.`id`, t.`name` from tbl t where t.`id` = ?",
				"postgres": `select t."id", t."name" from tbl t where t."id" = $1`,
			},
		},
		{
			row: struct {
				ID   string `sql:"primary key auto increment"`
				Home struct {
					Postcode string
				}
			}{},
			sql: "select {alias t} from tbl t where {pk,alias t}",
			queries: map[string]string{
				"mysql":    "select t.`id`, t.`home_postcode` from tbl t where t.`id` = ?",
				"postgres": `select t."id", t."home_postcode" from tbl t where t."id" = $1`,
			},
		},
		{
			row: struct {
				ID    string `sql:"primary key auto increment"`
				Hash  string `sql:"pk"`
				Name  string
				Count int
			}{},
			sql: "select {} from `xxx`\nwhere {}",
			queries: map[string]string{
				"mysql":    "select `id`, `hash`, `name`, `count` from `xxx` where `id` = ? and `hash` = ?",
				"postgres": `select "id", "hash", "name", "count" from "xxx" where "id" = $1 and "hash" = $2`,
			},
		},
	}

	for i, tt := range tests {
		for dialectName, query := range tt.queries {
			dialect := dialects[dialectName]
			schema := NewSchema(WithDialect(dialect))
			stmt, err := schema.Prepare(tt.row, tt.sql)
			if err != nil {
				t.Errorf("%d: expected no error: got %v", i, err)
				continue
			}
			if stmt.String() != query {
				t.Errorf("%d: %s: expected=%q, actual=%q", i, dialect, query, stmt.String())
			}
		}
	}
}
