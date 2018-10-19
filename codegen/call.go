// The one that makes the code from the tree.
package codegen

import (
	"../lexparse"
	"fmt"
	"strconv"
)

func call(up chan Assembly, ast lexparse.Ast, counter func() int, sym *safeSym, quoted bool) {
	up <- Assembly{"CALL-BEGIN _f", 0, 0, 0}
	// If this ast is an empty node, don't do anything
	if ast.IsEmpty() {
		close(up)
		return
	}

	// If this node holds a primitive, we want to allocate
	// memory for it and return a pointer to it
	if p := ast.Primitive(); p != nil {
		up <- Assembly{"VAL IS " + p.Value(), 0, 0, 0}
		switch p.Type() {
		case lexparse.LitInt:
			// If this is an integer, return the value
			up <- Assembly{"LITERAL INT _f", 0, 0, 0}
			up <- Assembly{"NEW", r3, 3, 0}
			up <- Assembly{"SET-INDEXED", r3, 0, 1}        // refcount
			up <- Assembly{"SET-INDEXED", r3, 1, Type_int} // type
			val, _ := strconv.Atoi(p.Value())
			up <- Assembly{"SET-INDEXED", r3, 2, val} // val
			up <- Assembly{"COPY-INDEXED", r0, 0, r3} // return
		case lexparse.Symbol:
			// If "quoted" is on (read: this is an argument in (symbol-quote),
			// return the raw symbol for this. Otherwise, return the value
			// it has in this scope
			if quoted {
				up <- Assembly{"LITERAL SYMBOL _f", 0, 0, 0}
				up <- Assembly{"NEW", r3, 3, 0}
				up <- Assembly{"SET-INDEXED", r3, 0, 1}                                // refcount
				up <- Assembly{"SET-INDEXED", r3, 1, Type_symbol}                      // type
				up <- Assembly{"SET-LITERAL", r3, 2, sym.getSymID(p.Value(), counter)} // val
				up <- Assembly{"COPY-INDEXED", r0, 0, r3}
			} else {
				up <- Assembly{"VARIABLE SYMBOL _f", 0, 0, 0}
				querySymtab(up, r4, r3, r2, sym.getSymID(p.Value(), counter), counter)
				up <- Assembly{"COPY-INDEXED", r0, 0, r4}
			}
		default:
			fmt.Printf("Unexpected primitive type %v\n", p.Type())
			up <- Assembly{"COPY-ADD", r2, r1, 0}
			up <- Assembly{"DEREF", r1, r2, 5}
			up <- Assembly{"DEREF", r0, r2, 4}
		}
		close(up)
		return
	}

	/* FYI: A note on SET (where [A] is the value held in cell number A)
	   SET-LITERAL R A sets the cell R to A
	   COPY-ADD R A N sets the cell R to [A]+N
	   SET-INDEXED R I A sets the cell [R]+I to A
	   COPY-INDEXED R I A sets the cell [R]+I to [A]
	   DEREF R A I sets the cell R to [[A]+I]
	*/

	// This differentiates between regular function calls and special
	// pseudocalls. "func" and "symbol-quote" are compiler flags that
	// let the compiler know that this is either a function definition
	// or a literal symbol. (quote) isn't a function either, it's a
	// recursive macro. (symbol-quote) can't handle lists, only individual
	// symbol primitives. At least (symbol-quote) is a real function though.
	// "func" isn't.
	var isFunc bool = false
	if p := ast.This().Primitive(); p != nil && p.Type() == lexparse.Symbol {
		if p.Value() == builtins["FUNCTION"] {
			isFunc = true
		} else if p.Value() == builtins["SYMBOL-QUOTE"] && !quoted {
			quoted = true
		}
	}

	// If this is a function definition, defer execution
	// Function definitions are basically just delayed "do"
	// statements. We skip over it for now and return a closure
	// with a PC pointer to this place, so that it can be run later
	var funcStart, funcEnd int
	var argAst lexparse.Ast
	if isFunc {
		funcStart = counter()
		funcEnd = counter()
		up <- Assembly{"FUNC DEF _f", 0, 0, 0}
		up <- Assembly{"JUMP-LABEL", funcEnd, 0, 0}
		up <- Assembly{"LABEL", funcStart, 0, 0}

		// We also prune off the "func" keyword
		// and the list of arguments, because
		// a) "func" isn't a real function, it's
		//    a compiler keyword
		// b) we'll need those args later, and
		// c) we don't want our function trying
		//    to execute a list of variables
		ast = ast.Next().Node()
		argAst = ast.This().Node()
		ast = ast.Next().Node()
	}

	// Count members of the s-expression which defines this
	// function call
	members := 0
	for t := ast; t.Node() != nil; t = t.Next() {
		if t.This().IsEmpty() == false {
			members++
		}
	}
	//fmt.Printf("MEMBERS: %d\n", members) //  a little piece of debug code

	up <- Assembly{"NEW ENV _f", 0, 0, 0} // Assembly calls ending in "_f" are just
	// comments I left so I could see where I
	// was in the code as I ran it

	// Create a new environment for this function call
	// An environment is a data structure holding all the relevant
	// information for a single function call, such as where to write
	// the return value, or all the local variables

	// The structure of an environment is:
	// i)    Reference Count: number of pointers to this structure that exist
	// ii)   Type: gonna beType_env
	// iii)  Length: Numberof local variables stored here
	// iv)   Saved PC: Progam counter to return to after this function call
	// v)    Return Location: Where to write the return value to
	// vi)   Parent Environment: The environment from which this call was made
	// vii)  Symbol Table: A pointer to a stack holding information about
	//       which symbols refer to which values in this scope
	// viii) Pointers: From here forward are pointers to all the variables used
	//       in this scope

	up <- Assembly{"NEW", r2, members + 7, 0}
	up <- Assembly{"COPY-INDEXED", r0, 0, r2} // Assumes r0 is return loc
	up <- Assembly{"SET-INDEXED", r2, 0, 1}
	up <- Assembly{"SET-INDEXED", r2, 1, Type_environment}
	up <- Assembly{"SET-INDEXED", r2, 2, members}
	up <- Assembly{"COPY-INDEXED", r2, 4, r0}
	up <- Assembly{"COPY-INDEXED", r2, 5, r1} // Assumes r1 is return env
	up <- Assembly{"DEREF", r3, r1, 6}
	up <- Assembly{"COPY-INDEXED", r2, 6, r3} // grab symbol table
	up <- Assembly{"ADD1", r3, 0, 0}          // Increment symtab refcount

	// Finished base env creation

	// This part calls the arguments and puts the results into
	// the "Pointers" slots
	argCode := make([]chan Assembly, members)
	up <- Assembly{"ARGS _f", 0, 0, 0}
	for m := 0; m < members; m++ {
		argCode[m] = make(chan Assembly, 100)
		//fmt.Println(lexparse.RPrint(ast))
		if !quoted || m == 0 {
			go call(argCode[m], ast.This(), counter, sym, false)
		} else {
			go call(argCode[m], ast.This(), counter, sym, true)
		}
		ast = ast.Next()
	}

	// The block above is run concurrently, but the code needs to be
	// in order. This block de-channels the code from the last block
	for m, c := range argCode {
		up <- Assembly{"COPY-ADD", r0, r2, 7 + m} // <- These two commands set the
		up <- Assembly{"COPY-ADD", r1, r2, 0}     // <- return location and parent env
		for a, b := <-c; b; a, b = <-c {          //    for each argument call
			up <- a
		}
	}

	// If this function call is a function definition, there are
	// extra steps
	if isFunc {
		up <- Assembly{"DEREF", r3, r2, members + 6} // The last "argument" of this
		up <- Assembly{"COPY-INDEXED", r0, 0, r3}    // function is the last expression
		// in the function's definition.
		// Here we take the result of it
		// and return it.

		// Finishfunc is a piece of code that does a handful of things,
		// most importantly garbage collection. It is called after any function
		// is executed.
		up <- Assembly{"JUMP-LABEL", sym.getSymID(builtins["FINISHFUNC"], counter), 0, 0}

		up <- Assembly{"LABEL", funcEnd, 0, 0} // End of function runtime
		up <- Assembly{"FUNC END _f", 0, 0, 0}

		// Code past this point is only executed when the function is defined,
		// not when it is executed

		// Count number of args
		args := 0
		for t := argAst; t.Node() != nil; t = t.Next() {
			if t.This().IsEmpty() == false {
				args++
			}
		}

		// Create and populate a closure to return
		// Closures are how the runtime stores defined functions.
		// Environments hold primarily variables defined in that scope,
		// closures hold a PC pointer to their definition and a regular pointer
		// to the environment in which they were defined, so they can access
		// variables local to that environment.

		// A closure is composed of:
		// i)   Reference Count: How many pointers exist to this closure
		// ii)  Type: Always gonna be Type_closure
		// iii) PC Address: The Program Counter address that the function is
		//              defined at. Practically speaking, it hold the func_start
		//              label.
		// iv)  Parent Environment: A pointer to the environment of the function
		//      in which this function was defined.
		// v)   Length: How many args this function has.
		// vi)  Args: This is the first of arbitrarily many pointers to symbols
		//      defining arguments for this function.
		// contents of a closure: refcount, type, pc addr, parent env loc, length, args ...
		up <- Assembly{"NEW", r3, args + 4, 0}
		up <- Assembly{"SET-INDEXED", r3, 0, 1}
		up <- Assembly{"SET-INDEXED", r3, 1, Type_closure}
		up <- Assembly{"SET-INDEXED", r3, 2, funcStart}
		up <- Assembly{"COPY-INDEXED", r3, 3, r2}
		up <- Assembly{"COPY-INDEXED", r3, 4, args}
		for i := 0; i < args; i++ {
			up <- Assembly{"NEW", r4, 3, 0}
			up <- Assembly{"SET-INDEXED", r4, 0, 1}
			up <- Assembly{"SET-INDEXED", r4, 1, Type_symbol}
			up <- Assembly{"SET-INDEXED", r4, 2, sym.getSymID(argAst.Node().This().Primitive().Value(), counter)}
			up <- Assembly{"COPY-INDEXED", r3, 5 + i, r4}
		}

		up <- Assembly{"COPY-INDEXED", r0, 0, r3} // Return this closure to the
		// function defining it
		// Jump to finishfunc to clean up
		up <- Assembly{"JUMP-LABEL", sym.getSymID(builtins["FINISHFUNC"], counter), 0, 0}
		up <- Assembly{"FUNC RETURN END _f", 0, 0, 0}

	} else {
		// If this is not a function definition, then we
		// have a function to execute
		up <- Assembly{"DEREF", r4, r2, 7} // Get the relevant closure
		up <- Assembly{"FUNC CALL _f", 0, 0, 0}

		// TODO: detect when a function is called with the wrong number of args

		// This is as good a place as any to define symbol table cells
		// A symbol table cell is a link between a symbol, i.e. a variable name
		// like "foo," to it's value, e.g. "the integer 6." They form a stack,
		// so that, if a function redefines "foo," it'll use its more recent
		// definition, while other environments, which have pointers to frames
		// lower in the stack, will still access their defintitions. As different
		// functions define different variables, the stack branches into a ~FIFO tree.
		//
		// Components of a symbol table frame:
		// i)   Reference count: How many pointers to this exist?
		// ii)  Type: Always gonna be Type_symtab.
		// iii) ID: A pointer to a symbol struct, which holds a unique
		//      numeric constant equivalent to a string like "foo."
		// iv)  Location: A pointer to a data type, like an int, which
		//      the ID "foo" is valued to in this scope.
		// v)   Next: A pointer to the symbol table frame below this one
		//      in the stack.

		// This section iterates through the argument symbols of the closure
		// and the arguments provided in the function call, and links them
		// with symbol table frames. It then pushes these frames to the symbol
		// table so that the top entry with that symbol will hold the relevant
		// argument for that function

		for m := 1; m < members; m++ {

			// I could *probably* replace this whole block with
			// addToSymtab if I was willing to rework it to take a
			// symbol from a register

			// Find the ID of the symbol held in the closure
			up <- Assembly{"DEREF", r3, r4, 4 + m} // r3 = [[r4] + 4+m]
			up <- Assembly{"DEREF", r3, r3, 2}     // r3 = [[r3] + 2]
			// Then find the related argument in the function call
			// and set the location pointer in the symtab frame to be that
			up <- Assembly{"DEREF", r7, r2, 7 + m} // r3 = [[r2] + 7+m]
			// Add it to symtab
			addToSymtabRegister(up, r5, r6, r3, r7, r2)

			/*	// So this part is completely redundant now
			
			   up <- Assembly{"NEW", r5, 5, 0}                 // This block creates a
			   up <- Assembly{"SET-INDEXED", r5, 0, 1}         // new symtab frame
			   up <- Assembly{"SET-INDEXED", r5, 1, Type_symtab}

			   // Find the ID of the symbol held in the closure
			   up <- Assembly{"DEREF", r3, r4, 4+m}    // r3 = [[r4] + 4+m]
			   up <- Assembly{"DEREF", r3, r3, 2}      // r3 = [[r3] + 2]

			   // Set the ID in the symbol table cell to that ID
			   up <- Assembly{"COPY-INDEXED", r5, 2, r3}       // [r5] + 2 = [r3]

			   // Then find the related argument in the function call
			   // and set the location pointer in the symtab frame to be that
			   up <- Assembly{"DEREF", r3, r2, 7+m}    // r3 = [[r2] + 7+m]
			   up <- Assembly{"COPY-INDEXED", r5, 3, r3}       // [r5]+3 = [r3]

			   // Then push it onto the symbol table stack
			   up <- Assembly{"DEREF", r3, r2, 6}      // r3 = [[r2]+6]
			   up <- Assembly{"COPY-INDEXED", r5, 4, r3}       // [r5]+4 = [r3]
			   up <- Assembly{"COPY-INDEXED", r2, 6, r5}       // [r2]+6 = [r5]
			*/
		}

		// Give it somewhere to return to
		up <- Assembly{"DEREF", r0, r2, 4}
		// Lastly, make the jump into the function's runtime
		up <- Assembly{"DEREF", r4, r4, 2}         // Grab jump location // changed 3 to 2
		up <- Assembly{"COPY-ADD", r3, r2, 3}      // r3 = [r2] + 3
		up <- Assembly{"JUMP-REMEMBER", r4, r3, 0} // saves next pc to the cell
		// in our environment and jump
		// into the function

		// Go to finishfunc to clear up this env
		// TODO this breaks the code.
		// TODO the argument here has to point to somewhere to store the ptr, it isn't
		// just a register to store it in. It's indirect.

		// Give it somewhere to return to
		up <- Assembly{"DEREF", r0, r2, 4}
		// Lastly, make the jump into the function's runtime
		up <- Assembly{"DEREF", r4, r4, 2}         // Grab jump location // changed 3 to 2
		up <- Assembly{"COPY-ADD", r3, r2, 3}      // r3 = [r2] + 3
		up <- Assembly{"JUMP-LABEL-REMEMBER", sym.getSymID(builtins["FINISHFUNC"], counter), r3, 0}

	}
	close(up)
	return
}
