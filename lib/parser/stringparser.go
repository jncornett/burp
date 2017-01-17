package parser

import (
	"io"

	"github.com/jncornett/burp/lib/context"
	"github.com/jncornett/burp/lib/scanner"
)

type StringParser struct {
	s scanner.Scanner
}

func NewStringParser(r io.Reader) *StringParser {
	return &StringParser{s: scanner.NewScanner(r)}
}

func (p *StringParser) Parse() (Template, error) {
	templ := listTemplate{}
	return &templ, nil
}

type listTemplate struct {
	list []Template
}

func (t listTemplate) Eval(ctx context.Context) (string, error) {
	return "", nil
}

func (t *listTemplate) push(templ Template) {
	t.list = append(t.list, templ)
}
