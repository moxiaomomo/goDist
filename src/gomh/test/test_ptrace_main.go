package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"text/tabwriter"

	"github.com/seccomp/libseccomp-golang"
)

type syscallCounter []int

const maxSyscalls = 303

func (s syscallCounter) init() syscallCounter {
	s = make(syscallCounter, maxSyscalls)
	return s
}

func (s syscallCounter) inc(syscallID uint64) error {
	if syscallID > maxSyscalls {
		return fmt.Errorf("invalid syscall ID (%x)", syscallID)
	}

	s[syscallID]++
	return nil
}

func (s syscallCounter) print() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', tabwriter.AlignRight|tabwriter.Debug)
	for k, v := range s {
		if v > 0 {
			name, _ := seccomp.ScmpSyscall(k).GetName()
			fmt.Fprintf(w, "%d\t%s\n", v, name)
		}
	}
	w.Flush()
}

func (s syscallCounter) getName(syscallID uint64) string {
	name, _ := seccomp.ScmpSyscall(syscallID).GetName()
	return name
}

func main() {
	var err error
	var regs syscall.PtraceRegs
	var ss syscallCounter
	ss = ss.init()

	fmt.Println("Run: ", os.Args[1:])

	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Ptrace: true,
	}

	cmd.Start()
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("Wait err %v \n", err)
	}

	pid := cmd.Process.Pid
	exit := true

	for {
		// 记得 PTRACE_SYSCALL 会在进入和退出syscall时使 tracee 暂停，所以这里用一个变量控制，RAX的内容只打印一遍
		if exit {
			err = syscall.PtraceGetRegs(pid, &regs)
			if err != nil {
				break
			}
			//fmt.Printf("%#v \n",regs)
			name := ss.getName(regs.Orig_rax)
			fmt.Printf("name: %s, id: %d \n", name, regs.Orig_rax)
			ss.inc(regs.Orig_rax)
		}
		// 上面Ptrace有提到的一个request命令
		err = syscall.PtraceSyscall(pid, 0)
		if err != nil {
			panic(err)
		}
		// 猜测是等待进程进入下一个stop，这里如果不等待，那么会打印大量重复的调用函数名
		_, err = syscall.Wait4(pid, nil, 0, nil)
		if err != nil {
			panic(err)
		}

		exit = !exit
	}

	ss.print()
}
