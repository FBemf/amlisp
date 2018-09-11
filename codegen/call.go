// The one that makes the code from the tree.
package codegen

import (
        "lexparse"
        "strconv"
        "sync"
)
import "fmt"

func call(up chan assembly, ast lexparse.Ast, counter func() int, sym *safeSym) {

        if p := ast.Primitive(); p != nil {
                switch p.Type() {
                        case lexparse.litInt:
                                up <- assembly{"NEW", r0, 3, 0}
                                up <- assembly{"SET-INDEXED", r0, 0, 1} // refcount
                                up <- assembly{"SET-INDEXED", r0, 1, Type_integer} // type
                                up <- assembly{"SET-INDEXED", r0, 2, strconv.Atoi(p.Value())} // val
                        case lexparse.Symbol:
                                up <- assembly{"NEW", r0, 3, 0}
                                up <- assembly{"SET-INDEXED", r0, 0, 1} // refcount
                                up <- assembly{"SET-INDEXED", r0, 1, Type_symbol} // type
                                name := strconv.Atoi(p.Value())
                                up <- assembly{"SET-LITERAL", r0, 2, sym.getSymID(name, counter))} // val
                                // TODO: automatically resolve symbol literals
                                // TODO: add keyword support for symbol-quote
                        default:
                                fmt.Println("Unexpected primitive type!")
                        // TODO: add support for other primitives
                }
        }

        /* FYI: A note on SET (where [A] is the value held in cell number A)
                SET-LITERAL R A sets the cell R to A
                COPY-ADD R A N sets the cell R to [A]+N
                SET-INDEXED R I A sets the cell [R]+I to A
                COPY-INDEXED R I A sets the cell [R]+I to [A]
                DEREF R A I sets the cell R to [[A]+I]
        */

        var isFunc bool = false
        if p := ast.Node().This().Primitive(); p.Type() == lexparse.Symbol {
                if p.Value() == builtins["FUNCTION"] {
                        isFunc = true
                }
        }

        // If this is a function definition, defer execution
        if isFunc {
                funcStart := counter()
                funcEnd := counter()
                up <- assembly{"JUMPLABEL", funcEnd, 0, 0}
                up <- assembly{"LABEL", funcStart, 0, 0}

                // Also get rid of the arg list
                ast = ast.Node().Next()
                argAst := ast.Node().This()
                ast = ast.Node().Next()
        }

        // Count members of s-exp
        members := 0
        for t := ast.Node(); t != nil; t = t.Next().Node() {
                if t.This() != nil {
                        members++
                }
        }

        // Make an env for it
        // an env has refcount, type, length, saved pc, return loc, parent env, symbol table, pointers

        up <- assembly{"NEW", r2, members+7, 0}
        up <- assembly{"COPY-INDEXED", r0, 0, r2} // Assumes r0 is return loc
        up <- assembly{"SET-INDEXED", r2, 0, 1}
        up <- assembly{"SET-INDEXED", r2, 1, Type_environment}
        up <- assembly{"SET-INDEXED", r2, 2, members+7}
        up <- assembly{"COPY-INDEXED", r2, 4, r0}
        up <- assembly{"COPY-INDEXED", r2, 5, r1} // Assumes r1 is return env
        up <- assembly{"DEREF", r3, r1, 6}
        up <- assembly{"COPY-INDEXED", r1, 6, r3}
        argCode := make([]chan assembly, members)
        for m := 0; m < members; m++ {
                argCode[m] = make(chan assembly, 100)
                ast = ast.Node().Next()
                go call(argCode[m], ast.Node().This(), counter, sym)
        }
        for m := 0; m < members; m++ {
                up <- assembly{"COPY-ADD", r0, r2, 7+m}
                up <- assembly{"COPY-ADD", r1, r2, 0}
                for {
                        if a, b := <-argCode[m]; b {
                                up <- a
                        } else {
                                break
                        }
                }
        }

        if isFunc {
                up <- assembly{"DEREF", r3, r2, members+6}
                up <- assembly{"COPY-INDEXED", r0, 0, r3}      // return

                up <- assembly{"ADD1", r3, 0, 0}       // add to closure refcount ([r3]++)

                up <- assembly{"DEREF", r3, r2, 3}       // Grab the pc to return to

                up <- assembly{"SUB1", r2, 0, 0}       // decrement env refcount ([r2]--)

                up <- assembly{"COPY-ADD", r4, r2, 0}       // the argument for dumpfunc

                // Ascend the registers to the previous environment
                up <- assembly{"COPY-ADD", r2, r1, 0}
                up <- assembly{"DEREF", r1, r2, 5}
                up <- assembly{"DEREF", r0, r2, 4}

                up <- assembly{"SET-LITERAL", r5, r5, 0}        // replace this if I find
                                                                // a better place to keep
                                                                // the return loc
                // TODO: This whole area needs to be reworked, preferably when I write dumpfunc
                /* Anatomy of dumpfunc:
                        - Lives on its own linked data structure
                        (refcount, type, current-target, next-dump-struct, return loc)
                        - Sets refcount to -1
                        - Switch on type to find if it's a pointer struct
                        - If it is, exit
                        - If it isn't, get the length of it
                        - Decrement each pointer's data struct. If they're 0,
                          create a new dump struct between this and the next one
                        - Upon exit, move to next dump struct and continue. If nil,
                          just return.
                        - Fun fact: its own refcount starts at 0. When it moves to the
                          next dump struct, it sets its own refcount to -1
                */

                up <- assembly{"REMEMBER-JUMP-LABEL", DUMPFUNC, r5, 0}
                // where dumpfunc is the address of our garbage collector
                // lemme outline how it works here: if the refcount of the place dumpfunc
                // was called is 0, it runs dumpfunc on all of that place's pointers, then
                // reduces that refcount to -1. Then the memory table will deallocate it
                // the next time it sees it.

                up <- assembly{"JUMP", r3, 0, 0}      // jump back
                up <- assembly{"LABEL", funcEnd, 0, 0}

                // Count number of args
                args := 0
                for t := argAst.Node(); t != nil; t = t.Next().Node() {
                        if t.This() != nil {
                                args++
                        }
                }

                // create and populate a closure to return
                // contents of a closure: refcount, type, pc addr, env loc, length, args ...
                up <- assembly{"NEW", r3, args+4, 0}
                up <- assembly{"SET-INDEXED", r3, 0, 1}
                up <- assembly{"SET-INDEXED", r3, 1, Type_closure}
                up <- assembly{"SET-INDEXED", r3, 2, funcStart}
                up <- assembly{"COPY-INDEXED", r3, 3, r2}
                up <- assembly{"COPY-INDEXED", r3, 4, args+5}
                for i := 0; i < args; i++ {
                        up <- assembly{"NEW", r4, 3, 0}
                        up <- assembly{"SET-INDEXED", r4, 0, 1}
                        up <- assembly{"SET-INDEXED", r4, 1, Type_symbol}
                        up <- assembly{"SET-INDEXED", r4, 2, getSymID(argAst.Node().This().Primitive().Value(), counter)}
                        up <- assembly{"COPY-INDEXED", r3, 5+i, r4}
                }

                up <- assembly{"COPY-INDEXED", r0, 0, r3}      // return it

                // Ascend the registers to the previous environment
                // BTW, rule of thumb for registers: r0 is where to write the return val,
                // r1 is the location of the parent environment, r2 is the location of the
                // current environment
                up <- assembly{"COPY-ADD", r2, r1, 0}
                up <- assembly{"DEREF", r1, r2, 5}
                up <- assembly{"DEREF", r0, r2, 4}

        } else {
                up <- assembly{"DEREF", r3, r2, 7}       // grab address of func sym
                up <- assembly{"DEREF", r3, r3, 2}       // grab sym id

                up <- assembly{"COPY-ADD", r6, r2, 6}       // grab sym table location
                up <- assembly{"DEREF", r4, r2, 6}       // grab sym table

                up <- assembly{"REMEMBERJUMP", // jump to the function that finds a symbol from the sym table

                // This one'll do it -- put it elsewhere though
                loop := counter()
                end := counter()
                up <- assembly{"LABEL", loop, 0, 0}
                up <- assembly{"DEREF", r5, r4, 1}
                up <- assembly{"JUMP-LABEL-IF-EQ", end, r3, r5} // leave the loop if the val of r3 == the val of r5
                up <- assembly{"DEREF", r4, r4, 4}       // else move on to next link in chain
                up <- assembly{"JUMP-LABEL", loop, 0, 0}
                up <- assembly{"LABEL", end, 0, 0}
                // TODO: MOVE THIS ^ ?
                // Should encapsulate as a function and call it in the primitive discovery section above

                up <- assembly{"DEREF", r4, r4, 2}      // grab closure

                for m := 1; m < members; m++ {
                        up <- assembly{"NEW", r5, 5, 0}
                        up <- assembly{"SET-INDEXED", r5, 0, 1}
                        up <- assembly{"SET-INDEXED", r5, 1, Type_symtab}
                        //loop = counter()
                        //end = counter()
                        //up <- assembly{"LABEL", loop, 0, 0}
                        // you now have the closure and a new symbol table cell.

                        up <- assembly{"DEREF", r3, r4, 4+m} // Got the address of the symbol
                        up <- assembly{"DEREF", r3, r3, 2} // Got the ID of the symbol

                        // grab the appropriate symbol from the closure and set the ID in the symbol table cell to its IDs
                        up <- assembly{"COPY-INDEXED", r5, 2, r3}

                        // then set the location to the contents of the appropriate cell in your current env.
                        up <- assembly{"DEREF", r3, r2, 7+m}
                        up <- assembly{"COPY-INDEXED", r5, 3, r3}

                        // Then plonk this onto the front of the symtab and continue
                        up <- assembly{"DEREF", r3, r6, 0}
                        up <- assembly{"COPY-INDEXED", r5, 4, r3}
                        up <- assembly{"COPY-INDEXED", r6, 0, r5}
                }

                // Lastly, make the jump
                up <- assembly{"DEREF", r4, r4, 3}      // Grab jump location
                up <- assebly{"COPY-ADD", r3, r2, 3}
                // TODO: Set r0 to be the env, set r1 to be the return pc
                up <- assembly{"REMEMBER-JUMP", r6, r3, 0} // saves next pc to [r3] and jumps to [r4]
                // TODO: Recover & ascend registers?
        }

        return
}
