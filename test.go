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
                fmt.Println(t)
                fmt.Println(ok)
        }
        return
}
