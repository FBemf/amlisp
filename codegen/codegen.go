package codegen

import (
        "../lexparse"
        "sync"
        "fmt"
)

func GenAssembly(ast lexparse.Ast) []Assembly {
        counter := makeCounter(0)
        code := make([]Assembly, 0, 100)
        sym := &safeSym{make(map[string]int), sync.Mutex{}}

        // Symbol table of builtin funcs
        builtins = map[string]string{
                "SYMBOL-QUOTE": "internal-quote",
                "add": "+",
                "subtract": "-",
                "cons": "cons",
                "car": "car",
                "cdr": "cdr",
                "empty": "empty",
                "if": "if",
                "define": "define",
                "FUNCTION": "func",
                "FINISHFUNC": "finishfunc"}

        // This declares the internal functions
        boilerplate := make(chan Assembly)
        go defaultFuncs(boilerplate, counter, sym)

        // This compiles the ast
        uparr := make([]chan Assembly, 0, 40)
        for i := 0; ast.This().Node() != nil; i++ {
                uparr = append(uparr, make(chan Assembly, 100))
                go call(uparr[i], ast.This(), counter, sym, false)
                ast = ast.Next()
        }

        // TODO: Initialize r0 and r1 properly
        // Create a top-level environment so definitions can happen
        code = append(code, Assembly{"SET-LITERAL", 0, 0, 0})
        code = append(code, Assembly{"SET-LITERAL", r0, 0, 0})
        code = append(code, Assembly{"SET-LITERAL", r1, 0, 0})
        code = append(code, Assembly{"SET-LITERAL", r2, 0, 0})
        up <- Assembly{"NEW ENV _f", 0, 0, 0}
        up <- Assembly{"NEW", r2, 7, 0}
        up <- Assembly{"SET-INDEXED", r2, 0, 0}
        up <- Assembly{"SET-INDEXED", r2, 1, Type_environment}
        up <- Assembly{"SET-INDEXED", r2, 2, 0}
        up <- Assembly{"COPY-INDEXED", r2, 4, 0}
        up <- Assembly{"COPY-INDEXED", r2, 5, 0} // Assumes r1 is return env
        up <- Assembly{"SET-INDEXED", r2, 6, 0}
        up <- Assemlby{"COPY-ADD", r1, r2, 0}

        // This unchannels all the compiled stuff
        for a, b := <-boilerplate; b; a, b = <-boilerplate {
                code = append(code, a)
        }

        for _, c := range uparr {
                for a, b := <-c; b; a, b = <-c {
                        code = append(code, a)
                }
        }
        fmt.Print(sym)
        return code
}