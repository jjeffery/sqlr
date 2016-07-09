// Package scanner implements a simple lexical scanner
// for SQL statements.
package scanner

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// Token is a lexical token for SQL.
type Token int

// Tokens
const (
	ILLEGAL     Token = iota // unexpected character
	EOF                      // End of input
	WS                       // White space
	COMMENT                  // SQL comment
	IDENT                    // identifer, including keywords such as "SELECT"
	LITERAL                  // string or numeric literal
	OP                       // operator
	PLACEHOLDER              // prepared statement placeholder
)

const (
	eof       = rune(0)
	operators = "%&()*+,-./:;<=>?^|{}"
)

// Scanner is a simple lexical scanner for SQL statements.
type Scanner struct {
	r   *bufio.Reader
	err error
}

// New returns a new scanner that takes its input from r.
func New(r io.Reader) *Scanner {
	return &Scanner{
		r: bufio.NewReader(r),
	}
}

// Scan the next SQL token.
func (s *Scanner) Scan() (tok Token, lit string) {
	ch := s.read()
	if ch == eof {
		return EOF, ""
	}

	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	}
	if ch == '-' {
		ch2 := s.read()
		if ch2 == '-' {
			return s.scanComment("--")
		}
		s.unread()
		return OP, runeToString(ch)
	}
	if ch == '[' {
		return s.scanDelimitedIdentifier('[', ']')
	}
	if ch == '`' {
		return s.scanDelimitedIdentifier('`', '`')
	}
	if ch == '"' {
		return s.scanDelimitedIdentifier('"', '"')
	}
	if ch == '\'' {
		return s.scanQuote(ch)
	}
	if strings.ContainsRune("NnXx", ch) {
		ch2 := s.read()
		if ch2 == '\'' {
			return s.scanQuote(ch, ch2)
		}
		s.unread()
		return s.scanIdentifier(ch)
	}
	if isStartIdent(ch) {
		return s.scanIdentifier(ch)
	}
	if isDigit(ch) {
		return s.scanNumber(ch)
	}
	if ch == '.' {
		ch2 := s.read()
		s.unread()
		if isDigit(ch2) {
			return s.scanNumber(ch)
		}
		return OP, runeToString(ch)
	}
	if ch == '<' {
		if ch2 := s.read(); ch2 == '>' {
			return OP, "<>"
		}
		s.unread()
		return OP, runeToString(ch)
	}
	if ch == '$' {
		ch2 := s.read()
		s.unread()
		if isDigit(ch2) {
			return s.scanPlaceholder(ch)
		}
		return OP, runeToString(ch)
	}
	if ch == '?' {
		return s.scanPlaceholder(ch)
	}
	if strings.ContainsRune(operators, ch) {
		return OP, runeToString(ch)
	}

	return ILLEGAL, runeToString(ch)
}

func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return WS, buf.String()
}

func (s *Scanner) scanComment(prefix string) (Token, string) {
	var buf bytes.Buffer
	buf.WriteString(prefix)
	for {
		if ch := s.read(); ch == eof {
			break
		} else {
			buf.WriteRune(ch)
			if ch == '\n' {
				break
			}
		}
	}
	return COMMENT, buf.String()
}

func (s *Scanner) scanDelimitedIdentifier(startCh rune, endCh rune) (Token, string) {
	var buf bytes.Buffer
	buf.WriteRune(startCh)
	for {
		if ch := s.read(); ch == eof {
			return ILLEGAL, buf.String()
		} else {
			buf.WriteRune(ch)
			if ch == endCh {
				// double endCh is an escape
				ch2 := s.read()
				if ch2 != endCh {
					s.unread()
					break
				}
				buf.WriteRune(ch2)
			}
		}
	}
	return IDENT, buf.String()
}

func (s *Scanner) scanIdentifier(startCh rune) (Token, string) {
	var buf bytes.Buffer
	buf.WriteRune(startCh)
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isIdent(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return IDENT, buf.String()
}

func (s *Scanner) scanNumber(startCh rune) (Token, string) {
	var buf bytes.Buffer

	// comparison function changes after first period encountered
	var cmp = func(ch rune) bool {
		return isDigit(ch) || ch == '.'
	}

	// add to buffer and change comparison function if period encountered
	var add = func(ch rune) {
		buf.WriteRune(ch)
		if ch == '.' {
			cmp = isDigit
		}
	}

	add(startCh)
	for {
		if ch := s.read(); ch == eof {
			break
		} else if cmp(ch) {
			add(ch)
		} else {
			s.unread()
			break
		}
	}

	return LITERAL, buf.String()
}

func (s *Scanner) scanQuote(startChs ...rune) (Token, string) {
	var buf bytes.Buffer
	var endCh rune
	for _, ch := range startChs {
		endCh = ch
		buf.WriteRune(ch)
	}
	for {
		if ch := s.read(); ch == eof {
			return ILLEGAL, buf.String()
		} else {
			buf.WriteRune(ch)
			if ch == endCh {
				if ch2 := s.read(); ch2 == endCh {
					buf.WriteRune(ch2)
				} else {
					s.unread()
					break
				}
			}
		}
	}
	return LITERAL, buf.String()
}

func (s *Scanner) scanPlaceholder(startCh rune) (Token, string) {
	var buf bytes.Buffer
	buf.WriteRune(startCh)
	for {
		if ch := s.read(); ch == eof {
			break
		} else if isDigit(ch) {
			buf.WriteRune(ch)
		} else {
			s.unread()
			break
		}
	}
	return PLACEHOLDER, buf.String()
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		s.err = err
		return eof
	}
	return ch
}

func (s *Scanner) unread() {
	err := s.r.UnreadRune()
	if err != nil {
		s.err = err
	}
}

func isWhitespace(ch rune) bool {
	return unicode.IsSpace(ch)
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

func isStartIdent(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch)
}

func isIdent(ch rune) bool {
	return isStartIdent(ch) || unicode.IsDigit(ch)
}

func runeToString(ch rune) string {
	return fmt.Sprintf("%c", ch)
}
