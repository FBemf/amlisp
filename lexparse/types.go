package lexparse

/* TODO the AST needs work
        the idea right now is that
        it holds the functions this, next,
        and isempty, all of which return
        another ast. And one you want somethign
        out of it, you call Node or Primitive on it,
        and it'll either give you the raw node, the
        raw primitive, or nil. But until then, there's
        no using nil at all.


        This file now holds with that; now you need to
        propogate it through the rest of the project
*/


type Ast interface {
	Node() *node
	Primitive() *primitive
        This() Ast
        Next() Ast
}

// The basic components of an amlisp program after
// all the reader macros have run.
type primitive struct {
	kind    int
	content string
}

type node struct {
	left  Ast
	right Ast
}

type empty struct {}

func (p *primitive) Type() int {
        if p != nil {
	        return p.kind
        } else {
                return NilType
        }
}

func (p *primitive) Value() string {
	if p != nil {
                return p.content
        } else {
                return ""
        }
}

const (
        NilType = iota
	Symbol
	LitInt
	LitFloat
	LitChar
	LitStr
	openParen
	closeParen
)

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

func (p *primitive) Node() *node {
	return nil
}

func (p *primitive) Primitive() *primitive {
	return p
}

func (p *primitive) This() Ast {
	return &empty{}
}

func (p *primitive) Next() Ast {
	return &empty{}
}

func (n *node) Node() *node {
	return n
}

func (n *node) Primitive() *primitive {
	return nil
}

func (n *node) This() Ast {
	return n.left
}

func (n *node) Next() Ast {
	return n.right
}

func (e *empty) Node() *node {
        return nil
}

func (e *empty) Primitive() Ast {
        return nil
}

func (e *empty) This() Ast {
	return e
}

func (e *empty) Next() Ast {
	return e
}
