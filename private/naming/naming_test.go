package naming

import (
	"fmt"
	"testing"
)

func Example() {
	conventions := []Convention{Snake, Same, Lower}

	for _, convention := range conventions {
		fmt.Printf("\n%s convention:\n\n", convention.Key())
		fmt.Println(convention.Convert("UserID"))
		fmt.Println(convention.Convert("HomeAddress"))
		fmt.Println(convention.Convert("StreetName"))
		fmt.Println(convention.Join([]string{
			convention.Convert("HomeAddress"),
			convention.Convert("StreetName"),
		}))
		fmt.Println(convention.Convert("HTMLElement"))
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
	//
	// lower convention:
	//
	// userid
	// homeaddress
	// streetname
	// homeaddressstreetname
	// htmlelement
}

func TestSnakeJoin(t *testing.T) {
	tests := []struct {
		names    []string
		expected string
	}{
		{
			names:    []string{"name"},
			expected: "name",
		},
		{
			names:    []string{"name1", "name2"},
			expected: "name1_name2",
		},
	}

	for _, tt := range tests {
		actual := Snake.Join(tt.names)
		if actual != tt.expected {
			t.Errorf("expected=%q, actual=%q", tt.expected, actual)
		}
	}
}
