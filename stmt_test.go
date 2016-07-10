package sqlf

func (stmt *InsertRowStmt) Query() string {
	return stmt.query
}

func (stmt *UpdateRowStmt) Query() string {
	return stmt.query
}
