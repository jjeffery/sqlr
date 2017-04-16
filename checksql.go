package sqlr

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jjeffery/sqlr/private/scanner"
)

// statement formats when only a table name is given
const (
	insertFormat = "insert into %s({}) values({})"
	updateFormat = "update %s set {} where {}"
	deleteFormat = "delete from %s where {}"
	selectFormat = "select{} from %s where {}"
)

var whiteSpaceRE = regexp.MustCompile(`\s`)

// checkSQL inspects the contents of sql, and performs the following
// replacements. The sole purpose of this is to minimise typing for
// commonly used statement patterns.
//  INSERT INTO <table> => insert into <table>({}) values ({})
//  INSERT <table>      => insert into <table>({}) values ({})
//  UPDATE <table>      => update <table> set({}) where({})
//  SELECT FROM <table>      => select {} from <table> where {}
//  SELECT <table>      => select {} from <table> where {}
// Note that we do not allow "DELETE FROM <table>" or "DELETE <table>"
// similar because that is actually valid SQL, and has the rather uncommon
// effect of deleting everything in the table.
func checkSQL(sql string) (string, error) {
	const maxWords = 3 // if the SQL has more than this number of words, leave it alone
	scan := scanner.New(strings.NewReader(sql))
	scan.IgnoreWhiteSpace = true
	scan.AddKeywords("insert", "update", "delete", "select", "into", "from")
	words := make([]string, 0, maxWords)
	for scan.Scan() {
		if len(words) >= maxWords {
			// the input is longer than the max number of words, then don't change it
			return sql, nil
		}
		if scan.Token() == scanner.KEYWORD {
			words = append(words, strings.ToLower(scan.Text()))
		} else {
			words = append(words, scan.Text())
		}
	}
	match := func(args ...string) bool {
		if len(args) != len(words) {
			return false
		}
		for i, word := range args {
			if word != "" && word != words[i] {
				return false
			}
		}
		return true
	}
	deleteTableError := func(tableName string) error {
		return fmt.Errorf("will not delete all rows in table %s: use database/sql if you want to do this", tableName)
	}

	if match("insert", "") {
		return fmt.Sprintf("insert into %s({}) values({})", words[1]), nil
	}
	if match("insert", "into", "") {
		return fmt.Sprintf("insert into %s({}) values({})", words[2]), nil
	}
	if match("update", "") {
		return fmt.Sprintf("update %s set {} where {}", words[1]), nil
	}
	if match("select", "") {
		return fmt.Sprintf("select {} from %s where {}", words[1]), nil
	}
	if match("select", "from", "") {
		return fmt.Sprintf("select {} from %s where {}", words[2]), nil
	}
	if match("delete", "") {
		return "", deleteTableError(words[1])
	}
	if match("delete", "from", "") {
		return "", deleteTableError(words[2])
	}

	return sql, nil
}
