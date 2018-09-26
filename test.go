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
                fmt.Println("Compiling...")
                code := codegen.GenAssembly(t)
                //for _, i := range code {
                        //fmt.Println(codegen.Disassemble(i))
                //}
                fmt.Println()
                //interpret(code)
                //run(assemble(code))
                h := assemble(code)
                run(h)
                //for i := 0; i < len(h); i++ {
                        //fmt.Println(h[i])
                //}
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

func interpret(cmds []codegen.Assembly) {
        mem := make([]int, 1000)
        rest := memuse{false, 13, 1000, nil}
        use := memuse{true, 0, 13, &rest}
        labels := make(map[int]int)
        largest := 0
        // switch on assembly funcs here
        for i := 0; i < len(cmds); i++ {
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
                                largest = max(largest, mem[cmd.Arg1]+cmd.Arg2)
                        case "COPY-ADD":
                                mem[cmd.Arg1] = mem[cmd.Arg2] + cmd.Arg3
                                largest = max(largest, mem[cmd.Arg2])
                        case "COPY-INDEXED":
                                mem[mem[cmd.Arg1] + cmd.Arg2] = mem[cmd.Arg3]
                                largest = max(largest, mem[cmd.Arg3], mem[mem[cmd.Arg1]+cmd.Arg2])
                        case "DEREF":
                                mem[cmd.Arg1] = mem[mem[cmd.Arg2] + cmd.Arg3]
                                largest = max(largest, mem[cmd.Arg1])
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
                        case "JUMP-REMEMBER":
                                mem[cmd.Arg2] = i
                                i = mem[cmd.Arg1]
                        case "EXCEPTION":
                                fmt.Println("You threw an exception! Oh my gosh!")
                                return
                        case "ADD1":
                                mem[mem[cmd.Arg1]] = mem[mem[cmd.Arg1]]+1
                                largest = max(largest, mem[cmd.Arg1])
                        case "SUB1":
                                mem[mem[cmd.Arg1]] = mem[mem[cmd.Arg1]]-1
                                largest = max(largest, mem[cmd.Arg1])
                        default:
                                fmt.Printf("SPECIAL COMMAND %s\n", cmd.Command)
                }
        }
}

func assemble(cmds []codegen.Assembly) (bc []codegen.Assembly) {
        labels := make(map[int]int)
        interm := make([]codegen.Assembly, 0, len(cmds))
        // switch on assembly funcs here
        newAddr := 0
        for i := 0; i < len(cmds); i++ {
                cmd := cmds[i]
                switch (cmd.Command) {
                        case "LABEL":
                                labels[cmd.Arg1] = newAddr
                        default:
                                interm = append(interm, cmd)
                                newAddr++
                }
        }
        bc = make([]codegen.Assembly, 0, len(interm))
        for i := 0; i < len(interm); i++ {
                cmd := interm[i]
                switch (cmd.Command) {
                        case "JUMP-LABEL":
                                bc = append(bc, codegen.Assembly{"JUMP-LITERAL", labels[cmd.Arg1], cmd.Arg2, cmd.Arg3})
                        case "JUMP-LABEL-IF-IS":
                                bc = append(bc, codegen.Assembly{"JUMP-LITERAL-IF-IS", labels[cmd.Arg1], cmd.Arg2, cmd.Arg3})
                        case "JUMP-LABEL-IF-EQ":
                                bc = append(bc, codegen.Assembly{"JUMP-LITERAL-IF-EQ", labels[cmd.Arg1], cmd.Arg2, cmd.Arg3})
                        //case "JUMP-LABEL-REMEMBER":
                                //bc = append(bc, codegen.Assembly{"JUMP-LITERAL-REMEMBER", labels[cmd.Arg1], cmd.Arg2, cmd.Arg3})
                        case "SET-LABEL-INDEXED":       // register, offset, label#
                                bc = append(bc, codegen.Assembly{"SET-INDEXED", cmd.Arg1, cmd.Arg2, labels[cmd.Arg3]})
                        default:
                                bc = append(bc, cmd)
                }
        }
        return
}

func run(cmds []codegen.Assembly) {
        mem := make([]int, 1000)
        rest := memuse{false, 13, 1000, nil}
        use := memuse{true, 0, 13, &rest}
        largest := 0
        // switch on "bytecode" here
        for i := 0; i < len(cmds); i++ {
                time.Sleep(time.Second/13)
                fmt.Print("COMMAND: ")
                fmt.Println(i)
                fmt.Println(mem[0:largest+1])
                fmt.Println(cmds[0:i])
                fmt.Println(cmds[i])
                fmt.Println(cmds[i+1:])
                use.printmemuse()
                cmd := cmds[i]
                fmt.Println(cmd)
                largest = max(largest, cmd.Arg1, cmd.Arg2, cmd.Arg3)
                switch (cmd.Command) {
                        case "SET-LITERAL":
                                mem[cmd.Arg1] = cmd.Arg2
                        case "SET-INDEXED":
                                mem[mem[cmd.Arg1] + cmd.Arg2] = cmd.Arg3
                                largest = max(largest, mem[cmd.Arg1]+cmd.Arg2)
                        case "COPY-ADD":
                                mem[cmd.Arg1] = mem[cmd.Arg2] + cmd.Arg3
                                largest = max(largest, mem[cmd.Arg2])
                        case "COPY-INDEXED":
                                mem[mem[cmd.Arg1] + cmd.Arg2] = mem[cmd.Arg3]
                                largest = max(largest, mem[cmd.Arg3], mem[mem[cmd.Arg1]+cmd.Arg2])
                        case "DEREF":
                                mem[cmd.Arg1] = mem[mem[cmd.Arg2] + cmd.Arg3]
                                largest = max(largest, mem[cmd.Arg1])
                        case "NEW":
                                mem[cmd.Arg1], _ = use.alloc(mem, cmd.Arg2, nil)
                        case "JUMP-LITERAL":
                                i = cmd.Arg1 - 1
                        case "JUMP-LITERAL-IF-IS":
                                if mem[cmd.Arg2] == cmd.Arg3 {
                                        i = cmd.Arg1 - 1
                                }
                        case "JUMP-LITERAL-IF-EQ":
                                if mem[cmd.Arg2] == mem[cmd.Arg3] {
                                        i = cmd.Arg1 - 1
                                }
                        case "JUMP":
                                i = mem[cmd.Arg1] - 1
                        case "JUMP-REMEMBER":
                                mem[cmd.Arg2] = i+1
                                i = mem[cmd.Arg1] - 1
                        case "EXCEPTION":
                                fmt.Println("You threw an exception! Oh my gosh!")
                                return
                        case "ADD1":
                                mem[mem[cmd.Arg1]] = mem[mem[cmd.Arg1]]+1
                                largest = max(largest, mem[cmd.Arg1])
                        case "SUB1":
                                mem[mem[cmd.Arg1]] = mem[mem[cmd.Arg1]]-1
                                largest = max(largest, mem[cmd.Arg1])
                        case "ADD":
                                mem[mem[cmd.Arg1]] = mem[mem[cmd.Arg2]] + mem[mem[cmd.Arg3]]
                                largest = max(largest, mem[cmd.Arg1])
                        case "SUB":
                                mem[mem[cmd.Arg1]] = mem[mem[cmd.Arg2]] - mem[mem[cmd.Arg3]]
                                largest = max(largest, mem[cmd.Arg1])
                        case "BREAK!":
                                return
                        default:
                                fmt.Printf("SPECIAL COMMAND %s\n", cmd.Command)
                }
        }
}
