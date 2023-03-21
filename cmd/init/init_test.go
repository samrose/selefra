package init

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/env"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_getDsn(t *testing.T) {
	projectWorkspace := "./test_data"
	downloadWorkspace := "./test_download"
	os.Setenv(env.DatabaseDsn, "")
	dsn, err := getDsn(context.Background(), projectWorkspace, downloadWorkspace)
	assert.Nil(t, err)
	assert.NotEmpty(t, dsn)
}
