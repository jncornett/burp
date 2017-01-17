package scanner

import (
	"bufio"
	"bytes"
	"io"
)

// Tag denotes the lexical type of a scanned Token
type Tag int

// The list of tags
const (
	EOF Tag = iota
	WS
	CHUNK
	REDIROUT
	REDIRERR
	PIPE
	BACKGRND
	STARTGRP
	ENDGRP
	EXE
	VAR
	AND
	OR
	NOT
	BREAK
)

const eof = rune(0)

// Token is the unit token produced by a single call to Scanner.Scan
type Token struct {
	Tag
	Value string
	Start int
}

// Scanner is a stateful scanner
type Scanner struct {
	r   *bufio.Reader
	pos int
}

// NewScanner creates a new scanner backed by r
func NewScanner(r io.Reader) Scanner {
	return Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) Scan() Token {
	start := s.pos // save this value
	ch := s.read()
	switch {
	case ch == eof:
		return Token{Tag: EOF, Value: "", Start: start}
	case ch == '(':
		return Token{Tag: STARTGRP, Value: "(", Start: start}
	case ch == ')':
		return Token{Tag: ENDGRP, Value: ")", Start: start}
	case ch == '^':
		return Token{Tag: REDIRERR, Value: "^", Start: start}
	case ch == '>':
		return Token{Tag: REDIROUT, Value: ">", Start: start}
	case ch == ';':
		return Token{Tag: BREAK, Value: ";", Start: start}
	case ch == '!':
		return Token{Tag: NOT, Value: "!", Start: start}
	case ch == '|':
		if s.read() == '|' {
			return Token{Tag: OR, Value: "||", Start: start}
		}
		s.unread()
		return Token{Tag: PIPE, Value: "|", Start: start}
	case ch == '&':
		if s.read() == '&' {
			return Token{Tag: AND, Value: "&&", Start: start}
		}
		s.unread()
		return Token{Tag: BACKGRND, Value: "&", Start: start}
	case ch == '"':
		return Token{Tag: CHUNK, Value: s.scanQuoted('"'), Start: start}
	case ch == '{':
		return Token{Tag: VAR, Value: s.scanQuoted('}'), Start: start}
	case ch == '[':
		return Token{Tag: EXE, Value: s.scanQuoted(']'), Start: start}
	case isWs(ch):
		s.unread()
		return Token{Tag: WS, Value: s.scanWs(), Start: start}
	default:
		s.unread()
		return Token{Tag: CHUNK, Value: s.scanChunk(), Start: start}
	}
}

func (s *Scanner) scanQuoted(endQuote rune) string {
	return s.scanWhile(true, func(ch rune) bool { return ch != endQuote })
}

func (s *Scanner) scanWs() string {
	return s.scanWhile(false, isWs)
}

func (s *Scanner) scanChunk() string {
	return s.scanWhile(false, isChunk)
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
	} else {
		s.pos++
	}
	return ch
}

func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
	s.pos--
}

func isWs(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isSpecial(ch rune) bool {
	return ch == '^' || ch == '>' || ch == '&' || ch == '|' || ch == '(' ||
		ch == ')' || ch == ';' || ch == '"' || ch == '[' || ch == ']' ||
		ch == '{' || ch == '}'
}

func isChunk(ch rune) bool {
	return !(isWs(ch) || isSpecial(ch))
}
