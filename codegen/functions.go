package codegen

func defaultFuncs(up chan assembly, counter func() int) {

        // basic outline of a builtin func:
        // at start, r0, r1, r2 are all as normal.
        // at end, are expected to: return return value
        // and call finishfunc.
        // ^ That's all enclosed in a jump--after that, there's
        // code creating a closure that points to the function

        // Finish func
        up <- assembly{"LABEL", finishfunc, 0, 0}
        up <- assembly{"DEREF", r3, r0, 0}      // Grab return value -- not generic
        up <- assembly{"ADD1", r3, 0, 0}       // add to the refcount of the returned value
        up <- assembly{"SUB1", r2, 0, 0}       // decrement current env refcount ([r2]--)

        up <- assembly{"DEREF", r3, r2, 3}       // Grab the pc to return to - an arg for dumpfunc

        up <- assembly{"COPY-ADD", r5, r2, 0}       // argument for dumpfunc

        // Ascend the registers to the previous environment
        up <- assembly{"COPY-ADD", r2, r1, 0}
        up <- assembly{"DEREF", r1, r2, 5}
        up <- assembly{"DEREF", r0, r2, 4}

        // dumpfunc, here and now
        // Check if refcount is zero--jump past dumpfunc if it isn't
        // TODO
        dump_end := counter()
        up <- assembly{"COPY-ADD", r4, r5, 1}
        up <- assembly{"JUMP-LABEL-IF-IS", dump_end, r4, 0}

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
        top := counter()
        end := counter()
        up <- assembly{"LABEL", top, 0, 0}

        // iv) Set refcount of env to -1
        up <- assembly{"COPY-ADD", r5, 0, -1}

        // v) Switch on type, either jump to "continue" or set start + length
        switch_end := counter()
        switch_type_int := counter()
        switch_type_env := counter()
        // TODO ... other types

        up <- assembly{"COPY-ADD", r5, r5, 1}
        up <- assembly("JUMP-LABEL-IF-IS", switch_type_int, r5, Type_int}
        up <- assembly("JUMP-LABEL-IF-IS", switch_type_env, r5, Type_env}
        // TODO ... other types

        up <- assembly{"LABEL", switch_type_int, 0, 0}
        // labels for other non-pointer data types
        up <- assembly{"JUMP-LABEL", dump_end, 0, 0}

        up <- assembly{"LABEL", switch_type_env, 0, 0}
        up <- assemlby{"COPY-INDEXED", r5, r5, 1} // length
        up <- assembly{"COPY-INDEXED", r6, r5, 5} // first pointer
        // TODO: turn this into current-pointer & last-pointer
        up <- assembly{"JUMP-LABEL", switch_end, 0, 0}

        // TODO .. other pointer-set types

        up <- assemblyl{"LABEL", switch_end, 0, 0}

        // vi) top of loop 2
        // vii) if length is zero, break
        // viii) otherwise, new ds frame
        // ix) populate data structure frame
        // x) stick ds frame into list
        // xi) decrement length, increment start
        // xii) continue loop 2
        // xiii) set current ds frame refcount to -1 (this is where "continue" is)
        // xiv) if next frame zero, exit
        // xv) otherwise, continue loop 1
        up <- assembly{"LABEL", dump_end, 0, 0}
}
