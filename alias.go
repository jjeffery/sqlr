package sqlf

import (
	"fmt"
	"strconv"
	"sync/atomic"
)

// newTableAlias generates a unique table name alias.
// The alias will always start with an underscore ("_").
var newTableAlias func() string

func init() {
	var lastID uint64
	newTableAlias = func() string {
		id := atomic.AddUint64(&lastID, 1)
		return fmt.Sprintf("_t%s", strconv.FormatUint(id, 36))
	}
}

// newColumnAlias generates a unique column name alias.
// the alias will always start with an underscore ("_")
var newColumnAlias func() string

func init() {
	var lastID uint64
	newColumnAlias = func() string {
		id := atomic.AddUint64(&lastID, 1)
		return fmt.Sprintf("_c%s", strconv.FormatUint(id, 36))
	}
}
