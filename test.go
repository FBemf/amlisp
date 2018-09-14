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
		fmt.Println(rPrint(t))
		fmt.Println(ok)
                fmt.Println("Compiling...")
                code := codegen.GenAssembly(t)
                for _, i := range code {
                        fmt.Println(codegen.Disassemble(i))
                }
	}
	return
}

func rPrint(ast lexparse.Ast) string {
	if ast == nil {
		return "nil"
	} else if ast.Node() != nil {
		return fmt.Sprintf("(%s.%s)", rPrint(ast.Node().This()), rPrint(ast.Node().Next()))
	} else {
		return fmt.Sprintf("%v", ast)
	}
}
