package rowtype

import "time"

type Row3 struct {
	ID   string
	Name string
	DOB  time.Time
}

type Row4ID int

type Row4 struct {
	ID   Row4ID `sql:"primary key"`
	Name string
}
