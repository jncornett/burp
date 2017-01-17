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
		{``, scanner.EOF, ""},
		{`{foobar}`, scanner.VAR, "foobar"},
		{`"foobar"`, scanner.TEXT, "foobar"},
		{`"foo\"bar"`, scanner.TEXT, "foo\"bar"},
		{"[", scanner.STARTEXEC, "["},
		{"]", scanner.ENDEXEC, "]"},
		{"(", scanner.STARTGROUP, "("},
		{")", scanner.ENDGROUP, ")"},
		{"^", scanner.REDIRERR, "^"},
		{">", scanner.REDIROUT, ">"},
		{";", scanner.BREAK, ";"},
		{":", scanner.IDENTITY, ":"},
		{"|", scanner.PIPE, "|"},
		{"|x", scanner.PIPE, "|"},
		{"&&", scanner.AND, "&&"},
		{"&x", scanner.ILLEGAL, "&"},
		{"||", scanner.OR, "||"},
		{"    ", scanner.WS, "    "},
		{"032abczz_ZXYAB", scanner.TEXT, "032abczz_ZXYAB"},
		{"?", scanner.ILLEGAL, "?"},
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
		{Tag: scanner.TEXT, Value: "foo", Start: 0},
		{Tag: scanner.WS, Value: " ", Start: 3},
		{Tag: scanner.AND, Value: "&&", Start: 4},
		{Tag: scanner.TEXT, Value: "bar", Start: 6},
		{Tag: scanner.OR, Value: "||", Start: 9},
		{Tag: scanner.STARTGROUP, Value: "(", Start: 11},
		{Tag: scanner.VAR, Value: "this.var", Start: 12},
		{Tag: scanner.PIPE, Value: "|", Start: 22},
		{Tag: scanner.TEXT, Value: "that stuff", Start: 23},
		{Tag: scanner.STARTEXEC, Value: "[", Start: 35},
		{Tag: scanner.TEXT, Value: "ok", Start: 36},
		{Tag: scanner.ENDEXEC, Value: "]", Start: 38},
		{Tag: scanner.ENDGROUP, Value: ")", Start: 39},
		{Tag: scanner.BREAK, Value: ";", Start: 40},
		{Tag: scanner.TEXT, Value: "echo", Start: 41},
		{Tag: scanner.EOF, Value: "", Start: 46},
	}
	s := scanner.NewScanner(strings.NewReader(program))
	for i, tok := range tokens {
		got := s.Scan()
		if !reflect.DeepEqual(tok, got) {
			t.Errorf("expected token %v to be %v, got %v", i, tok, got)
		}
	}
}
