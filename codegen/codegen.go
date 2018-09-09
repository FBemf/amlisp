// The one that makes the code from the tree.
package codegen

import (
        "lexparse"
        "strconv"
)
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

func call(up chan assembly, ast lexparse.Ast, sym map[string]int) {

        if p := ast.Primitive(); p != nil {
                switch p.Type() {
                        case lexparse.litInt:
                                up <- assembly{"NEW", r0, 3, 0}
                                up <- assembly{"SET", r0, 0, 1} // refcount
                                up <- assembly{"SET", r0, 1, Type_integer} // type
                                up <- assembly{"SET", r0, 2, strconv.Atoi(p.Value())} // val
                        default:
                                fmt.Println("Unexpected primitive type!")
                        // TODO: add support for other primitives
                }
        }

        // Count members of s-exp
        members := 0
        for t := ast.Node(); t != nil; t = t.Next().Node() {
                if t.This() != nil {
                        members++
                }
        }

        // Make an env for it
        // an env has refcount, type, length, saved pc, return loc, parent env, pointers

        up <- assembly{"NEW", r2, members+6, 0}
        up <- assembly{"SETID", r0, 0, r2}
        up <- assembly{"SETI", r2, 0, 1}
        up <- assembly{"SETI", r2, 1, Type_environment}
        up <- assembly{"SETI", r2, 2, members + 4}

        // SETP is like SET, but arg3 is a loc holding the number you want to set it to.
        up <- assembly{"SETID", r2, 4, r0}
        up <- assembly{"SETID", r2, 5, r1}
        argCode := make([]chan assembly, members)
        for m := 0; m < members; m++ {
                argCode[m] = make(chan assembly, 100)
        for m := 0; m < members; m++ {
                ast = ast.Node().Next()
                go call(argCode[m], ast.Node.This(), copySym(sym))
        }
        for m := 0; m < members; m++ {
                up <- assembly{"SETD", r0, r2, 6+m}
                up <- assembly{"SETD", r1, r2, 0}
                for {
                        if a, b := <-argCode[m]; b {
                                up <- a
                        } else {
                                break
                        }
                }
        }

        //up <- assembly{"REMEMBER-JUMP", r2, 3, sym[whatever]} // saves pc+1 to [r2]+3
        // TODO: work out how we're going to handle the symbol table
        // We may have to make it a data structure

        // When we regain control, the return value is set
        // decrement the refcount for this env
        // if we hit zero, recursively do so for the pointers in this env

        // This block of code ascends the registers to the previous environment
        up <- assembly{"SETD", r2, r1, 0}
        up <- assembly{"SETD", r1, r2, 5}
        up <- assembly{"SETD", r1, r1, 0}
        up <- assembly{"SETD", r0, r2, 4}
        up <- assembly{"SETD", r0, r0, 0}

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
)

const (
        r1 = iota
        r2
        r3
        r4
        r5
)

func copySym(sym map[string]int) (out map[string]int) {
        out := make(map[string]int)
        for j, k := range(sym) {
                out[j] = k

        }
        return
}
