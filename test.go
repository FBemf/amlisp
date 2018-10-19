package main

import (
	"./codegen"
	"./lexparse"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func main() {
	defer func() {
		exec.Command("stty", "-F", "/dev/tty", "-cbreak").Run()
	}()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter code: ")
	text, _ := reader.ReadString('\n')
	p := lexparse.Lex(text)
	fmt.Println(p)
	t, ok := lexparse.Parse(p)
	fmt.Println(ok)
	if ok != nil {
		fmt.Println("Could not parse!")
		return
	}
	fmt.Println(lexparse.RPrint(t))
	fmt.Println("Compiling...")
	code := codegen.GenAssembly(t)
	fmt.Println()
	h := assemble(code)
	fmt.Println("Executing...")
	// disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	interactive_interpret(h);
	/*
	events := run(h)
	//fmt.Print(events)
	fmt.Println("Finished. Post-mortem:")

	var oldCount int = 1
	var index int = 0
	var last byte = 'j'
	var b []byte = make([]byte, 1)
	var s string = ""
	fmt.Printf("Line %d.\n", index)
	for {
		os.Stdin.Read(b)
		if b[0] >= byte('0') && b[0] <= byte('9') {
			s += string(b)
			continue
		}
		c, _ := strconv.Atoi(s)
		s = ""
		if c == 0 {
			c++
		}
		fmt.Println()

		exec:
		for i := 0; i < c; i++ {
			switch (b[0]) {
			case 'j':
				if (index < len(events)-1) {
					index++
				}
			case 'k':
				if (index > 0) {
					index--
				}
			case 'p':
				fmt.Println(printmem(events[index].mem, &events[index].use, ""))
			case 'c':
				fmt.Println(events[index].command)
			case 'u':
				events[index].use.printmemuse()
			case 'g':
				if 0 > c {
					index = 0
				} else if len(events) <= c {
					index = len(events) - 1
				} else {
					index = c
				}
			default:
				b[0] = last
				c = oldCount
				goto exec
			}
		}
		fmt.Printf("Line %d, pos %d: %v\n", index, events[index].position, events[index].command)
		oldCount = c
		last = b[0]
	}
		//*/
	return
}

type memuse struct {
	used  bool
	start int
	end   int
	next  *memuse
}

func memuseCopy(m *memuse) *memuse {
	if m == nil {
		return nil
	} else {
		return &memuse{m.used, m.start, m.end, memuseCopy(m.next)}
	}
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

	if m.end-m.start < length {
		return m.next.alloc(mem, length, m)
	}

	if m.end-m.start == length {
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
	for _, i := range nums {
		if i > max {
			max = i
		}
	}
	return max
}

func printInd(a []int) (out string) {
	out = ""
	out += fmt.Sprintf("[")
	for i, v := range a {
		out += fmt.Sprintf(" %d:%d,", i, v)
	}
	out += fmt.Sprintf(" ]\n")
	return
}

func printmem(mem []int, use *memuse, glob string) string {
	if use == nil {
		return glob
	}
	if use.used == false {
		return printmem(mem, use.next, glob)
	}
	types := map[int]string {
		codegen.Type_environment: "ENV",
		codegen.Type_closure: "CLO",
		codegen.Type_dump: "DUMP",
		codegen.Type_symtab: "STAB",
		codegen.Type_cons: "CONS",
		codegen.Type_vector: "VEC",
		codegen.Type_int: "INT",
		codegen.Type_symbol: "SYM"}

	s := "["
	if use.start == 0 {
		s += fmt.Sprintf("(%d:%d), (%d:%d)", use.start, mem[use.start], use.start+1, mem[use.start+1])
	} else {
		s += fmt.Sprintf("(%d:%d), %s", use.start, mem[use.start], types[mem[use.start+1]])
	}
	for i := use.start+2; i < use.end; i++ {
		s += fmt.Sprintf(", (%d:%d)", i, mem[i])
	}
	s += "]\n"
	return printmem(mem, use.next, glob + s)
}

func assemble(cmds []codegen.Assembly) (bc []codegen.Assembly) {
	labels := make(map[int]int)
	interm := make([]codegen.Assembly, 0, len(cmds))
	// switch on assembly funcs here
	newAddr := 0
	for i := 0; i < len(cmds); i++ {
		cmd := cmds[i]
		switch cmd.Command {
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
		switch cmd.Command {
		case "JUMP-LABEL":
			bc = append(bc, codegen.Assembly{"JUMP-LITERAL", labels[cmd.Arg1], cmd.Arg2, cmd.Arg3})
		case "JUMP-LABEL-IF-IS":
			bc = append(bc, codegen.Assembly{"JUMP-LITERAL-IF-IS", labels[cmd.Arg1], cmd.Arg2, cmd.Arg3})
		case "JUMP-LABEL-IF-EQ":
			bc = append(bc, codegen.Assembly{"JUMP-LITERAL-IF-EQ", labels[cmd.Arg1], cmd.Arg2, cmd.Arg3})
		case "JUMP-LABEL-REMEMBER":
			bc = append(bc, codegen.Assembly{"JUMP-LITERAL-REMEMBER", labels[cmd.Arg1], cmd.Arg2, cmd.Arg3})
		case "SET-LABEL-INDEXED": // register, offset, label#
			bc = append(bc, codegen.Assembly{"SET-INDEXED", cmd.Arg1, cmd.Arg2, labels[cmd.Arg3]})
		default:
			bc = append(bc, cmd)
		}
	}
	return
}

type instant struct {
	command codegen.Assembly
	mem []int
	use memuse
	largest int
	message string
	position int
}

func enlargeMem(mem []int, size int) []int {
	if cap(mem) <= size {
		n := make([]int, 2*size)
		copy(n, mem)
		mem = n
	}
	if len(mem) <= size {
		mem = mem[:size+1]
	}
	return mem
}

func run(cmds []codegen.Assembly) []instant {
	mem := make([]int, 20, 1000)
	rest := memuse{false, 13, 1000, nil}
	use := memuse{true, 0, 13, &rest}
	largest := 0
	history := make([]instant, 0, 400)
	// switch on "bytecode" here
	for i := 0; i < len(cmds); i++ {
		cmd := cmds[i]
		message := ""
		//fmt.Println(i)
		//fmt.Println(cmd)
		//fmt.Println(printInd(mem))

		switch cmd.Command {
		case "SET-LITERAL":
			largest = max(largest, cmd.Arg1)
			mem = enlargeMem(mem, largest)
			mem[cmd.Arg1] = cmd.Arg2
		case "SET-INDEXED":
			largest = max(largest, mem[cmd.Arg1]+cmd.Arg2)
			mem = enlargeMem(mem, largest)
			mem[mem[cmd.Arg1]+cmd.Arg2] = cmd.Arg3
		case "COPY-ADD":
			largest = max(largest, cmd.Arg1, cmd.Arg2)
			mem = enlargeMem(mem, largest)
			mem[cmd.Arg1] = mem[cmd.Arg2] + cmd.Arg3
		case "COPY-INDEXED":
			largest = max(largest, cmd.Arg3, mem[cmd.Arg1]+cmd.Arg2)
			mem = enlargeMem(mem, largest)
			mem[mem[cmd.Arg1]+cmd.Arg2] = mem[cmd.Arg3]
		case "DEREF":
			largest = max(largest, mem[cmd.Arg2]+cmd.Arg3, mem[cmd.Arg1])
			mem = enlargeMem(mem, largest)
			mem[cmd.Arg1] = mem[mem[cmd.Arg2]+cmd.Arg3]
		case "NEW":
			largest = max(largest, cmd.Arg1)
			mem = enlargeMem(mem, largest)
			mem[cmd.Arg1], _ = use.alloc(mem, cmd.Arg2, nil)
			largest = max(largest, mem[cmd.Arg1]+cmd.Arg2-1)
			mem = enlargeMem(mem, largest)
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
			mem[mem[cmd.Arg2]] = i + 1
			i = mem[cmd.Arg1] - 1
		case "JUMP-LITERAL-REMEMBER":	// new, maybe broken
			mem[mem[cmd.Arg2]] = i + 1
			i = cmd.Arg1 - 1
		case "EXCEPTION":
			message += fmt.Sprintln("You threw an exception! Oh my gosh!")
			break
		case "ADD1":
			largest = max(largest, mem[cmd.Arg1])
			mem = enlargeMem(mem, largest)
			mem[mem[cmd.Arg1]] = mem[mem[cmd.Arg1]] + 1
		case "SUB1":
			largest = max(largest, mem[cmd.Arg1])
			mem = enlargeMem(mem, largest)
			mem[mem[cmd.Arg1]] = mem[mem[cmd.Arg1]] - 1
		case "ADD":
			largest = max(largest, cmd.Arg1, cmd.Arg2, cmd.Arg3)
			mem = enlargeMem(mem, largest)
			mem[cmd.Arg1] = mem[cmd.Arg2] + mem[cmd.Arg3]
		case "SUB":
			largest = max(largest, cmd.Arg1, cmd.Arg2, cmd.Arg3)
			mem = enlargeMem(mem, largest)
			mem[cmd.Arg1] = mem[cmd.Arg2] - mem[cmd.Arg3]
		case "BREAK!":
			break
		default:
			message += fmt.Sprintf("SPECIAL COMMAND %s\n", cmd.Command)
		}

		memcpy := make([]int, len(mem))
		copy(memcpy, mem)
		usecpy := memuseCopy(&use)
		history = append(history, instant{cmd, memcpy, *usecpy, largest, message, i})
	}
	return history
}

func cmd_push(history *[]instant, cmds *[]codegen.Assembly, mem *[]int, use *memuse, largest *int, i *int) {
	cmd := (*cmds)[*i]
	message := ""
	//fmt.Println(i)
	//fmt.Println(cmd)
	//fmt.Println(printInd(mem))

	switch cmd.Command {
	case "SET-LITERAL":
		*largest = max(*largest, cmd.Arg1)
		(*mem) = enlargeMem((*mem), *largest)
		(*mem)[cmd.Arg1] = cmd.Arg2
	case "SET-INDEXED":
		*largest = max(*largest, (*mem)[cmd.Arg1]+cmd.Arg2)
		(*mem) = enlargeMem((*mem), *largest)
		(*mem)[(*mem)[cmd.Arg1]+cmd.Arg2] = cmd.Arg3
	case "COPY-ADD":
		*largest = max(*largest, cmd.Arg1, cmd.Arg2)
		(*mem) = enlargeMem((*mem), *largest)
		(*mem)[cmd.Arg1] = (*mem)[cmd.Arg2] + cmd.Arg3
	case "COPY-INDEXED":
		*largest = max(*largest, cmd.Arg3, (*mem)[cmd.Arg1]+cmd.Arg2)
		(*mem) = enlargeMem((*mem), *largest)
		(*mem)[(*mem)[cmd.Arg1]+cmd.Arg2] = (*mem)[cmd.Arg3]
	case "DEREF":
		*largest = max(*largest, (*mem)[cmd.Arg2]+cmd.Arg3, (*mem)[cmd.Arg1])
		(*mem) = enlargeMem((*mem), *largest)
		(*mem)[cmd.Arg1] = (*mem)[(*mem)[cmd.Arg2]+cmd.Arg3]
	case "NEW":
		*largest = max(*largest, cmd.Arg1)
		(*mem) = enlargeMem((*mem), *largest)
		(*mem)[cmd.Arg1], _ = use.alloc((*mem), cmd.Arg2, nil)
		*largest = max(*largest, (*mem)[cmd.Arg1]+cmd.Arg2-1)
		(*mem) = enlargeMem((*mem), *largest)
	case "JUMP-LITERAL":
		*i = cmd.Arg1 - 1
	case "JUMP-LITERAL-IF-IS":
		if (*mem)[cmd.Arg2] == cmd.Arg3 {
			*i = cmd.Arg1 - 1
		}
	case "JUMP-LITERAL-IF-EQ":
		if (*mem)[cmd.Arg2] == (*mem)[cmd.Arg3] {
			*i = cmd.Arg1 - 1
		}
	case "JUMP":
		*i = (*mem)[cmd.Arg1] - 1
	case "JUMP-REMEMBER":
		(*mem)[(*mem)[cmd.Arg2]] = *i + 1
		*i = (*mem)[cmd.Arg1] - 1
	case "JUMP-LITERAL-REMEMBER":	// new, maybe broken
		(*mem)[(*mem)[cmd.Arg2]] = *i + 1
		*i = cmd.Arg1 - 1
	case "EXCEPTION":
		message += fmt.Sprintln("You threw an exception! Oh my gosh!")
		break
	case "ADD1":
		*largest = max(*largest, (*mem)[cmd.Arg1])
		(*mem) = enlargeMem((*mem), *largest)
		(*mem)[(*mem)[cmd.Arg1]] = (*mem)[(*mem)[cmd.Arg1]] + 1
	case "SUB1":
		*largest = max(*largest, (*mem)[cmd.Arg1])
		(*mem) = enlargeMem((*mem), *largest)
		(*mem)[(*mem)[cmd.Arg1]] = (*mem)[(*mem)[cmd.Arg1]] - 1
	case "ADD":
		*largest = max(*largest, cmd.Arg1, cmd.Arg2, cmd.Arg3)
		(*mem) = enlargeMem((*mem), *largest)
		(*mem)[cmd.Arg1] = (*mem)[cmd.Arg2] + (*mem)[cmd.Arg3]
	case "SUB":
		*largest = max(*largest, cmd.Arg1, cmd.Arg2, cmd.Arg3)
		(*mem) = enlargeMem((*mem), *largest)
		(*mem)[cmd.Arg1] = (*mem)[cmd.Arg2] - (*mem)[cmd.Arg3]
	case "BREAK!":
		break
	default:
		message += fmt.Sprintf("SPECIAL COMMAND %s\n", cmd.Command)
	}
	memcpy := make([]int, len((*mem)))
	copy(memcpy, (*mem))
	usecpy := memuseCopy(use)
	*history = append(*history, instant{cmd, memcpy, *usecpy, *largest, message, *i})
	(*i)++
}

func interactive_interpret(cmds []codegen.Assembly) {
	mem := make([]int, 20, 1000)
	rest := memuse{false, 13, 1000, nil}
	use := memuse{true, 0, 13, &rest}
	largest := 0
	history := make([]instant, 0, 400)
	position := 0

	var oldCount int = 1
	var index int = 0
	var top int = 0
	var last byte = 'j'
	var b []byte = make([]byte, 1)
	var s string = ""
	fmt.Printf("Line %d, pos %d: %v\n", index, 0, cmds[0])
	cmd_push(&history, &cmds, &mem, &use, &largest, &position)
	for {
		os.Stdin.Read(b)
		if b[0] >= byte('0') && b[0] <= byte('9') {
			s += string(b)
			continue
		}
		c, _ := strconv.Atoi(s)
		s = ""
		if c == 0 {
			c++
		}
		fmt.Println()

		exec:
		for i := 0; i < c; i++ {
			switch (b[0]) {
			case 'j':
				if index < len(history)-1 {
					index++
				} else if (position < len(cmds)) {
					index++
					if index > top {
						cmd_push(&history, &cmds, &mem, &use, &largest, &position)
						top++
					}
				}
			case 'k':
				if (index > 0) {
					index--
				}
			case 'p':
				fmt.Println(printmem(history[index].mem, &history[index].use, ""))
			case 'c':
				fmt.Println(history[index].command)
			case 'u':
				history[index].use.printmemuse()
			case 'g':
				if 0 > c {
					index = 0
				} else if index > c {
					index = c
				} else {
					for {
						if index < len(history)-1 && index < c {
							index++
						} else if (position < len(cmds) && index < c) {
							index++
							if index > top {
								cmd_push(&history, &cmds, &mem, &use, &largest, &position)
								top++
							}
						} else {
							break
						}
					}
				}
				goto loopBreak
			default:
				b[0] = last
				c = oldCount
				goto exec
			}
		}
		loopBreak:
		fmt.Printf("Line %d, pos %d: %v\n", index, history[index].position, history[index].command)
		oldCount = c
		last = b[0]
	}
}
