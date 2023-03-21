package provider

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

//import (
//	"context"
//	"github.com/selefra/selefra/global"
//	"testing"
//)
//
//func TestRemove(t *testing.T) {
//	global.Init("TestRemove", global.WithWorkspace("../../tests/workspace/offline"))
//	err := Remove([]string{"aws"})
//	if err != nil {
//		t.Error(err)
//	}
//	err = install(context.Background(), []string{"aws@latest"})
//	if err != nil {
//		t.Error(err)
//	}
//}

func TestRemove(t *testing.T) {

	provider := "mock@v0.0.3"

	err := Install(context.Background(), "./test_download", provider)
	assert.Nil(t, err)

	err = Remove(context.Background(), "./test_download", provider)
	assert.Nil(t, err)
}
