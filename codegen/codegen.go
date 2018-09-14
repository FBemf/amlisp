package codegen

import (
        "../lexparse"
        "sync"
        "fmt"
)

func GenAssembly(ast lexparse.Ast) []assembly {
        counter := makeCounter(0)
        code := make([]assembly, 0, 100)
        sym := &safeSym{make(map[string]int), sync.Mutex{}}

        // Symbol table of builtin funcs
        builtins = map[string]string{
                "SYMBOL-QUOTE": "internal-quote",
                "+": "+",
                "-": "-",
                "cons": "cons",
                "car": "car",
                "cdr": "cdr",
                "empty": "empty",
                "if": "if",
                "define": "define",
                "FUNCTION": "func"}

        // This declares the internal functions
        boilerplate := make(chan assembly)
        go defaultFuncs(boilerplate, counter, sym)

        // This compiles the ast
        uparr := make([]chan assembly, 0, 40)
        for i := 0; ast.This().Node() != nil; i++ {
                uparr = append(uparr, make(chan assembly, 100))
                go call(uparr[i], ast.This(), counter, sym, false)
                ast = ast.Next()
        }
        fmt.Println("hYE")

        // This unchannels all the compiled stuff
        for a, b := <-boilerplate; b; a, b = <-boilerplate {
                code = append(code, a)
        }
        fmt.Println("WOAH!")
        for _, c := range uparr {
                for a, b := <-c; b; a, b = <-c {
                        code = append(code, a)
                }
        }
        return code
}
