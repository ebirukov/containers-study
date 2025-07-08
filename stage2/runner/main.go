package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	log.Println("run ", os.Args, "child PID=", os.Getpid())

	if len(os.Args) < 2 {
		fmt.Println("Usage: ./main <command>")
		os.Exit(1)
	}

	cmd := exec.Command(os.Args[1], os.Args[2:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Chroot:                     "",
		Credential:                 nil,
		Ptrace:                     false,
		Setsid:                     false,
		Setpgid:                    false,
		Setctty:                    false,
		Noctty:                     false,
		Ctty:                       0,
		Foreground:                 false,
		Pgid:                       0,
		Pdeathsig:                  0,
		Cloneflags:                 0,
		Unshareflags:               syscall.CLONE_NEWUSER,
		UidMappings:                nil,
		GidMappings:                nil,
		GidMappingsEnableSetgroups: false,
		AmbientCaps:                nil,
		UseCgroupFD:                false,
		CgroupFD:                   0,
		PidFD:                      nil,
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("Error run command: %v", err)
	}

	host, err := os.Hostname()
	if err != nil {
		log.Printf("Error getting hostname: %v", err)
	}

	log.Printf("Run pid=%d on host %s", cmd.Process.Pid, host)

	if err := cmd.Process.Release(); err != nil {
		log.Fatalf("Error release command: %v", err)
	}
}
