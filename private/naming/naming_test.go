package naming

import (
	"fmt"
	"testing"
)

func Example() {
	type convention interface {
		Convert(string) string
		Join([]string) string
	}
	conventions := []convention{SnakeCase, SameCase, LowerCase}
	names := []string{"snake case", "same case", "lower case"}

	for i, convention := range conventions {
		fmt.Printf("\n%s:\n\n", names[i])
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
	// snake case:
	//
	// user_id
	// home_address
	// street_name
	// home_address_street_name
	// html_element
	//
	// same case:
	//
	// UserID
	// HomeAddress
	// StreetName
	// HomeAddressStreetName
	// HTMLElement
	//
	// lower case:
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
		if want, got := tt.expected, SnakeCase.Join(tt.names); want != got {
			t.Errorf("expected=%q, actual=%q", want, got)
		}
	}
}
