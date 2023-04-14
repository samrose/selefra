//go:build !windows
// +build !windows

package utils

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func getTerminalWidth() (int, error) {
	var size [4]uint16
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&size)))
	if err != 0 {
		return 0, fmt.Errorf("failed to get terminal width: %v", err)
	}
	return int(size[1]), nil
}
