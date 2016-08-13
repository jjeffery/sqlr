package sqlrow

import (
	"fmt"
	"regexp"
	"strings"
)

// statement formats when only a table name is given
const (
	insertFormat = "insert into %s({}) values({})"
	updateFormat = "update %s set {} where {}"
	deleteFormat = "delete from %s where {}"
	selectFormat = "select{} from %s where {}"
)

var whiteSpaceRE = regexp.MustCompile(`\s`)

// checkSQL inspects the contents of sql, and if it contains a table
// name (ie has not whitespace), then it returns SQL formatted with the
// table name.
func checkSQL(sql string, format string) string {
	sql = strings.TrimSpace(sql)
	if !whiteSpaceRE.MatchString(sql) {
		sql = fmt.Sprintf(format, sql)
	}
	return sql
}
