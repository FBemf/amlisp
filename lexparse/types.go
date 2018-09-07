package lexparse

type Ast interface {
	Node() *node
	Primitive() *primitive
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
