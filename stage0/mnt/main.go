package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"stage0/pkg/proc"
	"syscall"
)

//go:generate go build -o ../mnt-amd64 main.go

func init() {
	log.SetPrefix(fmt.Sprintf("[PPID %d, PID %d]: ", os.Getppid(), os.Getpid()))
	log.SetFlags(log.Lmsgprefix)
}

func main() {
	defer proc.Exit(0)

	log.Printf("process started with args: %s", os.Args)

	var action string
	flag.StringVar(&action, "action", "", "action")
	flag.Parse()

	switch action {
	case "unmount":
		err := syscall.Unmount("/sys", syscall.MNT_DETACH)
		if err != nil {
			log.Fatalf("Unmount failed: %s", err)
		}

		return
	}

	err := proc.MountProcFS()
	if err != nil {
		log.Fatalf("mount procfs failed: %s", err)
	}

	os.Mkdir("/sys", 0755)
	if err := syscall.Mount("sys", "/sys", "sysfs", 0, ""); err != nil {
		log.Fatalf("mount sysfs failed: %s", err)
	}

	log.Printf("num of mount points before unmount: %d", len(proc.SelfMounts()))
	for _, m := range proc.SelfMounts() {
		log.Println(m)
	}

	cmd := exec.Command(os.Args[0], "-action", "unmount")
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	log.Printf("num of mount points after finished unmount command: %d", len(proc.SelfMounts()))
	for _, m := range proc.SelfMounts() {
		log.Println(m)
	}
}
