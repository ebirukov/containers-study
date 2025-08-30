package main

import (
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

//go:generate go build -o ../nsproc-amd64 main.go

func main() {
	defer proc.Exit(0)

	log.SetPrefix(fmt.Sprintf("[PPID %d, PID %d]: ", os.Getppid(), os.Getpid()))
	log.SetFlags(log.Lmsgprefix)

	var sigPID int
	flag.IntVar(&sigPID, "pid", 0, "PID")
	flag.Parse()

	log.Printf("process started with args: %s", os.Args)

	if sigPID > 0 {
		if err := syscall.Kill(sigPID, syscall.SIGTERM); err != nil {
			log.Fatal(err)
		}

		log.Printf("send signal %s to process with PID %d\n", syscall.SIGTERM, sigPID)

		return
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	cmd := exec.Command(os.Args[0], "-pid", strconv.Itoa(os.Getpid()))
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	select {
	case sig := <-sigs:
		log.Printf("received signal %s\n", sig)
	}
}
