package utils

import (
	"fmt"
	"strings"
)

// ParseGitHubRepoFullName Parses the full name of GitHub's warehouse into owner and repoName parts
// example: selefra/registry
func ParseGitHubRepoFullName(repoFullName string) (owner, repo string, err error) {
	split := strings.Split(repoFullName, "/")
	if len(split) != 2 {
		return "", "", fmt.Errorf("%s is not a valid GitHub repository full name", repoFullName)
	}
	return split[0], split[1], nil
}
