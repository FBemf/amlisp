package lexparse

import "fmt"

// This one turns the token list into
// an AST.

type ast interface {
	Node() *node
}

func Parse(prims []primitive) (ast, error) {
	s := stack{nil} // stack of nodes to back up through tree structure
	a := new(node)

	for _, p := range prims {
		fmt.Println(p)
		switch p.kind { // Symbol, openParen, closeParen, LitInt, LitFloat, LitChar
		case openParen:
			s.push(a)
			a.left = new(node)
			a = a.left.Node()
		case closeParen:
			b, ok := s.pop()
			if ok == false {
				return nil, errorString{"Unexpected ')'"}
			}
			a.left = nil
			a.right = nil
			a = b
			a.right = new(node)
			a = a.right.Node()
		default:
			a.left = p
			a.right = new(node)
			a = a.right.Node()
		}
	}
	if !s.isEmpty() {
		return nil, errorString{"Unterminated '('"}
	}

	for {
		if b, ok := s.pop(); ok {
			a = b
		} else {
                        break
                }
	}

	return a, nil
}

type errorString struct {
	s string
}

func (e errorString) Error() string {
	return e.s
}

// lex tokens held in this
type stack struct {
	head *frame
}

type frame struct {
	val  *node
	next *frame
}

func (s *stack) push(p *node) {
	s.head = &frame{p, s.head}
}

func (s *stack) pop() (*node, bool) {
	if s.head != nil {
		v := s.head.val
		s.head = s.head.next
		return v, true
	} else {
		return nil, false
	}
}

func (s *stack) isEmpty() bool {
	if s.head == nil {
		return true
	} else {
		return false
	}
}

func (p primitive) Node() *node {
	return nil
}

// AST nodes
type node struct {
	left  ast
	right ast
}

func (n node) This() ast {
	return n.left
}

func (n node) Next() ast {
	return n.right
}

func (n node) Node() *node {
	return &n
}
