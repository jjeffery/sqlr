package wherein

import "fmt"

func ExampleExpand() {
	// Using MySQL/SQLite style positional placeholders
	{
		sql := "SELECT * FROM table_name WHERE column1 IN (?) and column2 = ?"
		args := []interface{}{
			[]int{101, 102, 103},
			"abc",
		}
		newSQL, newArgs, err := Expand(sql, args)
		if err != nil {
			fmt.Print(err)
			return
		}
		fmt.Printf("sql = %s\nargs = %v\n\n", newSQL, newArgs)
	}

	// Using PostgreSQL style numbered placeholders
	{
		sql := "SELECT * FROM table_name WHERE column1 IN ($1) and column2 = $2"
		args := []interface{}{
			[]int{101, 102, 103},
			"abc",
		}
		newSQL, newArgs, err := Expand(sql, args)
		if err != nil {
			fmt.Print(err)
			return
		}
		fmt.Printf("sql = %s\nargs = %v\n\n", newSQL, newArgs)
	}

	// Using MySQL/SQLite style numbered placeholders
	{
		sql := "SELECT * FROM table_name WHERE column1 IN (?2) and column2 IN (?1)"
		args := []interface{}{
			[]int{101, 102, 103},
			[]string{"abc", "def", "ghi"},
		}
		newSQL, newArgs, err := Expand(sql, args)
		if err != nil {
			fmt.Print(err)
			return
		}
		fmt.Printf("sql = %s\nargs = %v\n\n", newSQL, newArgs)
	}

	// Output:
	// sql = SELECT * FROM table_name WHERE column1 IN (?,?,?) and column2 = ?
	// args = [101 102 103 abc]
	//
	// sql = SELECT * FROM table_name WHERE column1 IN ($1,$2,$3) and column2 = $4
	// args = [101 102 103 abc]
	//
	// sql = SELECT * FROM table_name WHERE column1 IN (?4,?5,?6) and column2 IN (?1,?2,?3)
	// args = [101 102 103 abc def ghi]
}
