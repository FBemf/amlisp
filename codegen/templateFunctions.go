package codegen

// Makes assembly to iterate through the symtab
// in [env] looking for the symbol target
// (which is an id and not a register or anything)
// and puts the output into mem. Throws an exception
// if it isn't there.
func querySymtab(up chan Assembly, mem int, mem2 int, env int, target int, counter func() int) {
        loop := counter()
        end := counter()
        bad := counter()
        up <- Assembly{"OLE!",0,0,0}
        up <- Assembly{"DEREF", mem, env, 6} // set mem to table
        up <- Assembly{"LABEL", loop, 0, 0}
        // if symtab[2] points to x st x[2] == [target]
        up <- Assembly{"DEREF", mem2, mem, 2}
        up <- Assembly{"COPY-ADD", mem2, mem2, 2}
        up <- Assembly{"DEREF", mem2, mem2, 0}
        up <- Assembly{"JUMP-LABEL-IF-IS", end, mem2, target}
        up <- Assembly{"DEREF", mem, mem, 4}
        up <- Assembly{"JUMP-LABEL-IF-IS", bad, mem, 0}
        up <- Assembly{"JUMP-LABEL", loop, 0, 0}
        up <- Assembly{"LABEL", bad, 0, 0}
        up <- Assembly{"EXCEPTION", Ex_undefined, 0, 0}
        up <- Assembly{"LABEL", end, 0, 0}
        up <- Assembly{"DEREF", mem, mem, 3}
}

// mem mem2 are empty registers
// symbol is the id (not a register) of the symbol
// target is a register holding the primitive
// env is a register with a pointer to a frame
func addToSymtab(up chan Assembly, mem int, mem2 int, symbol int, target int, env int) {

        // symtab: [refcount] [Type symtab] [id] [loc] [next]
        // id: [refcount] [type symbol]

        // NB: This function starts the refcount at 1.
        up <- Assembly{"CHECK THIS OUT _f", 0, 0, 0}

        // Make symbol object
        up <- Assembly{"NEW", mem2, 3, 0}
        up <- Assembly{"SET-INDEXED", mem2, 0, 1}
        up <- Assembly{"SET-INDEXED", mem2, 1, Type_symbol}
        up <- Assembly{"SET-INDEXED", mem2, 2, symbol}

        // Makes new symtab frame
        up <- Assembly{"NEW", mem, 5, 0}
        up <- Assembly{"SET-INDEXED", mem, 0, 1}
        up <- Assembly{"SET-INDEXED", mem, 1, Type_symtab}
        up <- Assembly{"COPY-INDEXED", mem, 2, mem2}
        up <- Assembly{"COPY-INDEXED", mem, 3, target}  // this is a register

        up <- Assembly{"DEREF", mem2, env, 6}
        up <- Assembly{"COPY-INDEXED", mem, 4, mem2}
        up <- Assembly{"COPY-INDEXED", env, 6, mem}
}

// mem mem2 are empty registers
// symbol is a register holding the symbol
// target is a register holding the primitive
// env is a register with a pointer to a frame
func addToSymtabRegister(up chan Assembly, mem int, mem2 int, symbol int, target int, env int) {

        // symtab: [refcount] [Type symtab] [id] [loc] [next]
        // id: [refcount] [type symbol]

        // NB: This function starts the refcount at 1.
        up <- Assembly{"CHECK THIS OUT _f", 0, 0, 0}

        // Make symbol object
        up <- Assembly{"NEW", mem2, 3, 0}
        up <- Assembly{"SET-INDEXED", mem2, 0, 1}
        up <- Assembly{"SET-INDEXED", mem2, 1, Type_symbol}
        up <- Assembly{"COPY-INDEXED", mem2, 2, symbol}

        // Makes new symtab frame
        up <- Assembly{"NEW", mem, 5, 0}
        up <- Assembly{"SET-INDEXED", mem, 0, 1}
        up <- Assembly{"SET-INDEXED", mem, 1, Type_symtab}
        up <- Assembly{"COPY-INDEXED", mem, 2, mem2}
        up <- Assembly{"COPY-INDEXED", mem, 3, target}  // this is a register

        up <- Assembly{"DEREF", mem2, env, 6}
        up <- Assembly{"COPY-INDEXED", mem, 4, mem2}
        up <- Assembly{"COPY-INDEXED", env, 6, mem}
}
