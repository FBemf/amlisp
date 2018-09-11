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
