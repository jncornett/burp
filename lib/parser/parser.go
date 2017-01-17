package parser

import (
	"errors"
	"io"

	"github.com/jncornett/burp/lib/scanner"
)

type Parser struct {
	s        scanner.Scanner
	last     scanner.Token
	buffered bool
}

func NewParser(r io.Reader) Parser {
	return Parser{s: scanner.NewScanner(r)}
}

func (p *Parser) Parse() (Node, error) {
	return nil, nil
}

func (p *Parser) scan() scanner.Token {
	if p.buffered {
		p.buffered = false
		return p.last
	}
	p.last = p.s.Scan()
	return p.last
}

func (p *Parser) scanIgnoreWs() scanner.Token {
	t := p.scan()
	for t.Tag == scanner.WS {
		t = p.scan()
	}
	return t
}

func (p *Parser) unscan() {
	p.buffered = true
}

func (p *Parser) expect(s string) error {
	return errors.New(s)
}

func (p *Parser) parseExpression() (Node, error) {
	t := p.scanIgnoreWs()
	switch t.Tag {
	case scanner.STARTGROUP:
		return p.parseGroup()
	}
	return nil, nil
}

func (p *Parser) parseGroup() (Node, error) {
	node, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return nil, err
}

// parse a complete statement
func (p *Parser) parseStatement() (Node, error) {
	nodes, err := p.parseComponents()
	if err != nil {
		return nil, err
	}
	// need at least one component as the 'command'
	if len(nodes) < 1 {
		return nil, p.expect("Component")
	}
	redirs, err := p.parseRedirs()
	if err != nil {
		return nil, err
	}
	return &Statement{Command: nodes[0], Args: nodes[1:], Redirs: redirs}, nil
}

func (p *Parser) parseComponents() ([]Node, error) {
	var nodes []Node
	for {
		node, err := p.parseComponent()
		if err != nil {
			return nil, err
		}
		if node == nil {
			break
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (p *Parser) parseRedirs() ([]Redir, error) {
	var redirs []Redir
	for {
		redir, err := p.parseRedir()
		if err != nil {
			return nil, err
		}
		if redir == nil {
			break
		}
		redirs = append(redirs, *redir)
	}
	return redirs, nil
}

func (p *Parser) parseRedir() (*Redir, error) {
	return nil, nil
}

func (p *Parser) parseComponent() (Node, error) {
	return nil, nil
}
