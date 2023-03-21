package selefra_workspace

import "github.com/selefra/selefra/config"

// GetSelefraWorkspaceDirectory Gets the path of the workspace for selefra
func GetSelefraWorkspaceDirectory() (string, error) {
	// TODO Migrate the concrete implementation here
	return config.GetSelefraHomeWorkspacePath()
}
