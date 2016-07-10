package sqlf_test

import (
	"fmt"
	"time"

	"github.com/jjeffery/sqlf"

	// loading a driver will causes the default
	// dialect to be consistent with that driver (mysql)
	_ "github.com/go-sql-driver/mysql"
)

func ExampleColumns() {
	type Row struct {
		ID        int64 `db:",primary key autoincr"`
		Name      string
		UpdatedAt time.Time
	}

	table := sqlf.Table("examples", Row{})

	fmt.Println()
	fmt.Println("table.Select.Columns:", table.Select.Columns)
	fmt.Println("table.Select.OrderBy:", table.Select.OrderBy)

	fmt.Println()
	fmt.Println("table.Insert.Columns:", table.Insert.Columns)
	fmt.Println("table.Insert.Values:", table.Insert.Values)

	fmt.Println()
	fmt.Println("table.Update.SetColumns:", table.Update.SetColumns)
	fmt.Println("table.Update.WhereColumns:", table.Update.WhereColumns)

	// Output:
	//
	// table.Select.Columns: `id`,`name`,`updated_at`
	// table.Select.OrderBy: `id`
	//
	// table.Insert.Columns: `name`,`updated_at`
	// table.Insert.Values: ?,?
	//
	// table.Update.SetColumns: `name`=?,`updated_at`=?
	// table.Update.WhereColumns: `id`=?
}

func ExampleColumns_PK() {
	type Row struct {
		ID        int64 `db:",primary key autoincr"`
		Version   int64 `db:",version"`
		Name      string
		UpdatedAt time.Time
	}

	table := sqlf.Table("examples", Row{})

	fmt.Println()
	fmt.Println("table.Update.WhereColumns:", table.Update.WhereColumns)
	fmt.Println("table.Update.WhereColumns.PK():", table.Update.WhereColumns.PK())
	fmt.Println("table.Update.WhereColumns.PKV():", table.Update.WhereColumns.PKV())
	fmt.Println()
	fmt.Println("table.Select.Columns:", table.Select.Columns)
	fmt.Println("table.Select.Columns.PK():", table.Select.Columns.PK())
	fmt.Println("table.Select.Columns.PKV():", table.Select.Columns.PKV())

	// Output:
	//
	// table.Update.WhereColumns: `id`=? and `version`=?
	// table.Update.WhereColumns.PK(): `id`=?
	// table.Update.WhereColumns.PKV(): `id`=? and `version`=?
	//
	// table.Select.Columns: `id`,`version`,`name`,`updated_at`
	// table.Select.Columns.PK(): `id`
	// table.Select.Columns.PKV(): `id`,`version`
}

func ExampleColumns_Alias() {
	type Row struct {
		ID        int64 `db:",primary key autoincr"`
		Name      string
		UpdatedAt time.Time
	}

	table := sqlf.Table("examples", Row{})

	fmt.Println("table.Select.Columns:", table.Select.Columns)
	fmt.Println(`table.Select.Columns.Alias("a"):`, table.Select.Columns.Alias("a"))

	// Output:
	// table.Select.Columns: `id`,`name`,`updated_at`
	// table.Select.Columns.Alias("a"): `a`.`id`,`a`.`name`,`a`.`updated_at`
}
