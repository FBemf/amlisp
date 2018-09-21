package codegen

// Makes assembly to iterate through the symtab
// in [env] looking for the symbol target
// and puts the output into mem. Throws an exception
// if it isn't there.
func querySymtab(up chan Assembly, mem int, mem2 int, env int, target int, counter func() int) {
        loop := counter()
        end := counter()
        bad := counter()
        up <- Assembly{"DEREF", mem, env, 6} // set mem to table
        up <- Assembly{"LABEL", loop, 0, 0}
       // if symtab[2] points to x st x[2] == [target]
        up <- Assembly{"DEREF", mem2, mem, 2}
        up <- Assembly{"COPY-ADD", mem2, mem2, 2}
        up <- Assembly{"JUMP-LABEL-IF-EQ", end, target, mem2}
        up <- Assembly{"DEREF", mem, mem, 2}
        up <- Assembly{"JUMP-LABEL-IF-IS", bad, mem, 0}
        up <- Assembly{"JUMP-LABEL", loop, 0, 0}
        up <- Assembly{"LABEL", bad, 0, 0}
        up <- Assembly{"EXCEPTION", Ex_undefined, 0, 0}
        up <- Assembly{"LABEL", end, 0, 0}
        up <- Assembly{"DEREF", mem, mem, 1}
}
