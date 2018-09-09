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

func (sym *safeSym) getSymID(name string, counter func() int) (r int) {
        sym.mutex.Lock()
        if id, ok := sym.table[name]; ok {
                r = id
        } else {
                r = counter()
                sym.table[name] = r
        }
        sym.mutex.Unlock()
        return
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
                                up <- assembly{"SET", r0, 2, sym.getSymID(name, counter))} // val
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
        up <- assembly{"SETID", r0, 0, r2}
        up <- assembly{"SETI", r2, 0, 1}
        up <- assembly{"SETI", r2, 1, Type_environment}
        up <- assembly{"SETI", r2, 2, members+7}
        up <- assembly{"SETID", r2, 3, r3}
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

                up <- assembly{"SETD", r4, r2, 0}       // the argument for dumpfunc

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

                // Count number of args
                args := 0
                for t := argAst.Node(); t != nil; t = t.Next().Node() {
                        if t.This() != nil {
                                args++
                        }
                }

                // create and populate a closure to return
                up <- assembly{"NEW", r3, args+4, 0}
                up <- assembly{"SETI", r3, 0, 1}
                up <- assembly{"SETI", r3, 1, funcStart}
                up <- assembly{"SETID", r3, 2, r2}
                up <- assembly{"SETID", r3, 3, args+4}
                for i := 0; i < args; i++ {
                        up <- assembly{"NEW", r4, 3, 0}
                        up <- assembly{"SETI", r4, 0, 1}
                        up <- assembly{"SETI", r4, 1, Type_symbol}
                        up <- assembly{"SETI", r4, 2, getSymID(argAst.Node().This().Primitive().Value(), counter)}
                        up <- assembly{"SETID", r3, 4+i, r4}
                }

                up <- assembly{"SETID", r0, 0, r3}      // return it

                // Ascend the registers to the previous environment
                // BTW, rule of thumb for registers: r0 is where to write the return val,
                // r1 is the location of the parent environment, r2 is the location of the
                // current environment
                up <- assembly{"SETD", r2, r1, 0}
                up <- assembly{"SETD", r1, r2, 5}
                up <- assembly{"SETD", r1, r1, 0}
                up <- assembly{"SETD", r0, r2, 4}
                up <- assembly{"SETD", r0, r0, 0}

        } else {
                up <- assembly{"SETD", r3, r2, 7}       // grab address with ptr to func sym
                up <- assembly{"SETD", r3, r3, 0}       // grab address of func sym
                up <- assembly{"SETD", r3, r3, 2}       // grab address of sym id
                up <- assembly{"SETD", r3, r3, 0}       // grab sym id

                up <- assembly{"SETD", r4, r2, 6}       // grab sym table loc
                up <- assembly{"SETD", r4, r4, 0}       // grab sym table

                loop := counter()
                end := counter()
                up <- assembly{"LABEL", loop, 0, 0}
                up <- assembly{"SETD", r5, r4, 1}
                up <- assembly{"SETD", r5, r5, 0}
                up <- assembly{"JUMPLABELIFEQ", end, r3, r5} // leave the loop if the val of r3 == the val of r5
                up <- assembly{"SETD", r4, r4, 3}
                up <- assembly{"SETD", r4, r4, 0}       // else move on to next link in chain
                up <- assembly{"JUMPLABEL", loop, 0, 0}
                up <- assembly{"LABEL", end, 0, 0}

                for m := 1; m < members; m++ {
                        up <- assembly{"NEW", r4, 4, 0}
                        up <- assembly{"SETI", r4, 0, 1}
                        up <- assembly{"SETI", r4, 1, 1}
                        //up <- assembly{"SETI", r2, 7+m, 5}

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
