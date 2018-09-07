// The one that makes the code from the tree.
package codegen

import "lexparse"
import "fmt"

func CodeGen(ast lexparse.Ast) {
        fmt.Print("")
}

func callFunc(up chan assembly, ast lexparse.Asti, sym map[string]int) {
        asc := make(chan assembly)
        /*
                Requirements:
                Get a function location
                Get a parent env
                Make your own env
                Store the pc
                Hop the pc
                Go nuts
        */

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
