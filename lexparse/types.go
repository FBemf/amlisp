package lexparse

// TODO: Make Node() and Primitive() return nil on nils so you can chain

type Ast interface {
	Node() *node
	Primitive() *primitive
}

// The basic components of an amlisp program after
// all the reader macros have run.
type primitive struct {
	kind    int
	content string
}

func (p primitive) Type() int {
	return p.kind
}

func (p primitive) Value() string {
	return p.content
}

const (
	Symbol = iota
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

func (n *node) This() Ast {
	return n.left
}

func (n *node) Next() Ast {
	return n.right
}

func (n *node) Node() *node {
	return n
}

func (n *node) Primitive() *primitive {
	return nil
}
