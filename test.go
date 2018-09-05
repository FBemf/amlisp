package main

import (
        "bufio"
        "fmt"
        "os"
        "./lexparse"
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
