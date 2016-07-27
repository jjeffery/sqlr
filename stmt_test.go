package sqlstmt

import (
	"testing"
)

func TestNewInsertRowStmt(t *testing.T) {
	defer func() { Default.Dialect = nil }()
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
			sql: "tbl",
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
				"mysql":    "insert into tbl(`id`,`name`) values(?,?)",
				"postgres": `insert into tbl("id","name") values($1,$2)`,
			},
		},
		{
			row: struct {
				ID   string `sql:"primary key"`
				Name string
			}{},
			sql: "tbl",
			queries: map[string]string{
				"mysql":    "insert into tbl(`id`,`name`) values(?,?)",
				"postgres": `insert into tbl("id","name") values($1,$2)`,
			},
		},
	}

	for _, tt := range tests {
		for dialect, query := range tt.queries {
			Default.Dialect = NewDialect(dialect)
			stmts := []*InsertRowStmt{
				NewInsertRowStmt(tt.row, tt.sql),
				Default.NewInsertRowStmt(tt.row, tt.sql),
			}
			for _, stmt := range stmts {
				if stmt.String() != query {
					t.Errorf("%s: expected=%q, actual=%q", dialect, query, stmt.String())
				}
			}
		}
	}
}

func TestNewUpdateRowStmt(t *testing.T) {
	defer func() { Default.Dialect = nil }()
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
			sql: "tbl",
			queries: map[string]string{
				"mysql":    "update tbl set `name`=? where `id`=?",
				"postgres": `update tbl set "name"=$1 where "id"=$2`,
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
				"mysql":    "update `xxx` set `name`=?,`count`=? where `id`=? and `hash`=?",
				"postgres": `update "xxx" set "name"=$1,"count"=$2 where "id"=$3 and "hash"=$4`,
			},
		},
	}

	for _, tt := range tests {
		for dialect, query := range tt.queries {
			Default.Dialect = NewDialect(dialect)
			stmts := []*ExecRowStmt{
				NewUpdateRowStmt(tt.row, tt.sql),
				Default.NewUpdateRowStmt(tt.row, tt.sql),
			}
			for _, stmt := range stmts {
				if stmt.String() != query {
					t.Errorf("%s: expected=%q, actual=%q", dialect, query, stmt.String())
				}
			}
		}
	}
}

func TestNewDeleteRowStmt(t *testing.T) {
	defer func() { Default.Dialect = nil }()
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
			sql: "tbl",
			queries: map[string]string{
				"mysql":    "delete from tbl where `id`=?",
				"postgres": `delete from tbl where "id"=$1`,
			},
		},
		{
			row: struct {
				ID    string `sql:"primary key auto increment"`
				Hash  string `sql:"pk"`
				Name  string
				Count int
			}{},
			sql: "delete from `xxx`\nwhere {}",
			queries: map[string]string{
				"mysql":    "delete from `xxx` where `id`=? and `hash`=?",
				"postgres": `delete from "xxx" where "id"=$1 and "hash"=$2`,
			},
		},
	}

	for _, tt := range tests {
		for dialect, query := range tt.queries {
			Default.Dialect = NewDialect(dialect)
			stmts := []*ExecRowStmt{
				NewDeleteRowStmt(tt.row, tt.sql),
				Default.NewDeleteRowStmt(tt.row, tt.sql),
			}
			for _, stmt := range stmts {
				if stmt.String() != query {
					t.Errorf("%s: expected=%q, actual=%q", dialect, query, stmt.String())
				}
			}
		}
	}
}

func TestNewGetRowStmt(t *testing.T) {
	defer func() { Default.Dialect = nil }()
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
			sql: "tbl",
			queries: map[string]string{
				"mysql":    "select `id`,`name` from tbl where `id`=?",
				"postgres": `select "id","name" from tbl where "id"=$1`,
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
				"mysql":    "select `id`,`hash`,`name`,`count` from `xxx` where `id`=? and `hash`=?",
				"postgres": `select "id","hash","name","count" from "xxx" where "id"=$1 and "hash"=$2`,
			},
		},
	}

	for _, tt := range tests {
		for dialect, query := range tt.queries {
			Default.Dialect = NewDialect(dialect)
			stmts := []*GetRowStmt{
				NewGetRowStmt(tt.row, tt.sql),
				Default.NewGetRowStmt(tt.row, tt.sql),
			}
			for _, stmt := range stmts {
				if stmt.String() != query {
					t.Errorf("%s: expected=%q, actual=%q", dialect, query, stmt.String())
				}
			}
		}
	}
}

func TestNewSelectStmt(t *testing.T) {
	defer func() { Default.Dialect = nil }()
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
			sql: "select {} from tbl where name like ? order by {}",
			queries: map[string]string{
				"mysql":    "select `id`,`name` from tbl where name like ? order by `id`",
				"postgres": `select "id","name" from tbl where name like $1 order by "id"`,
			},
		},
		{
			row: struct {
				ID   string `sql:"primary key auto increment"`
				Name string
			}{},
			sql: "select {} from tbl where name like ? order by {pk}",
			queries: map[string]string{
				"mysql":    "select `id`,`name` from tbl where name like ? order by `id`",
				"postgres": `select "id","name" from tbl where name like $1 order by "id"`,
			},
		},
		{
			row: struct {
				ID   string `sql:"primary key auto increment"`
				Name struct {
					Given  string
					Family string
				}
			}{},
			sql: "select {} from tbl where name like ? order by {pk}",
			queries: map[string]string{
				"mysql":    "select `id`,`name_given`,`name_family` from tbl where name like ? order by `id`",
				"postgres": `select "id","name_given","name_family" from tbl where name like $1 order by "id"`,
			},
		},

		{
			row: struct {
				ID    string `sql:"primary key auto increment"`
				Hash  string `sql:"pk"`
				Name  string
				Count int
			}{},
			sql: "select {} from `xxx`\nwhere id=? and hash like ?",
			queries: map[string]string{
				"mysql":    "select `id`,`hash`,`name`,`count` from `xxx` where id=? and hash like ?",
				"postgres": `select "id","hash","name","count" from "xxx" where id=$1 and hash like $2`,
			},
		},
	}

	for _, tt := range tests {
		for dialect, query := range tt.queries {
			Default.Dialect = NewDialect(dialect)
			stmts := []*SelectStmt{
				NewSelectStmt(tt.row, tt.sql),
				Default.NewSelectStmt(tt.row, tt.sql),
			}
			for _, stmt := range stmts {
				if stmt.String() != query {
					t.Errorf("%s:\nexpected=%q,\n  actual=%q", dialect, query, stmt.String())
				}
			}
		}
	}
}
