package utils

import (
	"fmt"
	"testing"
)

func TestOpenBrowser(t *testing.T) {
	stdout, stderr, diagnostics := OpenBrowser("https://google.com")
	fmt.Println(stderr)
	fmt.Println(stdout)
	fmt.Println(diagnostics.Error())
}
