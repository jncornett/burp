package parser

import (
	"strings"

	"github.com/jncornett/burp/lib/context"
)

type Expression interface {
	Eval(context.Context) (bool, error)
}

type Template interface {
	Eval(context.Context) (string, error)
}

type ExpressionList struct {
	list []Expression
}

func (el ExpressionList) Eval(ctx context.Context) (b bool, err error) {
	for _, expr := range el.list {
		b, err = expr.Eval(ctx)
		if err != nil {
			return false, err
		}
	}
	return b, nil
}

type Chunk string

func (ch Chunk) Eval(ctx context.Context) (string, error) {
	templ, err := NewStringParser(strings.NewReader(string(ch))).Parse()
	if err != nil {
		return "", err
	}
	return templ.Eval(ctx)
}

type Not struct {
	expr Expression
}

func (n Not) Eval(ctx context.Context) (bool, error) {
	r, err := n.expr.Eval(ctx)
	if err != nil {
		return false, err
	}
	return !r, nil
}

type And struct {
	LHS Expression
	RHS Expression
}

func (a And) Eval(ctx context.Context) (bool, error) {
	lhs, err := a.LHS.Eval(ctx)
	if err != nil {
		return false, err
	}
	if !lhs {
		// short circuit
		return false, nil
	}
	return a.RHS.Eval(ctx)
}

type Or struct {
	LHS Expression
	RHS Expression
}

func (o Or) Eval(ctx context.Context) (bool, error) {
	lhs, err := o.LHS.Eval(ctx)
	if err != nil {
		return false, err
	}
	if lhs {
		// short circuit
		return true, nil
	}
	return o.RHS.Eval(ctx)
}

type Pipe struct {
	LHS Expression
	RHS Expression
}

func (p Pipe) Eval(ctx context.Context) (bool, error) {
	return false, nil // TODO Implement
}

type Command struct {
	Name       Chunk
	Args       []Chunk
	Redirs     []Redir
	Background bool // unused
}

func (c Command) Eval(ctx context.Context) (bool, error) {
	name, err := c.Name.Eval(ctx)
	if err != nil {
		return false, err
	}
	var args []string
	for _, arg := range c.Args {
		s, err := arg.Eval(ctx)
		if err != nil {
			return false, err
		}
		args = append(args, s)
	}
	// FIXME implement redirections
	return ctx.Run(name, args)
}

type RedirStream int

const (
	RedirStreamNone = iota
	RedirStreamErr
	RedirStreamOut
)

type Redir struct {
	Stream RedirStream
	Target Chunk
}
