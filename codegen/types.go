package codegen

import (
	"fmt"
	"sync"
)

type safeSym struct {
	table map[string]int
	mutex sync.Mutex
}

var builtins map[string]string

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

type Assembly struct {
	Command string
	Arg1    int
	Arg2    int
	Arg3    int
}

func Disassemble(a Assembly) string {
	return fmt.Sprintf("%s %d %d %d", a.Command, a.Arg1, a.Arg2, a.Arg3)
}

const (
	Type_environment = iota
	Type_closure
	Type_dump
	Type_symtab
	Type_cons
	Type_vector
	Type_int
	Type_symbol
)

const (
	Ex_undefined = iota
)

const (
	r0 = iota + 2
	r1
	r2
	r3
	r4
	r5
	r6
	r7
	r8
	r9
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
