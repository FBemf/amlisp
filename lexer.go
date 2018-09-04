// Turns ASCII amlisp into an AST for the parser
package lexer

import "regexp"

/* Current functionality:
   Will turn a series of lisp pieces into
   a series of tokens. Will not attempt to evaluate
   their legitimacy or types.
*/

// The basic components of an amlisp program after
// all the reader macros have run.
type Primitive struct {
        kind int
        content string
}

const (
        symbol = iota
        openParen
        closeParen
)


rWhitespace := regexp.MustCompile(`^[\s,]`)
rOpenParen := regexp.MustCompile(`^(`)
rCloseParen := regexp.MustCompile(`^)`)
rWord := regexp.MustCompile(`^\\.|[!-'*-\[\]-~]`) // Matches anything except ()\
                                                  // and accepts escapes
rInt := regexp.MustCompile(`^\d+$`)
rFloat := regexp.MustCompile(`^\d+\.\d+$`)


// Turns code into a list of Primitives
func getPrimitives(code []char) []Primitive {
        prims := make([]Primitive, 0)
        for {
                a, b = rWhitespace.grab(code)
                if len(a) != 0 {
                        code = b
                        continue
                }
                a, b = rOpenParen.grab(code)
                if len(a) != 0 {
                        code = b
                        prims = append(prims, Primitive{openParen, nil)
                        continue
                }
                a, b = rCloseParen.grab(code)
                if len(a) != 0 {
                        code = b
                        prims = append(prims, Primitive{closeParen, nil)
                        continue
                }
                a, b = rWord.grab(code)
                if len(a) != 0 {
                        code = b
                        prims = append(prims, Primitive{symbol, a)
                        continue
                }
                break
        }
        return prims
}

// Returns the first match and the rest of the string
func (re *Regexp) grab(s string) (string, string) {
        loc := re.regexp.FindStringIndex(s)
        return (s[loc[0]:loc[1]], s[loc[1]:])
}
