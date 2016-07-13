/*
Package sqlf provides assistance for writing and executing SQL statements.

It is intended for programmers who are comfortable with
writing SQL, but would like assistance with the sometimes tedious
process of preparing SELECT, INSERT, UPDATE and DELETE statements
for tables that have a large number of columns.

Select all columns from a table by primary key
 sqlf query: select {columns} from users where {where}
 mysql:      select `id`,`version`,`login`,`hash_pwd`,`full_name` from users where `id`=?
 postgres:   select "id","version","login","hash_pwd","full_name" from users where "id"=$1

Select all columns from a table by "login" column
 sqlf query: select {columns} from users where {where login}
 mysql:      select `id`,`version`,`login`,`hash_pwd`,`full_name` from users where `login`=?
 postgres:   select "id","version","login","hash_pwd","full_name" from users where "login"=$1


Insert a row into table (where column id is an auto-increment column)
 sqlf.PrepareInsertRow(User{}, "insert into users({columns}) values ({values})")
 mysql:     insert into users(`version`,`login`,`hash_pwd`,`full_name`) values(?,?,?,?)
 postgres:  insert into users("version","login","hash_pwd","full_name") values($1,$2,$3,$4)

*/
package sqlf
