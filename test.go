package main

import (
	"./lexparse"
        "./codegen"
	"bufio"
	"fmt"
	"os"
        "time"
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
                //for _, i := range code {
                        //fmt.Println(codegen.Disassemble(i))
                //}
                run(code)
	}
	return
}

type memuse struct {
        used bool
        start int
        end int
        next *memuse
}

func (m *memuse) alloc(mem []int, length int, prev *memuse) (addr int, ok bool) {
        if m == nil {
                return 0, false // if there's no space
        }

        //m.printmemuse()

        if m.used == true {
                if mem[m.start] == -1 {
                        m.used = false
                        if prev != nil && prev.used == false {
                                prev.end = m.end
                                prev.next = m.next
                                m = prev
                        }
                        if m.next != nil && m.next.used == false {
                                m.end = m.next.end
                                m.next = m.next.next
                        }
                } else {
                        return m.next.alloc(mem, length, m)
                }
        }

        if m.end - m.start < length {
                return m.next.alloc(mem, length, m)
        }

        if m.end - m.start == length {
                m.used = true
                return m.start, true
        }

        n := memuse{false, m.start + length, m.end, m.next}
        m.used = true
        m.end = m.start + length
        m.next = &n
        return m.start, ok
}

func (m *memuse) printmemuse() {
        if m == nil {
                return
        }
        fmt.Printf(">> %v\n", m)

        m.next.printmemuse()
}

func jumpLabel(labels map[int]int, cmd codegen.Assembly, cmds []codegen.Assembly, i int) int {
        loc, ok := labels[cmd.Arg1]
        if ok {
                i = loc
        } else {
                for {
                        i++
                        if cmds[i].Command == "LABEL" {
                                labels[cmds[i].Arg1] = i
                                if cmds[i].Arg1 == cmd.Arg1 {
                                        break
                                }
                        }
                }
        }
        return i
}

func max(nums ...int) int {
        max := nums[0]
        for _, i := range(nums) {
                if i > max {
                        max = i
                }
        }
        return max
}

func run(cmds []codegen.Assembly) {
        mem := make([]int, 1000)
        rest := memuse{false, 13, 1000, nil}
        use := memuse{true, 0, 13, &rest}
        labels := make(map[int]int)
        var i int = 0
        var largest int = 0
        // switch on assembly funcs here
        for {
                time.Sleep(time.Second/5)
                fmt.Println(mem[0:largest+1])
                cmd := cmds[i]
                fmt.Println(cmd)
                largest = max(largest, cmd.Arg1, cmd.Arg2, cmd.Arg3)
                switch (cmd.Command) {
                        case "SET-LITERAL":
                                mem[cmd.Arg1] = cmd.Arg2
                        case "SET-INDEXED":
                                mem[mem[cmd.Arg1] + cmd.Arg2] = cmd.Arg3
                        case "COPY-ADD":
                                mem[cmd.Arg1] = mem[cmd.Arg2] + cmd.Arg3
                        case "COPY-INDEXED":
                           mem[mem[cmd.Arg1] + cmd.Arg2] = mem[cmd.Arg3]
                        case "DEREF":
                                mem[cmd.Arg1] = mem[mem[cmd.Arg2] + cmd.Arg3]
                        case "NEW":
                                mem[cmd.Arg1], _ = use.alloc(mem, cmd.Arg2, nil)
                        case "LABEL":
                                labels[cmd.Arg1] = i
                        case "JUMP-LABEL":
                                i = jumpLabel(labels, cmd, cmds, i)
                        case "JUMP-LABEL-IF-IS":
                                if mem[cmd.Arg2] == cmd.Arg3 {
                                        i = jumpLabel(labels, cmd, cmds, i)
                                }
                        case "JUMP-LABEL-IF-EQ":
                                if mem[cmd.Arg2] == mem[cmd.Arg3] {
                                        i = jumpLabel(labels, cmd, cmds, i)
                                }
                        case "JUMP":
                                i = mem[cmd.Arg1]
                        case "JUMP-LABEL-REMEMBER":
                                mem[cmd.Arg2] = i
                                i = jumpLabel(labels, cmd, cmds, i)
                        case "EXCEPTION":
                                fmt.Println("You threw an exception! Oh my gosh!")
                                return
                        case "ADD1":
                                mem[cmd.Arg1] = mem[cmd.Arg1]+1
                        case "SUB1":
                                mem[cmd.Arg1] = mem[cmd.Arg1]-1
                        default:
                                fmt.Printf("SPECIAL COMMAND %s\n", cmd.Command)
                }
                i++
        }
}
