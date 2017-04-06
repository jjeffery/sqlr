package statement

import "bytes"

// sqlStringer produces a fragment of SQL given a dialect and a column naming convention.
type sqlStringer interface {
	SQLString(dialect Dialect, columnNamer ColumnNamer) string
}

// sqlStringerBuf is used to build an sqlStringer that in turns
// produces an SQL query given a dialect and naming convention.
type sqlStringerBuf struct {
	fragments  []sqlStringer
	literalBuf bytes.Buffer
	err        error
}

func (b *sqlStringerBuf) WriteRune(r rune) {
	if b.err != nil {
		return
	}
	if _, err := b.literalBuf.WriteRune(r); err != nil {
		b.err = err
	}
}

func (b *sqlStringerBuf) WriteString(s string) {
	if b.err != nil {
		return
	}
	if _, err := b.literalBuf.WriteString(s); err != nil {
		b.err = err
	}
}

func (b *sqlStringerBuf) flush() error {
	if b.err != nil {
		return b.err
	}
	if b.literalBuf.Len() > 0 {
		b.fragments = append(b.fragments, sqlLiteralFrag(b.literalBuf.String()))
		b.literalBuf.Reset()
	}
	return nil
}

func (b *sqlStringerBuf) WritePlaceholder(position int) {
	if b.flush() != nil {
		return
	}
	b.fragments = append(b.fragments, sqlPlaceholderFrag(position))
}

func (b *sqlStringerBuf) WriteColumns(cols columnsT) {
	if b.flush() != nil {
		return
	}
	b.fragments = append(b.fragments, cols)
}

func (b *sqlStringerBuf) WriteQuoted(lit string) {
	if b.flush() != nil {
		return
	}
	b.fragments = append(b.fragments, sqlQuoteFrag(lit))
}

func (b *sqlStringerBuf) Finish() (sqlStringer, error) {
	if err := b.flush(); err != nil {
		return nil, err
	}
	s := sqlFrags(b.fragments)
	b.fragments = nil
	return s, nil
}

// sqlLiteralFrag holds an SQL fragment containing literal text.
type sqlLiteralFrag string

func (f sqlLiteralFrag) SQLString(Dialect, ColumnNamer) string {
	return string(f)
}

// sqlPlaceholderFrag holds an SQL fragment containing a placeholder.
type sqlPlaceholderFrag int

func (f sqlPlaceholderFrag) SQLString(dialect Dialect, columnNamer ColumnNamer) string {
	return dialect.Placeholder(int(f))
}

// sqlQuoteFrag holds an SQL fragment containing a quoted identifier.
type sqlQuoteFrag string

func (f sqlQuoteFrag) SQLString(dialect Dialect, columnNamer ColumnNamer) string {
	return dialect.Quote(string(f))
}

// sqlFrags is an SQL stringer consisting of a list of SQL stringers to be
// appended together.
type sqlFrags []sqlStringer

func (f sqlFrags) SQLString(dialect Dialect, columnNamer ColumnNamer) string {
	var buf bytes.Buffer
	for _, frag := range f {
		buf.WriteString(frag.SQLString(dialect, columnNamer))
	}
	return buf.String()
}
