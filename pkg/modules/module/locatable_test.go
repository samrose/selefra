package module

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_baseYamlSelector(t *testing.T) {
	s := baseYamlSelector("foo.bar._key")
	assert.Equal(t, "foo.bar", s)

	s = baseYamlSelector("foo.bar[1]")
	assert.Equal(t, "foo.bar", s)

	s = baseYamlSelector("f")
	assert.Equal(t, "", s)
}
