package apply

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApply(t *testing.T) {
	//projectWorkspace := "D:\\workspace\\module-mock-test"
	projectWorkspace := "./test_data/test_query_module"
	//projectWorkspace := "D:\\selefra\\workplace\\sfslack-v2-bak"
	downloadWorkspace := "./test_download"
	err := Apply(context.Background(), projectWorkspace, downloadWorkspace)
	assert.Nil(t, err)
}
