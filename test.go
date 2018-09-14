package main

import (
	"./lexparse"
        "./codegen"
	"bufio"
	"fmt"
	"os"
)

func main() {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter code: ")
		text, _ := reader.ReadString('\n')
		p := lexparse.Lex(text)
		fmt.Println(p)
		t, ok := lexparse.Parse(p)
		fmt.Println(ok)
                if ok != nil {
                        continue
                }
		fmt.Println(lexparse.RPrint(t))
                var _ = codegen.GenAssembly
                fmt.Println("Compiling...")
                code := codegen.GenAssembly(t)
                for _, i := range code {
                        fmt.Println(codegen.Disassemble(i))
                }
	}
	return
}
