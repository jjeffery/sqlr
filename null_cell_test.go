package sqlr

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestNullScannerCell(t *testing.T) {
	db := postgresDB(t)
	defer db.Close()

	mustExec(t, db, `drop table if exists nullables;`)
	defer func() {
		mustExec(t, db, `drop table if exists nullables`)
	}()
	mustExec(t, db, `
		create table nullables(
			id integer not null primary key,
			enum_val text null
		);
		insert into nullables(id, enum_val) values(1, 'One');
		insert into nullables(id, enum_val) values(2, null);
	`)
	schema := NewSchema(ForDB(db))
	ctx := context.Background()
	sess := NewSession(ctx, db, schema)
	defer sess.Close()

	type rowT struct {
		ID      int
		EnumVal TestEnum `sql:"null"`
	}

	var nullables []rowT

	n, err := sess.Select(&nullables, "select {} from nullables order by id")
	if err != nil {
		t.Fatalf("got %v, want nil", err)
	}
	if got, want := n, 2; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	{
		want := []rowT{
			{ID: 1, EnumVal: TestEnumOne},
			{ID: 2, EnumVal: TestEnumZero},
		}
		if got := nullables; !reflect.DeepEqual(want, got) {
			t.Fatalf("want=%v, got=%v", want, got)
		}
	}
}

type TestEnum int

const (
	TestEnumZero TestEnum = 0
	TestEnumOne  TestEnum = 1
)

func (ne *TestEnum) Scan(src interface{}) error {
	str, ok := src.(string)
	if !ok {
		bytes, ok := src.([]byte)
		if !ok {
			return errors.New("invalid type for TestEnum")
		}

		str = string(bytes[:])
	}

	switch str {
	case "Zero":
		*ne = TestEnumZero
	case "One":
		*ne = TestEnumOne
	default:
		return fmt.Errorf("unknown value for TestEnum: %q", str)
	}
	return nil
}
