package lexparse

/* TODO the AST needs work
   the idea Right now is that
   it holds the functions this, next,
   and isEmpty, all of which return
   another ast. And one you want somethign
   out of it, you call Node or Primitive on it,
   and it'll either give you the raw Node, the
   raw Primitive, or nil. But until then, there's
   no using nil at all.


   This file now holds with that; now you need to
   propogate it through the rest of the project
*/

type Ast interface {
	Node() *Node
	Primitive() *Primitive
	This() Ast
	Next() Ast
	IsEmpty() bool
}

// The basic components of an amlisp program after
// all the reader macros have run.
type Primitive struct {
	Kind    int
	Content string
}

type Node struct {
	Left  Ast
	Right Ast
}

type Empty struct{}

func (p *Primitive) Type() int {
	if p != nil {
		return p.Kind
	} else {
		return NilType
	}
}

func (p *Primitive) Value() string {
	if p != nil {
		return p.Content
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
	val  *Node
	next *frame
}

func (s *stack) push(p *Node) {
	s.head = &frame{p, s.head}
}

func (s *stack) pop() (*Node, bool) {
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

func (p *Primitive) Node() *Node {
	return nil
}

func (p *Primitive) Primitive() *Primitive {
	return p
}

func (p *Primitive) This() Ast {
	return &Empty{}
}

func (p *Primitive) Next() Ast {
	return &Empty{}
}

func (p *Primitive) IsEmpty() bool {
	return false
}

func (n *Node) Node() *Node {
	return n
}

func (n *Node) Primitive() *Primitive {
	return nil
}

func (n *Node) This() Ast {
	return n.Left
}

func (n *Node) Next() Ast {
	return n.Right
}

func (n *Node) IsEmpty() bool {
	return false
}

func (e *Empty) Node() *Node {
	return nil
}

func (e *Empty) Primitive() *Primitive {
	return nil
}

func (e *Empty) This() Ast {
	return e
}

func (e *Empty) Next() Ast {
	return e
}

func (e *Empty) IsEmpty() bool {
	return true
}
