package codegen

import (
	"../lexparse"
	"fmt"
	"sync"
)

func GenAssembly(ast lexparse.Ast) []Assembly {
	counter := makeCounter(0)
	code := make([]Assembly, 0, 100)
	sym := &safeSym{make(map[string]int), sync.Mutex{}}

	// Symbol table of builtin funcs
	builtins = map[string]string{
		"SYMBOL-QUOTE": "iquote",
		"add":          "+",
		//"subtract":     "sub",
		//"cons":         "cons",
		//"car":          "car",
		//"cdr":          "cdr",
		//"empty":        "empty",
		//"if":           "if",
		"define":       "define",
		"FUNCTION":     "func",
		"FINISHFUNC":   "_finishfunc",
		"DEFUN":	"_defun" }

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

	// Create a top-level environment so definitions can happen
	code = append(code, Assembly{"SET-LITERAL", 0, 0, 0})
	code = append(code, Assembly{"SET-LITERAL", r0, 0, 0})
	code = append(code, Assembly{"SET-LITERAL", r1, 0, 0})
	code = append(code, Assembly{"SET-LITERAL", r2, 0, 0})
	code = append(code, Assembly{"NEW ENV _f", 0, 0, 0})
	code = append(code, Assembly{"NEW", r2, 7, 0})
	code = append(code, Assembly{"SET-INDEXED", r2, 0, 0})
	code = append(code, Assembly{"SET-INDEXED", r2, 1, Type_environment})
	code = append(code, Assembly{"SET-INDEXED", r2, 2, 0})
	code = append(code, Assembly{"COPY-INDEXED", r2, 4, 0})
	code = append(code, Assembly{"COPY-INDEXED", r2, 5, 0}) // Assumes r1 is return env
	code = append(code, Assembly{"SET-INDEXED", r2, 6, 0})
	code = append(code, Assembly{"COPY-ADD", r1, r2, 0})

	// This unchannels all the compiled stuff
	for a, b := <-boilerplate; b; a, b = <-boilerplate {
		code = append(code, a)
	}

	for _, c := range uparr {
		for a, b := <-c; b; a, b = <-c {
			code = append(code, a)
		}
	}
	//fmt.Print(sym)
	_ = fmt.Print
	return code
}
