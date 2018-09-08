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

func call(up chan assembly, ast lexparse.Ast, sym map[string]int) {
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

        if p := ast.Primitive(); p != nil {
                switch p.Type() {
                        case lexparse.litInt:
                                up <- assembly{"NEW", r0, 3, 0}
                                up <- assembly{"SET", r0, 0, 1} // refcount
                                up <- assembly{"SET", r0, 1, Type_integer} // type
                                up <- assembly{"SET", r0, 2, p.Value()} // val
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

        // NEW takes an address arg1, and an int arg2.
        // The loc pointed to by arg1 is a pointer to another location, which is
        // set to be a pointer to a newly-allocated block of memory arg2 locations long.
        up <- assembly{"NEW", r0, members, 0}

        // SET takes the location pointed to by the memory loc arg1, adds arg2 to it,
        // and sets it to arg3
        up <- assembly{"SET", r0, 0, 1}
        up <- assembly{"SET", r0, 1, Type_environment}
        up <- assembly{"SET", r0, 2, members + 4}
        // Skip this one for later

        // SETP is like SET, but arg3 is a loc holding the number you want to set it to.
        up <- assembly{"SETP", r0, 4, r0}
        up <- assembly{"SETP", r0, 5, r1}
        argCode := make([]chan assembly, members)
        for m := 0; m < members; m++ {
                ast = ast.Node().Next()
                go call(argCode[m], ast.Node.This(), copySym(sym))
        }
        for m := 0; m < members; m++ {
                up <- assembly{"SETDP", r1, // set r1 to 
                up <- assembly{
                for {
                        if a, b := <-argCode[m]; b {
                                up <- a
                        } else {
                                break
                        }
                }
        }


        // Either set each arg from a raw or go callFunc it
        // Set oldpc
        // jump
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
