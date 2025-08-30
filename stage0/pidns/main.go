package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"stage0/pkg/proc"
	"strconv"
	"syscall"
)

//go:generate go build -o ../pidns-amd64 main.go

func main() {
	proc.MountProcFS()

	selfNS := proc.SelfNamespace("pid")
	log.SetPrefix(fmt.Sprintf("[PID %d, PPID %d, %s]: ", os.Getpid(), os.Getppid(), selfNS))
	log.SetFlags(log.Lmsgprefix)

	log.Printf("process started with args: %s", os.Args)

	log.Printf("list of processes: %v\n", proc.Pids(nonKernelPid))

	var sigPID int
	flag.IntVar(&sigPID, "pid", 0, "PID")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	if sigPID > 0 {
		log.Printf("send signal %s to process with PID %d\n", syscall.SIGTERM, sigPID)

		if err := syscall.Kill(sigPID, syscall.SIGTERM); err != nil {
			log.Fatalf("can't send signal to process with PID %d: %v", sigPID, err)
		}

		proc.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-pid", strconv.Itoa(os.Getpid()))
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID,
	}

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			log.Fatalf("failed execute command: %v", err)
		}

		if cmd.ProcessState.ExitCode() < 0 {
			log.Printf("process %d terminated with status %d: %v", cmd.ProcessState.Pid(), cmd.ProcessState.ExitCode(), err)
		}

		proc.Exit(exitErr.ProcessState.ExitCode())
	}

	select {
	case sig := <-sigs:
		log.Printf("received signal %s\n", sig)
		proc.Exit(0)
	}
}

func nonKernelPid(pid int) bool {
	return len(proc.Cmdline(pid)) > 0
}
