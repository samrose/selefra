package telemetry

import (
	"context"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSubmit(t *testing.T) {
	d := Submit(context.Background(), NewEvent("do-something").Add("foo", "bar"))
	if utils.IsNotEmpty(d) {
		t.Log(d.String())
	}
	assert.False(t, utils.HasError(d))

	Close(context.Background())
}
