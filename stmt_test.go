package sqlf

func (stmt InsertRowStmt) Query() string {
	return stmt.query
}
