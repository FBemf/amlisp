package codegen

import (
        "./lexparse"
)

func GenAssebly(ast *lexparse.Ast) []assembly {
        counter := makeCounter(0)
        code := make([]assembly, 0, 100)
        sym = safeSym{make(map[string]int), sync.Mutex{})

        // This declares the internal functions
        boilerplate := make(chan assembly)
        go defaultFuncs(boilerplate, counter)

        // This compiles the ast
        uparr := make([]chan assembly, 40)
        for i := 0; ast.Node().This() != nil; i++ {
                go call(uparr[i], ast.Node().This(), counter, sym, false)
        }

        // This unchannels all the compiled stuff
        for a, b := <-boilerplate; b; a, b := <-boilerplate {
                code = append(code, a)
        }
        for _, c := range uparr {
                for a, b := <-c; b; a, b := <-c {
                        code = append(code, a)
                }
        }
        return code
}
