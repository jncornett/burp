package scanner

import (
	"bufio"
	"bytes"
	"io"
)

type Tag int

const (
	ILLEGAL Tag = iota
	EOF
	WS
	IDENTITY
	TEXT
	VAR
	EXEC
	ESCAPE
	STARTEXEC
	ENDEXEC
	STARTGROUP
	ENDGROUP
	REDIRERR
	REDIROUT
	PIPE
	BREAK
	AND
	OR
)

const eof = rune(0)

type Token struct {
	Tag
	Value string
	Start int
}

type Scanner struct {
	r   *bufio.Reader
	pos int
}

func NewScanner(r io.Reader) Scanner {
	return Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) Scan() Token {
	start := s.pos // save this value
	ch := s.read()
	tok := Token{Tag: ILLEGAL, Value: string(ch), Start: start}
	switch {
	case ch == eof:
		tok.Tag = EOF
		tok.Value = ""
	case ch == '{':
		tok.Tag = VAR
		tok.Value = s.scanQuoted('}')
	case ch == '"':
		tok.Tag = TEXT
		tok.Value = s.scanQuoted('"')
	case ch == '[':
		tok.Tag = STARTEXEC
	case ch == ']':
		tok.Tag = ENDEXEC
	case ch == '(':
		tok.Tag = STARTGROUP
	case ch == ')':
		tok.Tag = ENDGROUP
	case ch == '^':
		tok.Tag = REDIRERR
	case ch == '>':
		tok.Tag = REDIROUT
	case ch == ';':
		tok.Tag = BREAK
	case ch == ':':
		tok.Tag = IDENTITY
	case ch == '|':
		if s.read() == '|' {
			tok.Tag = OR
			tok.Value = "||"
		} else {
			s.unread()
			tok.Tag = PIPE
		}
	case ch == '&':
		if s.read() == '&' {
			tok.Tag = AND
			tok.Value = "&&"
		} else {
			s.unread()
			tok.Tag = ILLEGAL
		}
	case ch == '\\':
		s.unread()
		tok.Tag = TEXT
		tok.Value = s.scanWhile(false, isIdent)
	case isWs(ch):
		s.unread()
		tok.Tag = WS
		tok.Value = s.scanWhile(false, isWs)
	case isIdent(ch):
		s.unread()
		tok.Tag = TEXT
		tok.Value = s.scanWhile(false, isIdent)
	}
	return tok
}

func (s *Scanner) scanQuoted(endQuote rune) string {
	return s.scanWhile(true, func(ch rune) bool { return ch != endQuote })
}

func (s *Scanner) scanWhile(consumeLast bool, accept func(rune) bool) string {
	var buf bytes.Buffer
	escape := false
	for {
		ch := s.read()
		if ch == eof {
			break
		} else if escape {
			escape = false
			buf.WriteRune(ch)
		} else if ch == '\\' {
			escape = true
		} else if !accept(ch) {
			if !consumeLast {
				s.unread()
			}
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return buf.String()
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		ch = eof
	}
	s.pos++
	return ch
}

func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
	s.pos--
}

func isWs(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isIdent(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_'
}
