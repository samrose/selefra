package test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Test(t *testing.T) {
	projectWorkspace := "./test_data/module_test"
	downloadWorkspace := "./test_download"
	err := Test(context.Background(), projectWorkspace, downloadWorkspace)
	assert.Nil(t, err)
}
