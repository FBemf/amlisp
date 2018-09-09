// The one that makes the code from the tree.
package codegen

import (
        "lexparse"
        "strconv"
        "sync"
)
import "fmt"

func CodeGen(ast lexparse.Ast) {
        fmt.Print("")
}

// Symbol table of builtin funcs
const builtins map[string]string = map[string]string{
        "quote": "internalQuote",
        "+": "+",
        "-": "-",
        "cons": "cons",
        "car": "car",
        "cdr": "cdr",
        "empty": "empty",
        "if": "if",
        "define": "define",
        "FUNCTION": "func"
}

type safeSym struct {
        table map[string]int
        mutex sync.Mutex
}

func call(up chan assembly, ast lexparse.Ast, counter func() int, sym *safeSym) {

        if p := ast.Primitive(); p != nil {
                switch p.Type() {
                        case lexparse.litInt:
                                up <- assembly{"NEW", r0, 3, 0}
                                up <- assembly{"SET", r0, 0, 1} // refcount
                                up <- assembly{"SET", r0, 1, Type_integer} // type
                                up <- assembly{"SET", r0, 2, strconv.Atoi(p.Value())} // val
                        case lexparse.Symbol:
                                up <- assembly{"NEW", r0, 3, 0}
                                up <- assembly{"SET", r0, 0, 1} // refcount
                                up <- assembly{"SET", r0, 1, Type_symbol} // type
                                name := strconv.Atoi(p.Value())
                                sym.mutex.Lock()
                                if id, ok := sym.table[name]; ok {
                                        up <- assembly{"SET", r0, 2, id)} // val
                                } else {
                                        sym.table[name] = counter()
                                        up <- assembly{"SET", r0, 2, sym.table[name])} // val
                                }
                                sym.mutex.Unlock()
                        default:
                                fmt.Println("Unexpected primitive type!")
                        // TODO: add support for other primitives
                }
        }

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
                up <- assembly{"JUMP", funcEnd, 0, 0}
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

        // Note on symbols: A symbol is just a unique name. We'll need an internal function called RESOLVE
        // that will dereference a symbol using the table and return a pointer directly to the value you want.
        // DEFINE adds a new definition to the symbol table

        up <- assembly{"NEW", r2, members+7, 0}
        up <- assembly{"SETID", r0, 0, r2}
        up <- assembly{"SETI", r2, 0, 1}
        up <- assembly{"SETI", r2, 1, Type_environment}
        up <- assembly{"SETI", r2, 2, members+4}
        up <- assembly{"SETID", r2, 4, r0}
        up <- assembly{"SETID", r2, 5, r1}
        up <- assembly{"SETD", r3, r1, 6}
        up <- assembly{"SETD", r3, r3, 0}
        up <- assembly{"SETID", r1, 6, r3}
        argCode := make([]chan assembly, members)
        for m := 0; m < members; m++ {
                argCode[m] = make(chan assembly, 100)
                ast = ast.Node().Next()
                go call(argCode[m], ast.Node().This(), counter, sym)
        }
        for m := 0; m < members; m++ {
                up <- assembly{"SETD", r0, r2, 7+m}
                up <- assembly{"SETD", r1, r2, 0}
                for {
                        if a, b := <-argCode[m]; b {
                                up <- a
                        } else {
                                break
                        }
                }
        }

        if isFunc {
                up <- assembly{"SETD", r3, r2, members+6}
                up <- assembly{"SETD", r3, r3, 0}
                up <- assembly{"SETID", r0, 0, r3}      // return

                up <- assembly{"ADD1I", r3, 0, 0}       // add to refcount

                up <- assembly{"SETD", r3, r2, 3}
                up <- assembly{"SETD", r3, r3, 0}       // Grab the pc to return to

                up <- assembly{"SUB1I", r2, 0, 0}       // decrement refcount

                up <- assembly{"SETD", r4, r2, 0}
                // Ascend the registers to the previous environment
                up <- assembly{"SETD", r2, r1, 0}
                up <- assembly{"SETD", r1, r2, 5}
                up <- assembly{"SETD", r1, r1, 0}
                up <- assembly{"SETD", r0, r2, 4}
                up <- assembly{"SETD", r0, r0, 0}

                up <- assembly{"REMEMBERJUMP", DUMPFUNC, r5, 0}
                // where dumpfunc is the address of our garbage collector
                // lemme outline how it works here: if the refcount of the place dumpfunc
                // was called is 0, it runs dumpfunc on all of that place's pointers, then
                // reduces that refcount to -1. Then the memory table will deallocate it
                // the next time it sees it.

                up <- assembly{"JUMPD", r3, 0, 0}      // jump back
                up <- assembly{"LABEL", funcEnd, 0, 0}

                // TODO: Should return a closure structure
        } else {
                // Inject args into env, descend registers, call closure
                //up <- assembly{"REMEMBER-JUMP", r2, 3, sym[whatever]} // saves pc+1 to [r2]+3
        }

        return
}

type assembly struct {
        command string
        arg1 int
        arg2 int
        arg3 int
}

const (
        Type_environment = iota
        Type_cons
        Type_vector
        Type_integer
        Type_symbol
)

const (
        r1 = iota
        r2
        r3
        r4
        r5
)

func makeCounter(i int) func() int {
        var mux = &sync.Mutex{}
        return func() int {
                mux.Lock()
                o := i
                i++
                mux.Unlock()
                return o
        }
}


/*func copySym(sym map[string]int) (out map[string]int) {
        out := make(map[string]int)
        for j, k := range(sym) {
                out[j] = k

        }
        return
}*/
