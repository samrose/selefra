package module_loader

import (
	"testing"
)

func TestNewModuleLoaderBySource(t *testing.T) {
	source := NewModuleLoaderBySource("rules-aws-misconfigure-s3@v0.0.1")
	t.Log(source)
}
