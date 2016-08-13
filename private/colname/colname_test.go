package colname

import (
	"fmt"
	"testing"
)

func Example() {
	conventions := []Convention{Snake, Same}

	for _, convention := range conventions {
		fmt.Printf("\n%s convention:\n\n", convention.Name())
		fmt.Println(convention.ColumnName("UserID"))
		fmt.Println(convention.ColumnName("HomeAddress"))
		fmt.Println(convention.ColumnName("StreetName"))
		fmt.Println(convention.Join(
			convention.ColumnName("HomeAddress"),
			convention.ColumnName("StreetName")))
		fmt.Println(convention.ColumnName("HTMLElement"))
	}

	// Output:
	//
	// snake convention:
	//
	// user_id
	// home_address
	// street_name
	// home_address_street_name
	// html_element
	//
	// same convention:
	//
	// UserID
	// HomeAddress
	// StreetName
	// HomeAddressStreetName
	// HTMLElement
}

func TestSnakeJoin(t *testing.T) {
	tests := []struct {
		prefix, name, expected string
	}{
		{
			prefix:   "",
			name:     "name",
			expected: "name",
		},
		{
			prefix:   "prefix",
			name:     "",
			expected: "prefix",
		},
		{
			prefix:   "prefix",
			name:     "name",
			expected: "prefix_name",
		},
	}

	for _, tt := range tests {
		actual := Snake.Join(tt.prefix, tt.name)
		if actual != tt.expected {
			t.Errorf("expected=%q, actual=%q", tt.expected, actual)
		}
	}
}
