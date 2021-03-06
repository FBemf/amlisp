// Turns ASCII amlisp into an AST for the parser
package lexparse

import "regexp"

// import "fmt"

/* Current functionality:
   Will turn a series of lisp pieces into
   a series of tokens.
*/

// Turns code into a list of Primitives
func Lex(code string) []Primitive {

	rWhitespace := regexp.MustCompile(`^[\s,]`)
	rOpenParen := regexp.MustCompile(`^\(`)
	rCloseParen := regexp.MustCompile(`^\)`)
	rLitInt := regexp.MustCompile(`^\d+$`)
	rLitFloat := regexp.MustCompile(`^\d*\.\d+$`)
	rLitChar := regexp.MustCompile(`^#\\(?:(?:\\.)|[!-'*-+--\[\]-~])+$`)
	rLitStr := regexp.MustCompile(`^"(?:(?:\\.)|[^\\"])*"`)
	rWord := regexp.MustCompile(`^(?:(?:\\.)|[!-'*-+--\[\]-~])+`)
	// word matches anything except ( ,)\
	// and accepts escapes

	prims := make([]Primitive, 0)
	for {
		//fmt.Println(code)
		//fmt.Println(prims)
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
			prims = append(prims, Primitive{openParen, ""})
			continue
		}
		a, b = grab(rCloseParen, code)
		if len(a) != 0 {
			//fmt.Println("c")
			code = b
			prims = append(prims, Primitive{closeParen, ""})
			continue
		}
		/*
			a, b = grab(rLitInt, code)
			if len(a) != 0 {
				//fmt.Println("d")
				code = b
				prims = append(prims, Primitive{LitInt, a})
				continue
			}
			a, b = grab(rLitFloat, code)
			if len(a) != 0 {
				//fmt.Println("e")
				code = b
				prims = append(prims, Primitive{LitFloat, a})
				continue
			}
			a, b = grab(rLitChar, code)
			if len(a) != 0 {
				//fmt.Println("f")
				code = b
				a = a[2:]
				prims = append(prims, Primitive{LitChar, removeEscape(a)})
				continue
			}
		*/
		a, b = grab(rLitStr, code)
		if len(a) != 0 {
			//fmt.Println("g")
			code = b
			a = a[1 : len(a)-1]
			prims = append(prims, Primitive{LitStr, removeEscape(a)})
			continue
		}
		a, b = grab(rWord, code)
		if len(a) != 0 {
			//fmt.Println("h")
			code = b
			if rLitInt.MatchString(a) {
				prims = append(prims, Primitive{LitInt, a})
				continue
			} else if rLitFloat.MatchString(a) {
				prims = append(prims, Primitive{LitFloat, a})
				continue
			} else if rLitChar.MatchString(a) {
				a = a[2:]
				prims = append(prims, Primitive{LitChar, removeEscape(a)})
				continue
			} else {
				prims = append(prims, Primitive{Symbol, removeEscape(a)})
				continue
			}
		}
		break
	}
	return prims
}

// Removes \ before characters
func removeEscape(a string) string {
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
	return string(c)
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
