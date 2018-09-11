package codegen

import (
        "lexparse"
        "strconv"
        "sync"
)

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
