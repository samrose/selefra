package login

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRunFunc(t *testing.T) {
	err := RunFunc(nil, nil)
	assert.Nil(t, err)
}
