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
                text, _ := reader.ReadSlice('\n')
                p := GetPrimitives(text)
                fmt.Println(text)
        }
        return
}
