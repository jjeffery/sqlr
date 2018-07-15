package column

import (
	"testing"
)

func TestPathString(t *testing.T) {
	tests := []struct {
		path Path
		text string
	}{
		{
			path: nil,
			text: "",
		},
		{
			path: NewPath("A", ""),
			text: "A",
		},
		{
			path: NewPath("A", "").Append("B", ""),
			text: "A.B",
		},
		{
			path: NewPath("A", `sql:"a"`).Append("B", `sql:"b"`),
			text: "A.B",
		},
	}

	for _, tt := range tests {
		text := tt.path.String()
		if text != tt.text {
			t.Errorf("expected=%q, actual=%q", tt.text, text)
		}
	}
}

func TestPathEqual(t *testing.T) {
	tests := []struct {
		path  Path
		other Path
		equal bool
	}{
		{
			path:  nil,
			other: nil,
			equal: true,
		},
		{
			path:  nil,
			other: NewPath("", ""),
			equal: false,
		},
		{
			path:  NewPath("A", "a"),
			other: NewPath("A", ""),
			equal: false,
		},
	}

	for _, tt := range tests {
		equal := tt.path.Equal(tt.other)
		if equal != tt.equal {
			t.Errorf("expected=%v, actual=%v", tt.equal, equal)
		}
		equal = tt.other.Equal(tt.path)
		if equal != tt.equal {
			t.Errorf("expected=%v, actual=%v", tt.equal, equal)
		}
	}
}
