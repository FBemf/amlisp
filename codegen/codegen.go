// The one that makes the code from the tree.
package codegen

import "lexparse"
import "fmt"

func CodeGen(ast lexparse.Ast) {
        fmt.Print("")
}

// Symbol table of builtin funcs
const masterSym map[string]int = map[string]int{
        "_quote": 0,
        "_+": 1,
        "_-": 2,
        "_cons": 3,
        "_car": 4,
        "_cdr": 5,
        "_empty": 6,
        "_if": 7,
        "_define": 8,
        "_func": 9
}

func callFunc(up chan assembly, ast lexparse.Ast, sym map[string]int) {
        /*
                Requirements:
                Get a function location
                Get a parent env
                Make your own env
                Store the pc
                Hop the pc
                Go nuts
        */
        /*
                Model Environment:
                refcount
                typeconst
                oldpc
                length
                pointers
                ...
        /*
        /*
                Model other type in memory:
                refcount
                typeconst
                values
                ...
        */
        // Count top-level "define"s
        defines := 0
        for t := ast.Node(); t != nil; t = t.Next().Node() {
                if i := t.This().Node()); i != nil {
                        if j := i.This.Primitive(); j != nil {
                                if j.Type() == lexparse.Symbol {
                                        if masterSym[j.Value()] == masterSym["_define"] {
                                                defines++
                                        }
                                }
                        }
                }
        }


        // Make an env for it
        // Either set each arg from a raw or go callFunc it
        // Set oldpc
        // jump
}

func copySym(sym map[string]int) (out map[string]int) {
        out := make(map[string]int)
        for j, k := range(sym) {
                out[j] = k

        }
        return
}

type assembly struct {
        command string
        arg1 int
        arg2 int
        arg3 int
}
