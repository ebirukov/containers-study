package proc

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func Cmdline(pid int) []string {
	cmdline := mustRead(fmt.Sprintf("/proc/%d/cmdline", pid))
	return strings.Fields(strings.Trim(cmdline, "[]"))
}

func MountProcFS() error {
	if _, err := os.Stat("/proc"); errors.Is(err, fs.ErrNotExist) {
		os.Mkdir("/proc", 0755)

		return syscall.Mount("proc", "/proc", "proc", 0, "")
	}

	// /proc/[pid]/ns/pid — симлинк на идентификатор PID namespace процесса в текущем mount namespace
	// /proc/self/ns/pid — симлинк на идентификатор PID namespace текущего процесса
	// не совпадают, значит mount namespace наследован и /proc смотрит на объекты ядра родительского PID namespace
	if SelfNamespace("pid") != Namespace(os.Getpid(), "pid") {
		log.Printf("remount namespace\n")

		return errors.Join(
			syscall.Unmount("/proc", syscall.MNT_DETACH),
			syscall.Mount("proc", "/proc", "proc", 0, ""),
		)
	}

	return nil
}

func Mounts(pid int) []string {
	mount := mustRead(fmt.Sprintf("/proc/%d/mounts", pid))
	return strings.Split(strings.TrimSpace(mount), "\n")
}

func SelfMounts() []string {
	mount := mustRead("/proc/mounts")
	return strings.Split(strings.TrimSpace(mount), "\n")
}

func Namespace(pid int, nsType string) string {
	ns, err := os.Readlink(fmt.Sprintf("/proc/%d/ns/%s", pid, nsType))
	if err != nil {
		panic(err)
	}

	return ns
}

func SelfNamespace(nsType string) string {
	ns, err := os.Readlink(fmt.Sprintf("/proc/self/ns/%s", nsType))
	if err != nil {
		panic(err)
	}

	return ns
}

func Pids(match func(int) bool) (pids []int) {
	filepath.WalkDir("/proc", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			return filepath.SkipDir
		}

		pid, err := strconv.Atoi(d.Name())
		if errors.Is(err, strconv.ErrSyntax) {
			// Пропускаем не-числовые директории
			return nil
		}

		if match(pid) {
			pids = append(pids, pid)
		}

		return nil
	})

	return pids
}

func mustRead(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return strings.Trim(string(b), "\\x00")
}
