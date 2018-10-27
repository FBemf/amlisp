# AMLISP

## Application-eMbedded LISP

What if every application could have emacs lisp? Amlisp is a 
simple lisp dialect which compiles to extensible bytecode. 
Embedding amlisp in an application is as simple as setting up 
a while loop to execute the bytecode commands.

This project is currently on an indefinite hiatus, which is to
say that I'm happy with where it is now and will probably never
take it further, but still want to leave my options open.

I took on this project because I wanted to learn more about the
function/closure/environment/symbol-table architecture of functional
programming languages, and this language has all of those now. You
can run a command like

	((func ()
		(define (iquote a) (func (a b c d) (a (b c) d)))
		(define (iquote b) 101)
		(define (iquote c) (func (a) (+ a 202)))
		(a + c 202 b)))

and have it return `606`, just as expected. It sets up all the scopes
and stuff necessary for a functional language to work, and the fact
that this little language of mine doesn't actually have conditional
statements or any real function other than + doesn't really matter,
because I got out of this language everything I wanted to learn, and
I think that the rest is mostly legwork. Important legwork, no doubt,
but I'd rather move on to another project and learn about something
else rather than keep gradually fleshing this out.
