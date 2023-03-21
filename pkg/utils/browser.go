package utils

import (
	"errors"
	"runtime"
)

// OpenBrowser Open the given URL with your browser
func OpenBrowser(targetUrl string) (stdout string, stderr string, err error) {
	switch runtime.GOOS {
	case "windows":
		return RunCommand("cmd", "/c", "start", targetUrl)
	case "linux":
		return RunCommand("xdg-open", targetUrl)
	case "darwin":
		return RunCommand("open", targetUrl)
	default:
		return "", "", errors.New("open browser not supported on this platform")
	}
}
