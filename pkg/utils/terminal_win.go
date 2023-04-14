//go:build windows
// +build windows

package utils

import (
	"fmt"
	"golang.org/x/sys/windows"
	"os"
)

func getTerminalWidth() (int, error) {
	var info windows.ConsoleScreenBufferInfo
	err := windows.GetConsoleScreenBufferInfo(windows.Handle(os.Stdout.Fd()), &info)
	if err != nil {
		return 0, fmt.Errorf("failed to get terminal width: %v", err)
	}
	return int(info.Window.Right - info.Window.Left + 1), nil
}
