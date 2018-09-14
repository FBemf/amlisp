package codegen

func defaultFuncs(up chan assembly, counter func() int, sym *safeSym) {

        // basic outline of a builtin func:
        // at start, r0, r1, r2 are all as normal.
        // at end, are expected to: return return value
        // and call finishfunc.
        // ^ That's all enclosed in a jump--after that, there's
        // code creating a closure that points to the function

        // Finish func
        up <- assembly{"LABEL", sym.getSymID(builtins["finishfunc"], counter), 0, 0}
        up <- assembly{"DEREF", r3, r0, 0}      // Grab return value -- not generic
        up <- assembly{"ADD1", r3, 0, 0}       // add to the refcount of the returned value
        up <- assembly{"SUB1", r2, 0, 0}       // decrement current env refcount ([r2]--)

        up <- assembly{"DEREF", r3, r2, 3}       // Grab the pc to return to

        up <- assembly{"COPY-ADD", r5, r2, 0}       // env for dumpfunc

        // Ascend the registers to the previous environment
        up <- assembly{"COPY-ADD", r2, r1, 0}
        up <- assembly{"DEREF", r1, r2, 5}
        up <- assembly{"DEREF", r0, r2, 4}

        // dumpfunc, here and now
        // Check if refcount is zero--jump past dumpfunc if it isn't
        skip_jump_de := counter()
        dump_end := counter()
        up <- assembly{"COPY-ADD", r4, r5, 1}
        up <- assembly{"JUMP-LABEL-IF-IS", skip_jump_de, r4, 0}
        up <- assembly{"JUMP-LABEL", dump_end, 0, 0}
        up <- assembly{"LABEL", skip_jump_de, 0, 0}

        // You can use registers 4 and up
        // Reminder: The dumpfunc ds looks like:
        //      [refcount] [type] [current ds] [next frame]

        // Walking into this, r5 is the env to b dumped

        // i) new data structure
        // ii) populate new data structure
        up <- assembly{"NEW", r4, 4, 0}
        up <- assembly{"SET-INDEXED", r4, 1, 0}
        up <- assembly{"SET-INDEXED", r4, 2, Type_dump}
        up <- assembly{"COPY-INDEXED", r4, 3, r5}
        up <- assembly{"SET-INDEXED", r4, 4, 0}

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
        up <- assembly{"LABEL", dump_start, 0, 0}

        // iv) Set refcount of env to -1
        up <- assembly{"COPY-ADD", r5, 0, -1}

        // v) Switch on type, either jump to "continue" or set start + length
        switch_end := counter()
        switch_type_int := counter()
        switch_type_env := counter()
        // TODO ... other types

        up <- assembly{"COPY-ADD", r5, r5, 1}
        up <- assembly{"JUMP-LABEL-IF-IS", switch_type_int, r5, Type_int}
        up <- assembly{"JUMP-LABEL-IF-IS", switch_type_env, r5, Type_environment}
        // TODO ... other types

        up <- assembly{"LABEL", switch_type_int, 0, 0} // int
        // labels for other non-pointer data types
        up <- assembly{"JUMP-LABEL", dump_continue, 0, 0}

        // env
        up <- assembly{"LABEL", switch_type_env, 0, 0}
        up <- assembly{"COPY-INDEXED", r6, r5, 6} // first pointer
        up <- assembly{"DEREF", r5, r5, 1} // length
        up <- assembly{"ADD", r5, r5, r6} // one after last pointer
        up <- assembly{"JUMP-LABEL", switch_end, 0, 0}

        // TODO .. other pointer-set types
        // Have them all finish with r6 as the first ptr and r5 as the one
        // after the last ptr

        up <- assembly{"LABEL", switch_end, 0, 0}

        // vi) top of loop 2
        rec_loop := counter()
        up <- assembly{"LABEL", rec_loop, 0, 0}

        // vii) if at end, continue
        up <- assembly{"JUMP-LABEL-IF-EQ", dump_continue, r5, r6}      // if-is compares a register to a literal
                                                                       // if-eq compares two registers

        // viii) otherwise, new ds frame
        // ix) populate data structure frame
        up <- assembly{"NEW", r7, 4, 0}
        up <- assembly{"SET-INDEXED", r7, 1, 0}
        up <- assembly{"SET-INDEXED", r7, 2, Type_dump}
        up <- assembly{"COPY-INDEXED", r7, 3, r6}
        up <- assembly{"DEREF", r8, r4, 4}
        up <- assembly{"COPY-INDEXED", r7, 4, r8}

        // x) stick ds frame into list
        up <- assembly{"COPY-INDEXED", r4, 4, r7}

        // xi) next pointer
        up <- assembly{"COPY-ADD", r5, r5, 1}

        // xii) continue loop 2
        up <- assembly{"JUMP-LABEL", rec_loop, 0, 0}

        // xiii) set current ds frame refcount to -1 (this is where "continue" is)
        up <- assembly{"SET-INDEXED", r4, 0, -1}

        // xiv) if next frame zero, exit
        up <- assembly{"COPY-ADD", r4, r4, 3}
        up <- assembly{"JUMP-LABEL-IF-IS", dump_end, r4, 0}

        // xv) otherwise, next frame, continue loop 1
        up <- assembly{"DEREF", r4, r4, 3}
        up <- assembly{"JUMP-LABEL", dump_start, 0, 0}
        up <- assembly{"LABEL", dump_end, 0, 0}
        up <- assembly{"JUMP", r3, 0, 0}
}
