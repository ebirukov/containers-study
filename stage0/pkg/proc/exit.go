package proc

import (
	"log"
	"os"
	"syscall"
)

func Exit(code int) {
	if err := recover(); err != nil {
		log.Fatalln(err)

		code = 1
	}

	if os.Getpid() == 1 {
		log.Print("power off on exit\n")
		syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
	}

	if code != 0 {
		log.Printf("exit with code %d\n", code)
	}

	os.Exit(code)
}
