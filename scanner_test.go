package burp_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jncornett/burp"
)

func TestScannerScan(t *testing.T) {
	tests := []struct {
		Program string
		Tag     burp.Tag
		Value   string
	}{
		{``, burp.EOF, ""},
		{`{foobar}`, burp.VAR, "foobar"},
		{`"foobar"`, burp.TEXT, "foobar"},
		{`"foo\"bar"`, burp.TEXT, "foo\"bar"},
		{"[", burp.STARTEXEC, "["},
		{"]", burp.ENDEXEC, "]"},
		{"(", burp.STARTGROUP, "("},
		{")", burp.ENDGROUP, ")"},
		{"^", burp.REDIRERR, "^"},
		{">", burp.REDIROUT, ">"},
		{";", burp.BREAK, ";"},
		{":", burp.IDENTITY, ":"},
		{"|", burp.PIPE, "|"},
		{"|x", burp.PIPE, "|"},
		{"&&", burp.AND, "&&"},
		{"&x", burp.ILLEGAL, "&"},
		{"||", burp.OR, "||"},
		{"    ", burp.WS, "    "},
		{"032abczz_ZXYAB", burp.TEXT, "032abczz_ZXYAB"},
		{"?", burp.ILLEGAL, "?"},
	}
	for _, test := range tests {
		t.Run(test.Program, func(t *testing.T) {
			s := burp.NewScanner(strings.NewReader(test.Program))
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
	tokens := []burp.Token{
		{Tag: burp.TEXT, Value: "foo", Start: 0},
		{Tag: burp.WS, Value: " ", Start: 3},
		{Tag: burp.AND, Value: "&&", Start: 4},
		{Tag: burp.TEXT, Value: "bar", Start: 6},
		{Tag: burp.OR, Value: "||", Start: 9},
		{Tag: burp.STARTGROUP, Value: "(", Start: 11},
		{Tag: burp.VAR, Value: "this.var", Start: 12},
		{Tag: burp.PIPE, Value: "|", Start: 22},
		{Tag: burp.TEXT, Value: "that stuff", Start: 23},
		{Tag: burp.STARTEXEC, Value: "[", Start: 35},
		{Tag: burp.TEXT, Value: "ok", Start: 36},
		{Tag: burp.ENDEXEC, Value: "]", Start: 38},
		{Tag: burp.ENDGROUP, Value: ")", Start: 39},
		{Tag: burp.BREAK, Value: ";", Start: 40},
		{Tag: burp.TEXT, Value: "echo", Start: 41},
		{Tag: burp.EOF, Value: "", Start: 46},
	}
	s := burp.NewScanner(strings.NewReader(program))
	for i, tok := range tokens {
		got := s.Scan()
		if !reflect.DeepEqual(tok, got) {
			t.Errorf("expected token %v to be %v, got %v", i, tok, got)
		}
	}
}
