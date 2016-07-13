package sqlf_test

import (
	"fmt"

	"github.com/jjeffery/sqlf/sqlf"
)

func ExamplePrepareInsertRow() {
	type User struct {
		ID      int64 `sql:",primary key auto increment"`
		Login   string
		HashPwd string
		Name    string
	}

	stmt := sqlf.PrepareInsertRow(User{}, `
		insert into users({columns}) 
		values({values})
	`)
	fmt.Println(stmt.String())

	stmt = sqlf.PrepareInsertRow(User{}, `
		insert into users({columns including ID}) 
		values({values})
	`)
	fmt.Println(stmt.String())

	// Output:
	// insert into users(`login`,`hash_pwd`,`name`) values(?,?,?)
	// insert into users(`id`, `login`,`hash_pwd`,`name`) values(?,?,?,?)
}

func ExamplePrepareSelectRows() {
	type User struct {
		ID      int64 `sql:",primary key auto increment"`
		Login   string
		HashPwd string
		Name    string
	}

	stmt := sqlf.PrepareSelectRows(User{}, `
		select distinct {columns alias u} 
		from users u
		inner join user_search_terms t on t.user_id = u.id
		where t.search_term like ?
	`)
	fmt.Println(stmt.String())

	// Output:
	// select distinct u.`id`, u.`login`,u.`hash_pwd`,u.`name` from users u inner join user_search_terms t on t.user_id = u.id where t.search_term like ?
}
