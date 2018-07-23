package dialect

import (
	"database/sql/driver"
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		dialect             *Dialect
		expectedQuoted      string
		expectedPlaceholder string
	}{
		{
			dialect:             MySQL,
			expectedQuoted:      "`xxx`",
			expectedPlaceholder: "?",
		},
		{
			dialect:             &Postgres.Dialect,
			expectedQuoted:      `"xxx"`,
			expectedPlaceholder: "$2",
		},
		{
			dialect:             SQLite,
			expectedQuoted:      "`xxx`",
			expectedPlaceholder: "?",
		},
		{
			dialect:             MSSQL,
			expectedQuoted:      "[xxx]",
			expectedPlaceholder: "?",
		},
		{
			dialect:             ANSI,
			expectedQuoted:      `"xxx"`,
			expectedPlaceholder: "?",
		},
	}

	for _, tt := range tests {
		d := tt.dialect
		compareString(t, tt.expectedQuoted, d.Quote("xxx"))
		compareString(t, tt.expectedPlaceholder, d.Placeholder(2))
	}
}

func compareString(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Fatalf("expected=%q, actual=%q", expected, actual)
	}
}

type testDriver1 struct{}

type testDriver2 struct{}

func (d *testDriver1) Open(name string) (driver.Conn, error) {
	return nil, errors.New("not implemented")
}

func (d *testDriver2) Open(name string) (driver.Conn, error) {
	return nil, errors.New("not implemented")
}

func TestMatch(t *testing.T) {
	tests := []struct {
		dialect *Dialect
		driver  driver.Driver
		match   bool
	}{
		{
			dialect: &Dialect{
				driverTypes: []string{"*dialect.testDriver1"},
			},
			driver: &testDriver1{},
			match:  true,
		},
		{
			dialect: &Dialect{
				driverTypes: []string{"*dialect.testDriver1", "*dialect.testDriver2"},
			},
			driver: &testDriver2{},
			match:  true,
		},
		{
			dialect: &Dialect{
				driverTypes: []string{"*dialect.testDriver1"},
			},
			driver: &testDriver2{},
			match:  false,
		},
	}

	for i, tt := range tests {
		if got, want := tt.dialect.Match(tt.driver), tt.match; got != want {
			t.Errorf("%d: want=%v, got=%v", i, want, got)
		}
	}
}
