package sqlf

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func createDatabase(t *testing.T, option string) *sqlx.DB {
	url := ":memory:" + option

	db, err := sqlx.Open("sqlite3", url)
	if err != nil {
		t.Fatal(err)
	}

	for _, cmd := range create {
		_, err := db.Exec(cmd)
		if err != nil {
			t.Errorf("SQL failed: %v\n%s", err, cmd)
		}
	}

	return db
}

var create = []string{`
create table users(
  id integer primary key autoincrement,
  family_name text,
  given_name text
)`,
}

type User struct {
	ID         int `sql:"primary_key;auto_increment"`
	GivenName  string
	FamilyName string
}

func Test1(t *testing.T) {
	assert := assert.New(t)
	db := createDatabase(t, "")
	assert.NotNil(db)
	var err error

	tbl := Table("users", User{})
	user := User{
		FamilyName: "Citizen",
		GivenName:  "John",
	}

	ins := InsertRowf("insert into %s(%s) values(%s)", tbl.Insert.TableName, tbl.Insert.Columns, tbl.Insert.Values)
	t.Log(ins.Command())
	err = ins.Exec(db, &user)
	assert.NoError(err)

	sel := Queryf("select %s from %s order by %s", tbl.Select.Columns, tbl.Select.TableName, tbl.Select.OrderBy)
	t.Log(sel.Command())
	row := sel.QueryRow(db)

	var u User
	err = row.StructScan(&u)
	assert.NoError(err)
	assert.Equal(1, u.ID)

	ins2 := InsertRowf("insert into %s(%s) values(%s)", tbl.Insert.TableName, tbl.Insert.Columns.All(), tbl.Insert.Values.All())
	user.ID = 3
	err = ins2.Exec(db, user) // note: not a pointer
	assert.NoError(err)

	var users []User
	err = sel.Select(db, &users)
	assert.NoError(err)
	assert.Equal(2, len(users))
	assert.Equal(3, users[1].ID)

	upd1 := UpdateRowf("update %s set %s where %s", tbl.Update.TableName, tbl.Update.SetColumns, tbl.Update.WhereColumns)
	t.Log(upd1.Command())
	user.GivenName = "Jane"
	user.FamilyName = "Doe"
	numRows, err := upd1.Exec(db, user)
	assert.NoError(err)
	assert.Equal(1, numRows)

	users = nil
	err = sel.Select(db, &users)
	assert.NoError(err)
	assert.Equal(2, len(users))
	assert.Equal(1, users[0].ID)
	assert.Equal("Citizen", users[0].FamilyName)
	assert.Equal("John", users[0].GivenName)
	assert.Equal(3, users[1].ID)
	assert.Equal("Doe", users[1].FamilyName)
	assert.Equal("Jane", users[1].GivenName)
}
