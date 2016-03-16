package sqlf_test

import (
	"testing"
	"time"

	"github.com/spexp/sqlf"
	"github.com/stretchr/testify/assert"
)

type Row1 struct {
	Id         int64 `gorm:"AUTO_INCREMENT"`
	GivenName  string
	FamilyName string
	IgnoreMe   bool      `sql:"-"`
	DOB        time.Time `gorm:"COLUMN:Date_of_Birth"`
}

type Row2 struct {
	UserId     int    `gorm:"PRIMARY_KEY"`
	SearchTerm string `gorm:"PRIMARY_KEY"`
}

var Row1Table = sqlf.Table("table1", Row1{})
var Row2Table = sqlf.Table("table2", Row2{})

var insertRow1 func(db sqlf.Execer, row1 *Row1) error
var updateRow1 func(db sqlf.Execer, row1 *Row1) error

func init() {
	sqlf.DefaultDialect = sqlf.DialectMySQL
	insert := Row1Table.Insert
	cmd := sqlf.InsertRowf("insert into %s(%s) values(%s)", insert.TableName, insert.Columns, insert.Values)
	insertRow1 = func(db sqlf.Execer, row1 *Row1) error {
		err := cmd.Exec(db, row1)
		if err != nil {
			return err
		}
		return nil
	}
}
func init() {
	update := Row1Table.Update
	cmd := sqlf.UpdateRowf("update %s set %s where %s", update.TableName, update.SetColumns.Updateable(), update.WhereColumns.PrimaryKey())
	updateRow1 = func(db sqlf.Execer, row1 *Row1) error {
		_, err := cmd.Exec(db, row1)
		if err != nil {
			return err
		}
		return nil
	}
}

// This test is just for experimenting with the API. Will produce some
// more thorough tests once the API firms up a bit.
func TestExample(t *testing.T) {
	assert := assert.New(t)
	row1 := Row1Table
	//row2 := Row2Table
	assert.Equal("`id`,`given_name`,`family_name`,`Date_of_Birth`", row1.Select.Columns.String())
	assert.Equal("`given_name`=?,`family_name`=?,`Date_of_Birth`=?", row1.Update.SetColumns.String())
	assert.Equal("\"given_name\"=$0,\"family_name\"=$0,\"Date_of_Birth\"=$0", row1.WithDialect(sqlf.DialectPG).Update.SetColumns.String())
	assert.Equal("[given_name]=?,[family_name]=?,[Date_of_Birth]=?", row1.WithDialect(sqlf.DialectMSSQL).Update.SetColumns.String())
	assert.Equal("`id`=?", row1.Update.WhereColumns.String())
	assert.Equal("`given_name`,`family_name`,`Date_of_Birth`", row1.Insert.Columns.String())
	assert.Equal("?,?,?", row1.Insert.Values.Insertable().String())

	insertCmd := sqlf.InsertRowf("insert into %s(%s) values (%s)", row1.Insert.TableName, row1.Insert.Columns, row1.Insert.Values)
	assert.NotNil(insertCmd)
	assert.Equal("insert into `table1`(`given_name`,`family_name`,`Date_of_Birth`) values (?,?,?)", insertCmd.Command())
	insertArgs, err := insertCmd.Args(Row1{
		FamilyName: "Smith",
		GivenName:  "John",
		DOB:        time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.NoError(err)
	assert.Equal(3, len(insertArgs))
	assert.Equal("John", insertArgs[0])
	assert.Equal("Smith", insertArgs[1])
	assert.Equal(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), insertArgs[2])

	tblpg := row1.WithDialect(sqlf.DialectPG)
	insertCmd = sqlf.InsertRowf("insert into %s(%s) values (%s)", tblpg.Insert.TableName, tblpg.Insert.Columns, tblpg.Insert.Values)
	assert.NotNil(insertCmd)
	assert.Equal(`insert into "table1"("given_name","family_name","Date_of_Birth") values ($1,$2,$3)`, insertCmd.Command())

	updateCmd := sqlf.UpdateRowf("update %s set %s where %s", row1.Update.TableName, row1.Update.SetColumns, row1.Update.WhereColumns)
	assert.NotNil(updateCmd)
	assert.Equal("update `table1` set `given_name`=?,`family_name`=?,`Date_of_Birth`=? where `id`=?", updateCmd.Command())
	updateArgs, err := updateCmd.Args(Row1{
		Id:         244,
		FamilyName: "Citizen",
		GivenName:  "Jane",
		DOB:        time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.Equal(4, len(updateArgs))
	assert.Equal("Jane", updateArgs[0])
	assert.Equal("Citizen", updateArgs[1])
	assert.Equal(time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC), updateArgs[2])
	assert.Equal(int64(244), updateArgs[3])

	/**
	selectStmt := sqlf.Selectf("select %s from %s where %s = ?", row1.Select.Columns, row1.Select.TableName, row1.Select.Where("FamilyName"))

	selectCmd := sqlf.Selectf("select %s, %s from %s inner join %s on %s = %s where %s like ? and %s > 10 order by %s limit %s offset %s",
		row1.Select.Columns,
		row2.Select.Columns,
		row1.Select.TableName,
		row2.Select.TableName,
		row1.Select.Join("Id"),
		row2.Select.Join("UserId"),
		row2.Select.Where("SearchTerm"),
		row1.Select.WhereColumn("Id"),
		row1.Select.Order("FamilyName", "GivenName"),
		row1.Select.Limit(),
		row1.Select.Offset(),
	)

	// not implemented yet
	assert.Nil(selectStmt)
	**/
}
