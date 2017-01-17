package parser

type Node interface{}

type Statement struct {
	Command Node
	Args    []Node
	Redirs  []Redir
}

type Redir struct{}
