# What's Next?

The language can now do simple addition, as in "(+ 4 5)".
It's testable now. This is a big deal.

What's up next:
        - A more salient top-level management
        Right now, return values are just all shunted up
        towards [0], but then if something's overwritten,
        it isn't garbage-collected. The top-level environment
        should be handling return values from function calls.

        - Nested calls need to be debugged
        Right now, "(+ 2 (+ 4 5))" doesn't work, and idk why.
        This is probably a simple enough debugging job, albeit
        probably a pain in the ass.

        - A better debugging interface
        I have a headache. A nicer debugging interface would do
        wonders!

        - Floatify all the numbers
        Like in Lua. I don't actually need strict ints anywhere; it's
        easier if I just don't use them.

        - Quote, define
        Gotta get these working

        - Some kind of print function
        Not necessarily forever, but until I iron out extensibility

        - More comments
        Always good. My code's mostly commented, but some of the newer
        stuff doesn't have great comments. Bring that up to snuff.

        - All the other functions
        Subtract, multiply, cons/car/cdr, all the rest of the builtins

        - Strings
        These are just gonna be lists of chars, probably

        - Vectors
        Unless I make them vectors of chars. This may or may not
        make it in.

        - Macros
        Gonna need to be done at some point. A lot of things in the
        language are macros. Probably gonna make the macro parser an
        amlisp vm embedded in the compiler.
