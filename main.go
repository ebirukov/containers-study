package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	if len(os.Args) < 2 {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("cyr dir %s", dir)
		err = os.Chdir("/tmp/container")
		if err != nil {
			log.Fatal(err)
		}

		err = syscall.Chroot("/tmp/container")
		if err != nil {
			log.Fatal(err)
		}

		syscall.Mount("proc", "/proc", "proc", 0, "")
		time.Sleep(time.Hour)
		log.Fatal("Usage: go run main.go <command>")
	}

	os.Args = os.Args[1:]

	cmd := exec.Command(os.Args[0], os.Args[1:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("cur dir %s", dir)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Chroot:     dir, //"/tmp/container",
		Credential: nil,
		Setsid:     true,
		Setpgid:    false,
		Ctty:       0,
		Foreground: false,
		Pgid:       0,
		Pdeathsig:  0,
		Cloneflags: syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWTIME |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWCGROUP |
			syscall.CLONE_NEWUTS,
		Unshareflags: syscall.CLONE_NEWNS,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
		GidMappingsEnableSetgroups: false,
		AmbientCaps:                nil,
		UseCgroupFD:                false,
		CgroupFD:                   0,
		PidFD:                      nil,
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	log.Println("run ", os.Args, "child PID=", cmd.Process.Pid, "runner PID=", os.Getpid())

	_, err = cmd.Process.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
