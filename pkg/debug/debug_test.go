package debug

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewSamplingService(t *testing.T) {
	service := NewSamplingService("./test_download", time.Second*30)
	err := service.Start()
	assert.Nil(t, err)

	time.Sleep(time.Minute * 30)
	service.Stop()
}
