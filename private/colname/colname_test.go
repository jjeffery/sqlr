package colname

import (
	"fmt"
)

func Example() {
	fmt.Println("")
	fmt.Println("Snake examples")
	fmt.Println("--------------")
	fmt.Println(Snake.ColumnName("UserID"))
	fmt.Println(Snake.ColumnName("HomeAddress"))
	fmt.Println(Snake.ColumnName("StreetName"))
	fmt.Println(Snake.Join("home_address", "street_name"))
	fmt.Println(Snake.ColumnName("HTMLElement"))

	fmt.Println("")
	fmt.Println("Same examples")
	fmt.Println("-------------")
	fmt.Println(Same.ColumnName("UserID"))
	fmt.Println(Same.ColumnName("HomeAddress"))
	fmt.Println(Same.ColumnName("StreetName"))
	fmt.Println(Same.Join("HomeAddress", "StreetName"))
	fmt.Println(Same.ColumnName("HTMLElement"))

	// Output:
	//
	// Snake examples
	// --------------
	// user_id
	// home_address
	// street_name
	// home_address_street_name
	// html_element
	//
	// Same examples
	// -------------
	// UserID
	// HomeAddress
	// StreetName
	// HomeAddressStreetName
	// HTMLElement
}
