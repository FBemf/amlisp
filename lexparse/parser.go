package lexparse

import "fmt"

// This one turns the token list into
// an AST.

func Parse(prims []primitive) (Ast, error) {
	s := stack{} // stack of nodes to back up through tree structure
        var emptynode Ast = &empty{}
	a := &node{emptynode, emptynode}
	top := a

	for _, p := range prims {
		switch p.kind { // Symbol, openParen, closeParen, LitInt, LitFloat, LitChar
		case openParen:
			s.push(a)
			a.left = &node{emptynode, emptynode}
			a = a.left.(*node)
		case closeParen:
			b, ok := s.pop()
			if ok == false {
				return nil, errorString{"Unexpected ')'"}
			}
			a.left = emptynode
			a.right = emptynode
			a = b
			a.right = &node{emptynode, emptynode}
			a = a.right.(*node)
		default:
                        newP := p
			a.left = &newP
			a.right = &node{emptynode, emptynode}
			a = a.right.(*node)
		}
	}
	if !s.isEmpty() {
		return nil, errorString{"Unterminated '('"}
	}
	return top, nil
}

func RPrint(ast Ast) string {
	if ast.IsEmpty() {
		return "empty"
	} else if ast.Node() != nil {
		return fmt.Sprintf("(%s.%s)", RPrint(ast.This()), RPrint(ast.Next()))
	} else {
		return fmt.Sprintf("%v", ast.Primitive())
	}
}
