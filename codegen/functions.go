package codegen

import "strconv"

func defaultFuncs(up chan Assembly, counter func() int, sym *safeSym) {

        // TODO: Gotta a) enclose this in a jump, b) put the symbols into the symbol table

        // basic outline of a builtin func:
        // at start, r0, r1, r2 are all as normal.
        // at end, are expected to: return return value
        // and call finishfunc.
        // ^ That's all enclosed in a jump--after that, there's
        // code creating a closure that points to the function

        // Finish func
        endFinishFunc = counter()
        up <- Assembly{"JUMP-LABEL", endFinishFunc, 0, 0}
        up <- Assembly{"BOILERPLATE _f", 0, 0, 0}
        up <- Assembly{"LABEL", sym.getSymID(builtins["FINISHFUNC"], counter), 0, 0}
        up <- Assembly{"DEREF", r3, r0, 0}      // Grab return value
        up <- Assembly{"ADD1", r3, 0, 0}       // add to the refcount of the returned value
        up <- Assembly{"SUB1", r2, 0, 0}       // decrement current env refcount ([r2]--)

        up <- Assembly{"DEREF", r3, r2, 3}       // Grab the pc to return to

        up <- Assembly{"COPY-ADD", r5, r2, 0}       // env for dumpfunc

        // Ascend the registers to the previous environment
        up <- Assembly{"COPY-ADD", r2, r1, 0}
        up <- Assembly{"DEREF", r1, r2, 5}
        up <- Assembly{"DEREF", r0, r2, 4}

        // dumpfunc, here and now
        // Check if refcount is zero--jump past dumpfunc if it isn't
        skip_jump_de := counter()
        dump_end := counter()
        up <- Assembly{"COPY-ADD", r4, r5, 1}
        up <- Assembly{"JUMP-LABEL-IF-IS", skip_jump_de, r4, 0}
        up <- Assembly{"JUMP-LABEL", dump_end, 0, 0}
        up <- Assembly{"LABEL", skip_jump_de, 0, 0}

        // You can use registers 4 and up
        // Reminder: The dumpfunc ds looks like:
        //      [refcount] [type] [current ds] [next frame]

        // Walking into this, r5 is the env to be dumped

        // i) new data structure
        // ii) populate new data structure
        up <- Assembly{"NEW", r4, 4, 0}
        up <- Assembly{"SET-INDEXED", r4, 1, 0}
        up <- Assembly{"SET-INDEXED", r4, 2, Type_dump}
        up <- Assembly{"COPY-INDEXED", r4, 3, r5}
        up <- Assembly{"SET-INDEXED", r4, 4, 0}

        /* Review of set commands
                SET-LITERAL a b c ->    a = b
                SET-INDEXED a b c ->    [a] + b = c
                COPY-ADD a b c ->       a = [b] + c
                COPY-INDEXED a b c ->   [a] + b = [c]
                DEREF a b c ->          a = [[b] + c]
        */

        // iii) top of loop
        dump_start := counter()
        dump_continue := counter()
        up <- Assembly{"LABEL", dump_start, 0, 0}

        // iv) Set refcount of env to -1
        up <- Assembly{"COPY-ADD", r5, 0, -1}

        // v) Switch on type, either jump to "continue" or set start + length
        switch_end := counter()
        switch_type_int := counter()
        switch_type_env := counter()
        // TODO ... other types

        up <- Assembly{"COPY-ADD", r5, r5, 1}
        up <- Assembly{"JUMP-LABEL-IF-IS", switch_type_int, r5, Type_int}
        up <- Assembly{"JUMP-LABEL-IF-IS", switch_type_env, r5, Type_environment}
        // TODO ... other types

        up <- Assembly{"LABEL", switch_type_int, 0, 0} // int
        // labels for other non-pointer data types
        up <- Assembly{"JUMP-LABEL", dump_continue, 0, 0}

        // env
        up <- Assembly{"LABEL", switch_type_env, 0, 0}
        up <- Assembly{"COPY-ADD", r6, r5, 5} // first pointer is r5+6, minus one to get symtab too
        up <- Assembly{"DEREF", r5, r5, 1} // length
        up <- Assembly{"ADD", r5, r5, r6} // one after last pointer
        up <- Assembly{"JUMP-LABEL", switch_end, 0, 0}

        // TODO .. other pointer-set types
        // Have them all finish with r6 as the first ptr and r5 as the one
        // after the last ptr

        up <- Assembly{"LABEL", switch_end, 0, 0}

        // vi) top of loop 2
        rec_loop := counter()
        up <- Assembly{"LABEL", rec_loop, 0, 0}

        // vii) if at end, continue
        up <- Assembly{"JUMP-LABEL-IF-EQ", dump_continue, r5, r6}      // if-is compares a register to a literal
                                                                       // if-eq compares two registers

        // viii) otherwise, new ds frame
        // ix) populate data structure frame
        up <- Assembly{"NEW", r7, 4, 0}
        up <- Assembly{"SET-INDEXED", r7, 1, 0}
        up <- Assembly{"SET-INDEXED", r7, 2, Type_dump}
        up <- Assembly{"COPY-INDEXED", r7, 3, r6}
        up <- Assembly{"DEREF", r8, r4, 4}
        up <- Assembly{"COPY-INDEXED", r7, 4, r8}

        // x) stick ds frame into list
        up <- Assembly{"COPY-INDEXED", r4, 4, r7}

        // xi) next pointer
        up <- Assembly{"COPY-ADD", r5, r5, 1}

        // xii) continue loop 2
        up <- Assembly{"JUMP-LABEL", rec_loop, 0, 0}

        // xiii) set current ds frame refcount to -1 (this is where "continue" is)
        up <- Assembly{"SET-INDEXED", r4, 0, -1}

        // xiv) if next frame zero, exit
        up <- Assembly{"COPY-ADD", r4, r4, 3}
        up <- Assembly{"JUMP-LABEL-IF-IS", dump_end, r4, 0}

        // xv) otherwise, next frame, continue loop 1
        up <- Assembly{"DEREF", r4, r4, 3}
        up <- Assembly{"JUMP-LABEL", dump_start, 0, 0}
        up <- Assembly{"LABEL", dump_end, 0, 0}
        up <- Assembly{"JUMP", r3, 0, 0}
        up <- Assembly{"BOILERPLATE END _f", 0, 0, 0}
        up <- Assembly{"LABEL", endFinishFunc, 0, 0}

        // '+' func
        endAddFunc := counter()
        up <- Assembly{"JUMP-LABEL", endAddFunc, 0, 0}
        up <- Assembly{"LABEL", sym.getSymID(builtins["add"], counter), 0, 0}
        up <- Assembly{"NEW", r3, 3, 0} // new int
        up <- Assembly{"SET-INDEXED", r3, 0, 1}
        up <- Assembly{"SET-INDEXED", r3, 1, Type_int}

        // grab args from symtab
        querySymtab(up, r4, r5, r2, sym.getSymID(_add_arg_0, counter), counter)
        querySymtab(up, r5, r6, r2, sym.getSymID(_add_arg_1, counter), counter)

        //up <- Assembly{"DEREF", r4, r2, 8}   // first arg
        //up <- Assembly{"DEREF", r5, r2, 9}   // second arg
        up <- Assembly{"ADD", r4, r4, r5}       // new: [r4] = [r4] + [r5]
        up <- Assembly{"COPY-INDEXED", r3, 2, r4}
        up <- Assembly{"DEREF", r0, r3, 0}
        up <- Assembly{"JUMP-LABEL", sym.getSymID(builtins["FINISHFUNC"], counter), 0, 0}
        up <- Assembly{"LABEL", endAddFunc, 0, 0}
        // Create closure
        up <- Assembly{"NEW", r3, 6, 0}
        up <- Assembly{"SET-INDEXED", r3, 0, 1}
        up <- Assembly{"SET-INDEXED", r3, 1, Type_closure}
        up <- Assembly{"SET-LABEL-INDEXED", r3, 2, sym.getSymID(builtins["add"], counter)}      // sets a cell to be a label.
        up <- Assembly{"COPY-INDEXED", r3, 3, r2}
        up <- Assembly{"COPY-INDEXED", r3, 4, 2}
        for i := 0; i < 2; i++ {
                up <- Assembly{"NEW", r4, 3, 0}
                up <- Assembly{"SET-INDEXED", r4, 0, 1}
                up <- Assembly{"SET-INDEXED", r4, 1, Type_symbol}
                up <- Assembly{"SET-INDEXED", r4, 2, sym.getSymID("_add_arg_"+strconv.Itoa(i) , counter)}
                up <- Assembly{"COPY-INDEXED", r3, 5+i, r4}
        }

        // Add to symtab
        /*up <- Assembly{"DEREF", r4, r2, 6}
        up <- Assembly{"COPY-INDEXED", r3, 3, r4}
        up <- Assembly{"COPY-INDEXED", r2, 6, r3}*/
        addToSymtab(up, r4, r3, r2)

        close(up)
}
