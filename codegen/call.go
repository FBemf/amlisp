// The one that makes the code from the tree.
package codegen

import (
        "../lexparse"
        "strconv"
        "fmt"
)

func call(up chan Assembly, ast lexparse.Ast, counter func() int, sym *safeSym, quoted bool) {
        if ast.IsEmpty() {
                close(up)
                return
        }
        if p := ast.Primitive(); p != nil {
                switch p.Type() {
                        case lexparse.LitInt:
                                up <- Assembly{"LITERAL INT _f", 0, 0, 0}
                                up <- Assembly{"NEW", r3, 3, 0}
                                up <- Assembly{"SET-INDEXED", r3, 0, 1} // refcount
                                up <- Assembly{"SET-INDEXED", r3, 1, Type_int} // type
                                val, _ := strconv.Atoi(p.Value())
                                up <- Assembly{"SET-INDEXED", r3, 2, val} // val
                                up <- Assembly{"SET-INDEXED", r0, 0, r3}        // return
                        case lexparse.Symbol:
                                if quoted {
                                        up <- Assembly{"LITERAL SYMBOL _f", 0, 0, 0}
                                        up <- Assembly{"NEW", r3, 3, 0}
                                        up <- Assembly{"SET-INDEXED", r3, 0, 1} // refcount
                                        up <- Assembly{"SET-INDEXED", r3, 1, Type_symbol} // type
                                        up <- Assembly{"SET-LITERAL", r3, 2, sym.getSymID(p.Value(), counter)} // val
                                        up <- Assembly{"COPY-INDEXED", r0, 0, r3}
                                } else {
                                        // dereferencing func start -- none of this block makes sense
                                        // a symtab frame looks like
                                        // [refcount] [type] [id] [loc] [next]
                                        up <- Assembly{"VARIABLE SYMBOL _f", 0, 0, 0}
                                        loop := counter()
                                        end := counter()
                                        up <- Assembly{"DEREF", r4, r2, 6} // grab symtab - r4 = [[r2]+6]
                                        up <- Assembly{"LABEL", loop, 0, 0}
                                        up <- Assembly{"DEREF", r3, r4, 2} // grab id
                                        // leave the loop if the val of r5 == the val of r3
                                        up <- Assembly{"JUMP-LABEL-IF-IS", end, r3, sym.getSymID(p.Value(), counter)}
                                        // TODO: Func does not break at end of symbol table.
                                        // Error catching needed all through this program. 
                                        // Let's get it to at least work on
                                        // good code first.

                                        up <- Assembly{"DEREF", r4, r4, 4}       // else move on to next link in chain
                                        up <- Assembly{"JUMP-LABEL", loop, 0, 0}
                                        up <- Assembly{"LABEL", end, 0, 0}

                                        up <- Assembly{"DEREF", r5, r4, 3}
                                        up <- Assembly{"COPY-INDEXED", r0, 0, r5} // return
                                }
                        default:
                                fmt.Printf("Unexpected primitive type %v\n", p.Type())
                        // TODO: add support for other primitives
                }
                close(up)
                return
        }

        /* FYI: A note on SET (where [A] is the value held in cell number A)
                SET-LITERAL R A sets the cell R to A
                COPY-ADD R A N sets the cell R to [A]+N
                SET-INDEXED R I A sets the cell [R]+I to A
                COPY-INDEXED R I A sets the cell [R]+I to [A]
                DEREF R A I sets the cell R to [[A]+I]
        */

        var isFunc bool = false
        if p := ast.This().Primitive(); (p != nil && p.Type() == lexparse.Symbol) {
                if p.Value() == builtins["FUNCTION"] {
                        isFunc = true
                } else if p.Value() == builtins["SYMBOL-QUOTE"] && !quoted {
                        quoted = true
                }
        }

        // If this is a function definition, defer execution
        var funcStart, funcEnd int;
        var argAst lexparse.Ast;
        if isFunc {
                funcStart = counter()
                funcEnd = counter()
                up <- Assembly{"FUNC DEF _f", 0, 0, 0}
                up <- Assembly{"JUMP-LABEL", funcEnd, 0, 0}
                up <- Assembly{"LABEL", funcStart, 0, 0}

                // Also get rid of the arg list
                ast = ast.Next().Node()
                argAst = ast.This().Node()
                ast = ast.Next().Node()
        }

        // Count members of s-exp
        members := 0
        for t := ast; t.Node() != nil; t = t.Next() {
                if (t.This().IsEmpty() == false) {
                        members++
                }
        }
        fmt.Printf("MEMBERS: %d\n", members)

        // Make an env for it
        // an env has refcount, type, length, saved pc, return loc, parent env, symbol table, pointers
        up <- Assembly{"NEW ENV _f", 0, 0, 0}

        up <- Assembly{"NEW", r2, members+7, 0}
        up <- Assembly{"COPY-INDEXED", r0, 0, r2} // Assumes r0 is return loc
        up <- Assembly{"SET-INDEXED", r2, 0, 1}
        up <- Assembly{"SET-INDEXED", r2, 1, Type_environment}
        up <- Assembly{"SET-INDEXED", r2, 2, members}
        up <- Assembly{"COPY-INDEXED", r2, 4, r0}
        up <- Assembly{"COPY-INDEXED", r2, 5, r1} // Assumes r1 is return env
        up <- Assembly{"DEREF", r3, r1, 6}
        up <- Assembly{"COPY-INDEXED", r2, 6, r3}       // grab symbol table
        argCode := make([]chan Assembly, members)
        up <- Assembly{"ARGS _f", 0, 0, 0}
        for m := 0; m < members; m++ {
                argCode[m] = make(chan Assembly, 100)
                fmt.Println(lexparse.RPrint(ast))
                if !quoted || m == 0 {
                        go call(argCode[m], ast.This(), counter, sym, false)
                } else {
                        go call(argCode[m], ast.This(), counter, sym, true)
                }
                ast = ast.Next()
        }
        for m, c := range argCode {
                up <- Assembly{"COPY-ADD", r0, r2, 7+m} // r0 = [r2] + 7+m
                up <- Assembly{"COPY-ADD", r1, r2, 0}   // r1 = [r2] + 0
                for a, b := <-c; b; a, b = <-c {
                        up <- a
                }
        }

        if isFunc {
                up <- Assembly{"DEREF", r3, r2, members+6}      // Grab return value
                up <- Assembly{"COPY-INDEXED", r0, 0, r3}      // return
                up <- Assembly{"JUMP-LABEL", sym.getSymID(builtins["FINISHFUNC"], counter), 0, 0}

                up <- Assembly{"LABEL", funcEnd, 0, 0}  // end of part of func executed when it's called
                up <- Assembly{"FUNC END _f", 0, 0, 0}

                // Where we land when we're defining the func
                // Count number of args
                args := 0
                for t := argAst; t.Node() != nil; t = t.Next() {
                        if t.This().IsEmpty() == false {
                                args++
                        }
                }

                // create and populate a closure to return
                // contents of a closure: refcount, type, pc addr, env loc, length, args ...
                up <- Assembly{"NEW", r3, args+4, 0}
                up <- Assembly{"SET-INDEXED", r3, 0, 1}
                up <- Assembly{"SET-INDEXED", r3, 1, Type_closure}
                up <- Assembly{"SET-INDEXED", r3, 2, funcStart}
                up <- Assembly{"COPY-INDEXED", r3, 3, r2}
                up <- Assembly{"COPY-INDEXED", r3, 4, args}
                for i := 0; i < args; i++ {
                        up <- Assembly{"NEW", r4, 3, 0}
                        up <- Assembly{"SET-INDEXED", r4, 0, 1}
                        up <- Assembly{"SET-INDEXED", r4, 1, Type_symbol}
                        up <- Assembly{"SET-INDEXED", r4, 2, sym.getSymID(argAst.Node().This().Primitive().Value(), counter)}
                        up <- Assembly{"COPY-INDEXED", r3, 5+i, r4}
                }

                up <- Assembly{"COPY-INDEXED", r0, 0, r3}      // return it
                // Finished actual function action

                up <- Assembly{"JUMP-LABEL", sym.getSymID(builtins["FINISHFUNC"], counter), 0, 0}
                up <- Assembly{"FUNC RETURN END _f", 0, 0, 0}

        } else {
                up <- Assembly{"DEREF", r4, r2, 7}      // grab closure
                up <- Assembly{"FUNC CALL _f", 0, 0, 0}

        // TODO: detect when a function is called with the wrong number of args
        // a symbol table is [refcount] [type] [id] [location] [next]
                for m := 1; m < members; m++ {
                        up <- Assembly{"NEW", r5, 5, 0}
                        up <- Assembly{"SET-INDEXED", r5, 0, 1}
                        up <- Assembly{"SET-INDEXED", r5, 1, Type_symtab}
                        // you now have the closure and a new symbol table cell.

                        // grab the ID of the relevant symbol
                        up <- Assembly{"DEREF", r3, r4, 4+m}    // r3 = [[r4] + 4+m]
                        up <- Assembly{"DEREF", r3, r3, 2}      // r3 = [[r3] + 2]

                        // Set the ID in the symbol table cell to the ID
                        up <- Assembly{"COPY-INDEXED", r5, 2, r3}       // [r5] + 2 = [r3]

                        // then set the location to the contents of the appropriate cell in your current env.
                        up <- Assembly{"DEREF", r3, r2, 7+m}    // r3 = [[r2] + 7+m]
                        up <- Assembly{"COPY-INDEXED", r5, 3, r3}       // [r5]+3 = [r3]

                        // Then plonk this onto the front of the symtab and continue
                        // A symbol table cell is [ refcount, type, symbol id, value location, next cell ]
                        up <- Assembly{"DEREF", r3, r2, 6}      // r3 = [[r2]+6]
                        up <- Assembly{"COPY-INDEXED", r5, 4, r3}       // [r5]+4 = [r3]
                        up <- Assembly{"COPY-INDEXED", r2, 6, r5}       // [r2]+6 = [r5]
                }

                // Lastly, make the jump
                up <- Assembly{"DEREF", r4, r4, 3}      // Grab jump location
                up <- Assembly{"COPY-ADD", r3, r2, 3}    // r3 = [r2] + 3
                up <- Assembly{"JUMP-LABEL-REMEMBER", r4, r3, 0} // saves next pc to [r3] and jumps to [r4]
        }
        close(up)
        return
}
