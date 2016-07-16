package dialect

import (
	"database/sql"
	"database/sql/driver"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		driverName          string
		dataSource          string
		expectedName        string
		expectedQuoted      string
		expectedPlaceholder string
	}{
		{
			driverName:          "mysql",
			dataSource:          "/dbname",
			expectedName:        "mysql",
			expectedQuoted:      "`xxx`",
			expectedPlaceholder: "?",
		},
		{
			driverName:          "postgres",
			dataSource:          "user=test dbname=test sslmode=none",
			expectedName:        "postgres",
			expectedQuoted:      `"xxx"`,
			expectedPlaceholder: "$2",
		},
		{
			driverName:          "sqlite3",
			dataSource:          ":memory:",
			expectedName:        "sqlite",
			expectedQuoted:      "`xxx`",
			expectedPlaceholder: "?",
		},
		{
			driverName:          "mssql",
			dataSource:          "server=.;user id=dbo;password=whatever",
			expectedName:        "mssql",
			expectedQuoted:      "[xxx]",
			expectedPlaceholder: "?",
		},
		{
			driverName:          "ql-mem",
			dataSource:          "memory",
			expectedName:        "ql",
			expectedQuoted:      "xxx",
			expectedPlaceholder: "?2",
		},
		{
			driverName:          "ql-mem",
			dataSource:          "database.dat",
			expectedName:        "ql",
			expectedQuoted:      "xxx",
			expectedPlaceholder: "?2",
		},
		{
			driverName:          "whatever",
			dataSource:          "server=.;user id=dbo;password=whatever",
			expectedName:        "default",
			expectedQuoted:      `"xxx"`,
			expectedPlaceholder: "?",
		},
	}

	for _, tt := range tests {
		d := New(tt.driverName)
		compareString(t, tt.expectedName, d.Name())
		compareString(t, tt.expectedQuoted, d.Quote("xxx"))
		compareString(t, tt.expectedPlaceholder, d.Placeholder(2))
	}
}

func compareString(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Fatalf("expected=%q, actual=%q", expected, actual)
	}
}

type testDriver struct{}

func (d *testDriver) Open(name string) (driver.Conn, error) {
	return nil, nil
}

func TestInferFromDriver(t *testing.T) {
	sql.Register("mysql", &testDriver{})
	dialect := New("")

	if dialect.Name() != "mysql" {
		t.Errorf("expected=%q, actual=%q", "mysql", dialect.Name())
	}
}
