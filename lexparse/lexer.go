// Turns ASCII amlisp into an AST for the parser
package lexparse

import "regexp"

//import "fmt"

/* Current functionality:
   Will turn a series of lisp pieces into
   a series of tokens. Will not attempt to evaluate
   their legitimacy or types.
*/

// The basic components of an amlisp program after
// all the reader macros have run.
type primitive struct {
	kind    int
	content string
}

func (p primitive) Type() int {
        return p.kind
}

func (p primitive) Value() string {
        return p.content
}

const (
	Symbol = iota
	LitInt
	LitFloat
	LitChar
	openParen
	closeParen
)

// Turns code into a list of primitives
func Lex(code string) []primitive {

	rWhitespace := regexp.MustCompile(`^[\s,]`)
	rOpenParen := regexp.MustCompile(`^\(`)
	rCloseParen := regexp.MustCompile(`^\)`)
	rLitInt := regexp.MustCompile(`^\d+$`)
	rLitFloat := regexp.MustCompile(`^\d*\.\d+$`)
	rLitChar := regexp.MustCompile(`^#\\(?:(?:\\.)|[!-'*-+--\[\]-~])+$`)
	rWord := regexp.MustCompile(`^(?:(?:\\.)|[!-'*-+--\[\]-~])+`)
	// word matches anything except ( ,)\
	// and accepts escapes

	prims := make([]primitive, 0)
	for {
		var a, b string
		a, b = grab(rWhitespace, code)
		if len(a) != 0 {
			//fmt.Println("a")
			code = b
			continue
		}
		a, b = grab(rOpenParen, code)
		if len(a) != 0 {
			//fmt.Println("b")
			code = b
			prims = append(prims, primitive{openParen, ""})
			continue
		}
		a, b = grab(rCloseParen, code)
		if len(a) != 0 {
			//fmt.Println("c")
			code = b
			prims = append(prims, primitive{closeParen, ""})
			continue
		}
		a, b = grab(rLitInt, code)
		if len(a) != 0 {
			//fmt.Println("d")
			code = b
			prims = append(prims, primitive{LitInt, a})
			continue
		}
		a, b = grab(rLitFloat, code)
		if len(a) != 0 {
			//fmt.Println("e")
			code = b
			prims = append(prims, primitive{LitFloat, a})
			continue
		}
		a, b = grab(rLitChar, code)
		if len(a) != 0 {
			//fmt.Println("f")
			code = b
			a = a[2:]
			// The below block handles the escape character
			// The regex matches, it, this bit removes it
			c := make([]rune, 0)
			for i := 0; i < len(a); i++ {
				if a[i] == '\\' {
					if i < len(a)-1 {
						c = append(c, rune(a[i+1]))
						i++
					}
				} else {
					c = append(c, rune(a[i]))
				}
			}
			prims = append(prims, primitive{LitChar, string(c)})
			continue
		}
		a, b = grab(rWord, code)
		if len(a) != 0 {
			//fmt.Println("g")
			code = b
			if rLitInt.MatchString(a) {
				prims = append(prims, primitive{LitInt, a})
				continue
			} else if rLitFloat.MatchString(a) {
				prims = append(prims, primitive{LitFloat, a})
				continue
			} else if rLitChar.MatchString(a) {
				a = a[2:]
				c := make([]rune, 0)
				for i := 0; i < len(a); i++ {
					if a[i] == '\\' {
						if i < len(a)-1 {
							c = append(c, rune(a[i+1]))
							i++
						}
					} else {
						c = append(c, rune(a[i]))
					}
				}
				prims = append(prims, primitive{LitChar, a})
				continue
			}
			// The below block handles the escape character
			// The regex matches, it, this bit removes it
			c := make([]rune, 0)
			for i := 0; i < len(a); i++ {
				if a[i] == '\\' {
					if i < len(a)-1 {
						c = append(c, rune(a[i+1]))
						i++
					}
				} else {
					c = append(c, rune(a[i]))
				}
			}
			prims = append(prims, primitive{Symbol, string(c)})
			continue
		}
		break
	}
	return prims
}

// Returns the first match and the rest of the string
func grab(re *regexp.Regexp, s string) (string, string) {
	loc := re.FindStringIndex(s)
	if len(loc) > 0 {
		return s[loc[0]:loc[1]], s[loc[1]:]
	} else {
		return "", ""
	}
}
