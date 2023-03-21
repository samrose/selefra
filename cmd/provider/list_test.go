package provider

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

//import (
//	"github.com/selefra/selefra/global"
//	"testing"
//)
//
//func TestList(t *testing.T) {
//	global.Init("TestList", global.WithWorkspace("../../tests/workspace/offline"))
//	err := list()
//	if err != nil {
//		t.Error(err)
//	}
//}

func Test_list(t *testing.T) {
	err := List("./test_download")
	assert.Nil(t, err)
}
