package codegen

func defaultFuncs(up chan assembly, counter func() int) {

        // basic outline of a builtin func:
        // at start, r0, r1, r2 are all as normal.
        // at end, are expected to:
        //      a) return value to [r0]
        //      b) add to the refcount of the struct being returned
        //      c) decrement its own
        //      d) ascend its r0, r1, r2 registers to the parent env
        //      d) set off dumpfunc
        //      e) jump back to the old pc
        // All this is enclosed in a big JUMP statement
        // After the jump statement, we make a closure for ourselves and
        // stick that in the symbol table

        // All of this is so similar to what "func" already does
        // consider turning func to some extent into a builtin func
        // instead of just a flag

        // Template:
        up <- assembly{"DEREF", r3, r2, members+6}      // Grab return value -- not generic
        up <- assembly{"COPY-INDEXED", r0, 0, r3}      // return
        up <- assembly{"JUMP-LABEL", finishfunc, 0, 0}

        // Finish func boilerplate
        up <- assembly{"LABEL", finishfunc, 0, 0}
        up <- assembly{"DEREF", r3, r0, 0}      // Grab return value -- not generic
        up <- assembly{"ADD1", r3, 0, 0}       // add to the refcount of the returned value
        up <- assembly{"SUB1", r2, 0, 0}       // decrement current env refcount ([r2]--)

        up <- assembly{"DEREF", r3, r2, 3}       // Grab the pc to return to - an arg for dumpfunc

        up <- assembly{"COPY-ADD", r4, r2, 0}       // argument for dumpfunc

        // Ascend the registers to the previous environment
        up <- assembly{"COPY-ADD", r2, r1, 0}
        up <- assembly{"DEREF", r1, r2, 5}
        up <- assembly{"DEREF", r0, r2, 4}

        // dumpfunc, here and now
        // Check if refcount is zero--jump past dumpfunc if it isn't

        // i) new data structure
        // ii) populate new data structure
        // iii) top of loop
        // iv) Set refcount of env to -1
        // v) Switch on type, either jump to "continue" or set start + length
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
}
