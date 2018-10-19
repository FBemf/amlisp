package codegen

import "strconv"

func defaultFuncs(up chan Assembly, counter func() int, sym *safeSym) {

	// All builtin funcs follow these rules:
	//      - At the start, r0 is the return location,
	//        r1 is the parent environment, r2 is the
	//        current environment.
	//      - At the end, they are expected to return
	//        the return value to [r0] and call finishfunc
	//      - The above is the runtime, which is enclosed inside
	//        a jump. The definition is not hidden away, and
	//        creates a closure and a symbol in the symbol table

	// Finish func
	// This one's a big deal. It ascends the registers
	// back to the previous environment (so that r2 is now
	// the parent environment and r1 is now *that* environment's
	// parent environment) and garbage-collects the current env
	// if the last reference is disappearing.
	endFinishFunc := counter()
	up <- Assembly{"BOILERPLATE-FOR_FF _f", 0, 0, 0}
	up <- Assembly{"JUMP-LABEL", endFinishFunc, 0, 0}
	up <- Assembly{"LABEL", sym.getSymID(builtins["FINISHFUNC"], counter), 0, 0}
	up <- Assembly{"FINFUN _f", 0, 0, 0}
	up <- Assembly{"DEREF", r3, r0, 0} // Grab return value
	up <- Assembly{"ADD1", r3, 0, 0}   // Add to the refcount of the thing being returned

	fakeEnvLabel := counter()
	// a func with no env (hardcoded & optimized) shouldn't have its env
	// garbage-collected
	up <- Assembly{"JUMP-LABEL-IF-EQ", fakeEnvLabel, r2, r1}
	up <- Assembly{"SUB1", r2, 0, 0} // Decrement current environment's refcount
	up <- Assembly{"LABEL", fakeEnvLabel, 0, 0}

	up <- Assembly{"DEREF", r3, r2, 3} // Find the PC to return to

	up <- Assembly{"COPY-ADD", r5, r2, 0} // Take a reference to the current environment
	// so it can be GC'd

	// Ascend the registers to the previous environment

	up <- Assembly{"COPY-ADD", r2, r1, 0}
	up <- Assembly{"DEREF", r1, r2, 5}
	up <- Assembly{"DEREF", r0, r2, 4}

	// Garbage Collector
	//
	// This part's important. Garbage collection is managed by reference counting.
	// The first entiry in any data structure is it's ref count, or the number of
	// pointers to it that exist. When that number reaches zero, it's dead.
	// Now, when an environment goes out of scope, it gets garbage-collected.
	// This means that it is checked to see if its ref count is zero, and if it is,
	// then its ref count is set to -1, and all the data structures pointed to by
	// the environment are garbage-collected. Then, next time memory is allocated,
	// any block the VM comes across that has a refcount of -1 is deallocated.

	// Check if refcount is zero--jump past GC if it isn't
	skip_jump_de := counter()
	dump_end := counter()
	up <- Assembly{"DEREF", r4, r5, 0}
	up <- Assembly{"JUMP-LABEL-IF-IS", skip_jump_de, r4, 0}
	up <- Assembly{"JUMP-LABEL", dump_end, 0, 0}
	up <- Assembly{"LABEL", skip_jump_de, 0, 0}

	// You can use registers 4 and up
	// Reminder: The dumpfunc ds looks like:
	//      [refcount] [type] [current ds] [next frame]
	// The GC has its own data structure. It's a queue,
	// so that nested data structures can be recursively
	// deallocated.
	// i) Refcount: By necessity. It gets set to -1 when
	//      the GC is finished with it.
	// ii) Type: Type_dump, not that it matters, since
	//      there's no reason this should ever see the light
	//      of day.
	// iii) Current data structure: A pointer to the data
	//      structure being garbage-collected in this frame.
	// iv)  Next frame: A pointer to the next frame in the
	//      stack.

	// Walking into this, r5 is the env to be dumped

	// i) new data structure
	// ii) populate new data structure
	up <- Assembly{"NEW", r4, 4, 0}
	up <- Assembly{"SET-INDEXED", r4, 0, 0}
	up <- Assembly{"SET-INDEXED", r4, 1, Type_dump}
	up <- Assembly{"COPY-INDEXED", r4, 2, r5}
	up <- Assembly{"SET-INDEXED", r4, 3, 0}

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
	// THIS STEP IS WRONG. You have to do this AFTER
	// you allocate the new dump frames or the current env
	// will be overwritten!
	//up <- Assembly{"SET-INDEXED", r5, 0, -1}
	// TODO remove this shit

	// v) Switch on type, either jump to "continue" or set start + length
	switch_end := counter()
	switch_type_int := counter()
	switch_type_env := counter()
	switch_type_symtab := counter()
	// TODO ... other types

	up <- Assembly{"DEREF", r6, r5, 1}
	up <- Assembly{"JUMP-LABEL-IF-IS", switch_type_int, r6, Type_int}
	up <- Assembly{"JUMP-LABEL-IF-IS", switch_type_env, r6, Type_environment}
	up <- Assembly{"JUMP-LABEL-IF-IS", switch_type_symtab, r6, Type_symtab}
	// TODO ... other types
	up <- Assembly{"EXCEPTION", 456, 0, 0}

	up <- Assembly{"LABEL", switch_type_int, 0, 0} // int
	// labels for other non-pointer data types
	up <- Assembly{"JUMP-LABEL", dump_continue, 0, 0}

	// env -- at the start of all of these, r5 is pointing to the "type" box
	up <- Assembly{"LABEL", switch_type_env, 0, 0}
	up <- Assembly{"COPY-ADD", r6, r5, 6} // first pointer is r5+7, minus one to get symtab too
	up <- Assembly{"DEREF", r5, r5, 2}    // length
	up <- Assembly{"ADD", r5, r5, r6}     // last pointer
	up <- Assembly{"COPY-ADD", r5, r5, 1}     // one after last pointer
	up <- Assembly{"JUMP-LABEL", switch_end, 0, 0}

	// symtab
	up <- Assembly{"LABEL", switch_type_symtab, 0, 0}
	up <- Assembly{"COPY-ADD", r6, r5, 2} // first pointer is r5+2
	up <- Assembly{"COPY-ADD", r5, r5, 5} // one after last pointer
	up <- Assembly{"JUMP-LABEL", switch_end, 0, 0}

	// TODO .. other pointer-set types
	// Have them all finish with r6 as the first ptr and r5 as the one
	// after the last ptr

	up <- Assembly{"LABEL", switch_end, 0, 0}

	// vi) top of loop 2
	rec_loop := counter()
	up <- Assembly{"LABEL", rec_loop, 0, 0}

	// vii) if at end, continue
	up <- Assembly{"JUMP-LABEL-IF-EQ", dump_continue, r5, r6} // if-is compares a register to a literal
	// if-eq compares two registers

	// IMPORTANT STEPS THAT WERE MISSING
	//   i) Decrement the refcount
	up <- Assembly{"DEREF", r7, r6, 0}
	up <- Assembly{"SUB1", r7, 0, 0}
	//   ii) Check if the refcount is zero
	isDeadMem := counter()
	up <- Assembly{"DEREF", r7, r7, 0}
	up <- Assembly{"JUMP-LABEL-IF-IS", isDeadMem, r7, 0}
	//   iii) If it isn't, r6++, next segment
	up <- Assembly{"COPY-ADD", r6, r6, 1}
	up <- Assembly{"JUMP-LABEL", rec_loop, 0, 0}
	//   if it is, continue
	up <- Assembly{"LABEL", isDeadMem, 0, 0}

	// viii) otherwise, new ds frame
	// ix) populate data structure frame
	up <- Assembly{"NEW", r7, 4, 0}
	up <- Assembly{"SET-INDEXED", r7, 0, 0}
	up <- Assembly{"SET-INDEXED", r7, 1, Type_dump}
	up <- Assembly{"DEREF", r8, r6, 0}
	up <- Assembly{"COPY-INDEXED", r7, 2, r8}
	up <- Assembly{"DEREF", r8, r4, 3}
	up <- Assembly{"COPY-INDEXED", r7, 3, r8}

	// x) stick ds frame into list
	up <- Assembly{"COPY-INDEXED", r4, 3, r7}

	// xi) next pointer
	up <- Assembly{"COPY-ADD", r6, r6, 1}

	// xii) continue loop 2
	up <- Assembly{"JUMP-LABEL", rec_loop, 0, 0}

	// xiii) set current ds frame refcount to -1 (this is where "continue" is)
	up <- Assembly{"LABEL", dump_continue, 0, 0}
	up <- Assembly{"DEREF", r7, r4, 2}
	up <- Assembly{"SET-INDEXED", r7, 0, -1}
	up <- Assembly{"SET-INDEXED", r4, 0, -1}

	// xiv) if next frame zero, exit
	up <- Assembly{"DEREF", r4, r4, 3}
	up <- Assembly{"JUMP-LABEL-IF-IS", dump_end, r4, 0}

	// xv) otherwise, next frame, continue loop 1
	// set r5 to be the data struct
	up <- Assembly{"DEREF", r5, r4, 2}
	up <- Assembly{"JUMP-LABEL", dump_start, 0, 0}
	up <- Assembly{"LABEL", dump_end, 0, 0}
	up <- Assembly{"BOILERPLATE END _f", 0, 0, 0}
	up <- Assembly{"JUMP", r3, 0, 0}
	up <- Assembly{"LABEL", endFinishFunc, 0, 0}

	// End of finishfunc.
	//
	up <- Assembly{"SDLFJSDLKFJSDLKFJ _f", 0, 0, 0}

	// '+' func
	// First actual function I've made. Adds 2 numbers. Not hard.
	endAddFunc := counter()
	up <- Assembly{"JUMP-LABEL", endAddFunc, 0, 0}
	up <- Assembly{"LABEL", sym.getSymID(builtins["add"], counter), 0, 0}
	up <- Assembly{"NEW", r3, 3, 0} // new int
	up <- Assembly{"SET-INDEXED", r3, 0, 1}
	up <- Assembly{"SET-INDEXED", r3, 1, Type_int}

	// grab args from symtab
	querySymtab(up, r4, r5, r2, sym.getSymID("_add_arg_0", counter), counter)
	querySymtab(up, r5, r6, r2, sym.getSymID("_add_arg_1", counter), counter)

	up <- Assembly{"DEREF", r4, r4, 2} // extract raw ints from int structures
	up <- Assembly{"DEREF", r5, r5, 2}

	up <- Assembly{"ADD", r4, r4, r5} // new: r4 = [r4] + [r5]
	up <- Assembly{"COPY-INDEXED", r3, 2, r4}
	//up <- Assembly{"DEREF", r0, r3, 0}
	up <- Assembly{"COPY-INDEXED", r0, 0, r3}
	up <- Assembly{"JUMP-LABEL", sym.getSymID(builtins["FINISHFUNC"], counter), 0, 0}
	up <- Assembly{"LABEL", endAddFunc, 0, 0}
	// Create closure
	up <- Assembly{"NEW", r3, 7, 0}
	up <- Assembly{"SET-INDEXED", r3, 0, 1}
	up <- Assembly{"SET-INDEXED", r3, 1, Type_closure}
	up <- Assembly{"SET-LABEL-INDEXED", r3, 2, sym.getSymID(builtins["add"], counter)} // sets a cell to be a label.
	up <- Assembly{"COPY-INDEXED", r3, 3, r2}
	up <- Assembly{"COPY-INDEXED", r3, 4, 2}
	for i := 0; i < 2; i++ {
		up <- Assembly{"NEW", r4, 3, 0}
		up <- Assembly{"SET-INDEXED", r4, 0, 1}
		up <- Assembly{"SET-INDEXED", r4, 1, Type_symbol}
		up <- Assembly{"SET-INDEXED", r4, 2, sym.getSymID("_add_arg_"+strconv.Itoa(i), counter)}
		up <- Assembly{"COPY-INDEXED", r3, 5 + i, r4}
	}

	addToSymtab(up, r4, r5, sym.getSymID(builtins["add"], counter), r3, r2)

	close(up)
}
