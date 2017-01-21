package wherein

import "fmt"

func Example() {
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
	fmt.Printf("sql = %s\nargs = %v", newSQL, newArgs)

	// Output:
	// sql = SELECT * FROM table_name WHERE column1 IN (?,?,?) and column2 = ?
	// args = [101 102 103 abc]
}
