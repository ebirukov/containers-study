package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"stage0/pkg/proc"
	"strconv"
	"syscall"
)

//go:generate go build -o ../cgroup-amd64 main.go

const (
	arrSize  = 1 << 17
	cgroupFS = "/sys/fs/cgroup"
)

var arr [arrSize]int

func main() {
	defer proc.Exit(0)

	log.SetPrefix(fmt.Sprintf("[PPID %d, PID %d]: ", os.Getppid(), os.Getpid()))
	log.SetFlags(log.Lmsgprefix)

	log.Printf("process started with args: %s", os.Args)

	var cgroupName string
	flag.StringVar(&cgroupName, "cgroup", filepath.Base(os.Args[0]), "cgroup name")
	flag.Parse()

	// путь к контрольной группе в cgroupfs
	cgroupPath := filepath.Join(cgroupFS, cgroupName)

	if _, err := os.Stat(cgroupPath); os.IsNotExist(err) {
		// Создаем путь для монтирования файловой системы cgroup
		must(os.MkdirAll(cgroupFS, 0700))
		// Монтировуем cgroup v2
		must(syscall.Mount("none", cgroupFS, "cgroup2", 0, ""))
		// Включаем контроллеры "pids" и "memory" в cgroup
		must(os.WriteFile(filepath.Join(cgroupFS, "cgroup.subtree_control"), []byte("+pids +memory"), 0644))
		// При создании директории cgroup-amd64 в файловой системе cgroup,
		// ядро инициализирует файлы для управления контрольными группами в созданной директории
		must(os.MkdirAll(cgroupPath, 0755))
		// Устанавливаем лимит на 20 процессов в созданной группе
		must(os.WriteFile(filepath.Join(cgroupPath, "pids.max"), []byte("10"), 0644))
		// Устанавливаем для созданной группе лимит на 8М доступной для процесса памяти
		must(os.WriteFile(filepath.Join(cgroupPath, "memory.max"), []byte(strconv.Itoa(1<<23)), 0644))
	}

	// Добавляет PID текущего процесса в группу для применения лимитов группы к нему
	must(os.WriteFile(filepath.Join(cgroupPath, "cgroup.procs"), []byte(strconv.Itoa(os.Getppid())), 0644))
	//Выделяем массив в памяти
	arr = [arrSize]int{}

	//запускаем дочерний процесс в контрольной группе
	cmd := exec.Command(os.Args[0], "-cgroup", cgroupName)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin

	if err := cmd.Start(); err != nil {
		log.Fatalf("failed to start command: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Printf("command finished with error: %v\n", err)
	}

	proc.Exit(cmd.ProcessState.ExitCode())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
