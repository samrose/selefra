package gpt

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGpt(t *testing.T) {
	//projectWorkspace := "D:\\workspace\\module-mock-test"
	projectWorkspace := "./test_data/test_query_module"
	//projectWorkspace := "D:\\selefra\\workplace\\sfslack-v2-bak"
	downloadWorkspace := "./test_download"
	Instructions := make(map[string]interface{})
	Instructions["query"] = "Please help me analyze the vulnerabilities in AWS S3?"
	Instructions["openai_api_key"] = "xx"
	Instructions["openai_mode"] = "gpt-3.5"
	Instructions["openai_limit"] = uint64(10)
	err := Gpt(context.Background(), Instructions, projectWorkspace, downloadWorkspace)
	assert.Nil(t, err)
}
