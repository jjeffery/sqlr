package scanner

import (
	"strings"
)

type quotePair struct {
	start      string
	end        string
	minLen     int
	escapedEnd string
}

func newQuotePair(start, end string) quotePair {
	return quotePair{
		start:      start,
		end:        end,
		minLen:     len(start) + len(end),
		escapedEnd: end + end,
	}
}

func (qp *quotePair) isQuoted(ident string) bool {
	return len(ident) >= qp.minLen &&
		strings.HasPrefix(ident, qp.start) &&
		strings.HasSuffix(ident, qp.end)
}

func (qp *quotePair) unQuote(ident string) string {
	ident = ident[len(qp.start) : len(ident)-len(qp.end)]
	ident = strings.Replace(ident, qp.escapedEnd, qp.end, -1)
	return ident
}

var quotePairs = []quotePair{
	newQuotePair("\"", "\""),
	newQuotePair("`", "`"),
	newQuotePair("[", "]"),
	newQuotePair("'", "'"),
	newQuotePair("{", "}"),
}

// IsQuoted returns true if the identifier is a quoted identifier.
func IsQuoted(ident string) bool {
	for _, qp := range quotePairs {
		if qp.isQuoted(ident) {
			return true
		}
	}
	return false
}

// Unquote will unquote an identifier, if it is quoted.
// If the syntax of the identifier is not valid the result is
// undefined.
func Unquote(ident string) string {
	for _, qp := range quotePairs {
		if qp.isQuoted(ident) {
			return qp.unQuote(ident)
		}
	}
	return ident
}

// Quote the identifer using the start and end quote strings.
// If the end quote string occurs in ident, it is escaped.
func Quote(ident, start, end string) string {
	ident = Unquote(ident)
	escapeEnd := end + end
	ident = strings.Replace(ident, end, escapeEnd, -1)
	return start + ident + end
}
