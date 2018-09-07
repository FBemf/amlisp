package lexparse

// This one turns the token list into
// an AST.

// AST nodes
type node struct {
	left  Ast
	right Ast
}

func Parse(prims []primitive) (Ast, error) {
	s := stack{} // stack of nodes to back up through tree structure
	a := new(node)
        top := a

	for _, p := range prims {
		switch p.kind { // Symbol, openParen, closeParen, LitInt, LitFloat, LitChar
		case openParen:
			s.push(a)
			a.left = new(node)
			a = a.left.(*node)
		case closeParen:
			b, ok := s.pop()
			if ok == false {
				return nil, errorString{"Unexpected ')'"}
			}
			a.left = nil
			a.right = nil
			a = b
			a.right = new(node)
			a = a.right.(*node)
		default:
			a.left = &p
			a.right = new(node)
			a = a.right.(*node)
		}
	}
	if !s.isEmpty() {
		return nil, errorString{"Unterminated '('"}
	}
	return top, nil
}

