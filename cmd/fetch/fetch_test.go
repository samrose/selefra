package fetch

import (
	"testing"
)

func TestFetch(t *testing.T) {
	projectWorkspace := "./test_data/test_fetch_module"
	downloadWorkspace := "./test_download"
	Fetch(projectWorkspace, downloadWorkspace)
}
