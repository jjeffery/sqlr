package sqlstmt

import (
	"testing"
)

func TestDialog(t *testing.T) {
	tests := []struct {
		name        string
		quoted      string
		placeholder string
	}{
		{
			name:        "mysql",
			quoted:      "`quoted`",
			placeholder: "?",
		},
		{
			name:        "postgres",
			quoted:      `"quoted"`,
			placeholder: "$1",
		},
	}

	for _, tt := range tests {
		dialog := NewDialect(tt.name)
		quoted := dialog.Quote("quoted")
		placeholder := dialog.Placeholder(1)
		if quoted != tt.quoted {
			t.Errorf("expected=%q, actual=%q", tt.quoted, quoted)
		}
		if placeholder != tt.placeholder {
			t.Errorf("expected=%q, actual=%q", tt.placeholder, placeholder)
		}
	}
}
