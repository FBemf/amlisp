package lexparse

import "fmt"

// This one turns the token list into
// an AST.

func Parse(prims []Primitive) (Ast, error) {
	s := stack{} // stack of Nodes to back up through tree structure
        var emptyNode Ast = &Empty{}
	a := &Node{emptyNode, emptyNode}
	top := a

	for _, p := range prims {
		switch p.Kind { // Symbol, openParen, closeParen, LitInt, LitFloat, LitChar
		case openParen:
			s.push(a)
			a.Left = &Node{emptyNode, emptyNode}
			a = a.Left.(*Node)
		case closeParen:
			b, ok := s.pop()
			if ok == false {
				return nil, errorString{"Unexpected ')'"}
			}
			a.Left = emptyNode
			a.Right = emptyNode
			a = b
			a.Right = &Node{emptyNode, emptyNode}
			a = a.Right.(*Node)
		default:
                        newP := p
			a.Left = &newP
			a.Right = &Node{emptyNode, emptyNode}
			a = a.Right.(*Node)
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
