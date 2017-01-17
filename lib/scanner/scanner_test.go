package scanner_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jncornett/burp/lib/scanner"
)

func TestScannerScan(t *testing.T) {
	tests := []struct {
		Program string
		Tag     scanner.Tag
		Value   string
	}{
		{"", scanner.EOF, ""},
		{" ", scanner.WS, " "},
		{"  x", scanner.WS, "  "},
		{"foobar", scanner.CHUNK, "foobar"},
		{"foobar  ", scanner.CHUNK, "foobar"},
		{"foobar|", scanner.CHUNK, "foobar"},
		{">", scanner.REDIROUT, ">"},
		{"^", scanner.REDIRERR, "^"},
		{";", scanner.BREAK, ";"},
		{"|", scanner.PIPE, "|"},
		{"|x", scanner.PIPE, "|"},
		{"&", scanner.BACKGRND, "&"},
		{"&x", scanner.BACKGRND, "&"},
		{"(", scanner.STARTGRP, "("},
		{")", scanner.ENDGRP, ")"},
		{"!", scanner.NOT, "!"},
		{"||", scanner.OR, "||"},
		{"&&", scanner.AND, "&&"},

		// escaping
		{"\\ ", scanner.CHUNK, " "},
		{"\\foobar", scanner.CHUNK, "foobar"},
		{"\\>", scanner.CHUNK, ">"},
		{"\\^", scanner.CHUNK, "^"},
		{"\\|", scanner.CHUNK, "|"},
		{"\\(", scanner.CHUNK, "("},
		{"\\)", scanner.CHUNK, ")"},
		{"\\[", scanner.CHUNK, "["},
		{"\\]", scanner.CHUNK, "]"},
		{"\\&", scanner.CHUNK, "&"},
		{"\\;", scanner.CHUNK, ";"},
		{"\\!", scanner.CHUNK, "!"},
		{"\\\"", scanner.CHUNK, "\""},

		// quoting
		{"\"foo bar\"", scanner.CHUNK, "foo bar"},
		{"\"foo\\\"bar\"", scanner.CHUNK, "foo\"bar"},
		{"\"\\foobar\"", scanner.CHUNK, "foobar"},

		// exe
		{"[foo bar]", scanner.EXE, "foo bar"},
		{"[foo\\]bar]", scanner.EXE, "foo]bar"},
		{"[\\foobar]", scanner.EXE, "foobar"},

		// var
		{"{foo bar}", scanner.VAR, "foo bar"},
		{"{foo\\}bar}", scanner.VAR, "foo}bar"},
		{"{\\foobar}", scanner.VAR, "foobar"},
	}
	for _, test := range tests {
		t.Run(test.Program, func(t *testing.T) {
			s := scanner.NewScanner(strings.NewReader(test.Program))
			tok := s.Scan()
			if test.Tag != tok.Tag {
				t.Errorf("expected type %v, got %v", test.Tag, tok.Tag)
			}
			if test.Value != tok.Value {
				// t.Logf("expected: %v", []byte(test.Value))
				// t.Logf("actual: %v", []byte(test.Value))
				t.Errorf("expected value %q, got %q", test.Value, tok.Value)
			}
		})
	}
}

func TestScannerScanRepeated(t *testing.T) {
	program := "foo &&bar||({this.var}|\"that stuff\"[ok]);echo"
	tokens := []scanner.Token{
		{Tag: scanner.CHUNK, Value: "foo", Start: 0},
		{Tag: scanner.WS, Value: " ", Start: 3},
		{Tag: scanner.AND, Value: "&&", Start: 4},
		{Tag: scanner.CHUNK, Value: "bar", Start: 6},
		{Tag: scanner.OR, Value: "||", Start: 9},
		{Tag: scanner.STARTGRP, Value: "(", Start: 11},
		{Tag: scanner.VAR, Value: "this.var", Start: 12},
		{Tag: scanner.PIPE, Value: "|", Start: 22},
		{Tag: scanner.CHUNK, Value: "that stuff", Start: 23},
		{Tag: scanner.EXE, Value: "ok", Start: 35},
		{Tag: scanner.ENDGRP, Value: ")", Start: 39},
		{Tag: scanner.BREAK, Value: ";", Start: 40},
		{Tag: scanner.CHUNK, Value: "echo", Start: 41},
		{Tag: scanner.EOF, Value: "", Start: 45},
		{Tag: scanner.EOF, Value: "", Start: 45},
	}
	s := scanner.NewScanner(strings.NewReader(program))
	for i, tok := range tokens {
		got := s.Scan()
		if !reflect.DeepEqual(tok, got) {
			t.Errorf("expected token %v to be %v, got %v", i, tok, got)
		}
	}
}
