package main

import (
        "bufio"
        "fmt"
        "os"
)

func main() {
        for {
                reader := bufio.NewReader(os.Stdin)
                fmt.Print("Enter code: ")
                text, _ := reader.ReadString('\n')
                p := getPrimitives(text)
                fmt.Println(p)
        }
        return
}
