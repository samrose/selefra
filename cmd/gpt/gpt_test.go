package apply

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
	err := Gpt(context.Background(), "Please help me analyze the vulnerabilities in AWS S3?", projectWorkspace, downloadWorkspace)
	assert.Nil(t, err)
}
