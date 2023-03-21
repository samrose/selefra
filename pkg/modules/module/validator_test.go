package module

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseDuration(t *testing.T) {
	s := "1d"
	duration, err := ParseDuration(s)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, duration)
}
