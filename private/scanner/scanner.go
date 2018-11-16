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
	IDENT                    // identifer, which may be quoted
	KEYWORD                  // keyword as per AddKeywords
	LITERAL                  // string or numeric literal
	OP                       // operator
	PLACEHOLDER              // prepared statement placeholder
)

func (t Token) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case WS:
		return "WS"
	case COMMENT:
		return "COMMENT"
	case IDENT:
		return "IDENT"
	case KEYWORD:
		return "KEYWORD"
	case LITERAL:
		return "LITERAL"
	case OP:
		return "OP"
	case PLACEHOLDER:
		return "PLACEHOLDER"
	default:
		return fmt.Sprintf("Token-%d", t)
	}
}

const (
	eof                 = rune(0)
	multiCharOperators  = "%&*+-/:<=>^|@!~#"
	singleCharOperators = "(),;"
)

// Scanner is a simple lexical scanner for SQL statements.
type Scanner struct {
	IgnoreWhiteSpace bool

	r        *bufio.Reader
	keywords map[string]bool
	err      error
	token    Token
	text     string
}

// New returns a new scanner that takes its input from r.
func New(r io.Reader) *Scanner {
	return &Scanner{
		r:        bufio.NewReader(r),
		keywords: make(map[string]bool),
	}
}

// AddKeywords informs the scanner of keywords. Keywords
// are not case sensitive.
func (s *Scanner) AddKeywords(keywords ...string) {
	for _, keyword := range keywords {
		key := strings.TrimSpace(strings.ToLower(keyword))
		s.keywords[key] = true
	}
}

func (s *Scanner) isKeyword(lit string) bool {
	return s.keywords[strings.ToLower(lit)]
}

// Token returns the token from the last scan.
func (s *Scanner) Token() Token {
	return s.token
}

// Text returns the token's text from the last scan.
func (s *Scanner) Text() string {
	return s.text
}

// Err returns the first non-EOF error that was
// encountered by the Scanner.
func (s *Scanner) Err() error {
	return s.err
}

// Scan the next SQL token.
func (s *Scanner) Scan() bool {
	ch := s.read()
	for s.IgnoreWhiteSpace && isWhitespace(ch) {
		ch = s.read()
	}
	if ch == eof {
		return s.setToken(EOF, "")
	}
	if isWhitespace(ch) {
		s.unread(ch)
		return s.scanWhitespace()
	}
	if ch == '-' {
		ch2 := s.read()
		if ch2 == '-' {
			return s.scanComment("--")
		}
		s.unread(ch2)
		return s.scanOperator(ch)
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
	if ch == '{' {
		return s.scanDelimitedIdentifier('{', '}')
	}
	if strings.ContainsRune("NnXxBb", ch) {
		ch2 := s.read()
		if ch2 == '\'' {
			return s.scanQuote(ch, ch2)
		}
		s.unread(ch2)
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
		s.unread(ch2)
		if isDigit(ch2) {
			return s.scanNumber(ch)
		}
		return s.setToken(OP, runeToString(ch))
	}
	if ch == '$' || ch == '?' {
		return s.scanPlaceholder(ch)
	}
	if strings.ContainsRune(singleCharOperators, ch) {
		return s.setToken(OP, runeToString(ch))
	}
	if strings.ContainsRune(multiCharOperators, ch) {
		return s.scanOperator(ch)
	}

	return s.setToken(ILLEGAL, runeToString(ch))
}

func (s *Scanner) setToken(tok Token, text string) bool {
	s.token = tok
	s.text = text
	if tok == ILLEGAL {
		s.err = fmt.Errorf("unrecognised input near %q", text)
		return false
	}
	return tok != EOF
}

func (s *Scanner) scanWhitespace() bool {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread(ch)
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return s.setToken(WS, buf.String())
}

func (s *Scanner) scanComment(prefix string) bool {
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
	return s.setToken(COMMENT, buf.String())
}

func (s *Scanner) scanDelimitedIdentifier(startCh rune, endCh rune) bool {
	var buf bytes.Buffer
	buf.WriteRune(startCh)
	for {
		ch := s.read()
		if ch == eof {
			return s.setToken(ILLEGAL, buf.String())
		}
		buf.WriteRune(ch)
		if ch == endCh {
			// double endCh is an escape
			ch2 := s.read()
			if ch2 != endCh {
				s.unread(ch2)
				break
			}
			buf.WriteRune(ch2)
		}
	}
	return s.setToken(IDENT, buf.String())
}

func (s *Scanner) scanIdentifier(startCh rune) bool {
	var buf bytes.Buffer
	buf.WriteRune(startCh)
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isIdent(ch) {
			s.unread(ch)
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	lit := buf.String()
	if s.isKeyword(lit) {
		return s.setToken(KEYWORD, lit)
	}

	return s.setToken(IDENT, lit)
}

func (s *Scanner) scanOperator(startCh rune) bool {
	var buf bytes.Buffer
	buf.WriteRune(startCh)
	for {
		if ch := s.read(); ch == eof {
			break
		} else if strings.ContainsRune(multiCharOperators, ch) {
			buf.WriteRune(ch)
		} else {
			s.unread(ch)
			break
		}
	}
	return s.setToken(OP, buf.String())
}

func (s *Scanner) scanNumber(startCh rune) bool {
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
			s.unread(ch)
			break
		}
	}

	return s.setToken(LITERAL, buf.String())
}

func (s *Scanner) scanQuote(startChs ...rune) bool {
	var buf bytes.Buffer
	var endCh rune
	for _, ch := range startChs {
		endCh = ch
		buf.WriteRune(ch)
	}
	for {
		ch := s.read()
		if ch == eof {
			return s.setToken(ILLEGAL, buf.String())
		}
		buf.WriteRune(ch)
		if ch == endCh {
			if ch2 := s.read(); ch2 == endCh {
				buf.WriteRune(ch2)
			} else {
				s.unread(ch2)
				break
			}
		}
	}
	return s.setToken(LITERAL, buf.String())
}

func (s *Scanner) scanPlaceholder(startCh rune) bool {
	if startCh == '?' {
		// postgres has the following geometric operators
		//  which look a bit like placeholders:
		// ?- ?# ?| ?-| ?||
		ch := s.read()
		s.unread(ch)
		if ch == '-' || ch == '#' || ch == '|' {
			return s.scanOperator(startCh)
		}
	}
	var buf bytes.Buffer
	buf.WriteRune(startCh)
	for {
		if ch := s.read(); ch == eof {
			break
		} else if isDigit(ch) {
			buf.WriteRune(ch)
		} else {
			s.unread(ch)
			break
		}
	}
	return s.setToken(PLACEHOLDER, buf.String())
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		if err != io.EOF {
			s.err = err
		}
		return eof
	}
	return ch
}

func (s *Scanner) unread(ch rune) {
	if ch != eof {
		err := s.r.UnreadRune()
		if err != nil {
			s.err = err
		}
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
