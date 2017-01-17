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

func (p *Parser) Parse() (Expression, error) {
	return p.parseExpressions()
}

func (p *Parser) parseExpressions() (Expression, error) {
	var list []Expression
	for {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if expr == nil {
			break
		}
		list = append(list, expr)
		p.accept(scanner.BREAK)
	}
	return &ExpressionList{list: list}, nil
}

func (p *Parser) parseExpression() (Expression, error) {
	var (
		expr Expression
		err  error
	)
	t := p.scanIgnoreWs()
	switch t.Tag {
	case scanner.EOF:
		return nil, nil
	case scanner.STARTGRP:
		expr, err = p.parseExpressionGroup()
	case scanner.CHUNK:
		expr, err = p.parseCommand(t.Value)
	case scanner.NOT:
		expr, err = p.parseNotExpression()
	}
	if err != nil {
		return nil, err
	}
	if expr == nil {
		return nil, p.expect("Expression")
	}
	// Lookahead 1 token for a possible binary operator
	next := p.scanIgnoreWs()
	switch next.Tag {
	case scanner.AND:
		expr, err = p.parseAndExpression(expr)
	case scanner.OR:
		expr, err = p.parseOrExpression(expr)
	case scanner.PIPE:
		expr, err = p.parsePipeExpression(expr)
	default:
		p.unscan()
	}
	return expr, err
}

func (p *Parser) parseNotExpression() (Expression, error) {
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if expr == nil {
		return nil, p.expect("Expression")
	}
	return &Not{expr: expr}, nil
}

func (p *Parser) parseAndExpression(lhs Expression) (Expression, error) {
	rhs, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if rhs == nil {
		return nil, p.expect("Expression")
	}
	return &And{LHS: lhs, RHS: rhs}, nil
}

func (p *Parser) parseOrExpression(lhs Expression) (Expression, error) {
	rhs, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if rhs == nil {
		return nil, p.expect("Expression")
	}
	return &Or{LHS: lhs, RHS: rhs}, nil
}

func (p *Parser) parsePipeExpression(lhs Expression) (Expression, error) {
	rhs, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if rhs == nil {
		return nil, p.expect("Expression")
	}
	return &Pipe{LHS: lhs, RHS: rhs}, nil
}

func (p *Parser) parseCommand(first string) (Expression, error) {
	var args []Chunk
	for {
		t := p.scanIgnoreWs()
		if t.Tag != scanner.CHUNK {
			p.unscan()
			break
		}
		args = append(args, Chunk(t.Value))
	}
	var redirs []Redir
	for {
		var stream RedirStream
		next := p.scanIgnoreWs()
		if next.Tag == scanner.REDIRERR {
			stream = RedirStreamErr
		} else if next.Tag == scanner.REDIROUT {
			stream = RedirStreamOut
		} else {
			p.unscan()
			break
		}
		// redirect must be followed by a chunk
		next = p.scanIgnoreWs()
		if next.Tag != scanner.CHUNK {
			return nil, p.expect("Chunk")
		}
		redir := Redir{
			Stream: stream,
			Target: Chunk(next.Value),
		}
		redirs = append(redirs, redir)
	}
	cmd := &Command{
		Name:       Chunk(first),
		Args:       args,
		Redirs:     redirs,
		Background: p.accept(scanner.BACKGRND),
	}
	return cmd, nil
}

func (p *Parser) parseExpressionGroup() (Expression, error) {
	expr, err := p.parseExpressions()
	if err != nil {
		return nil, err
	}
	if !p.accept(scanner.ENDGRP) {
		return nil, p.expect(")")
	}
	return expr, nil
}

func (p *Parser) scan() scanner.Token {
	if p.buffered {
		p.buffered = false
		return p.last
	}
	p.last = p.s.Scan()
	return p.last
}

func (p *Parser) unscan() {
	p.buffered = true
}

func (p *Parser) scanIgnoreWs() scanner.Token {
	t := p.scan()
	for t.Tag == scanner.WS {
		t = p.scan()
	}
	return t
}

func (p *Parser) accept(tag scanner.Tag) bool {
	t := p.scanIgnoreWs()
	if t.Tag == tag {
		return true
	}
	p.unscan()
	return false
}

func (p *Parser) expect(s string) error {
	return errors.New(s) // FIXME implement
}
