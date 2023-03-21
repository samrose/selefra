package local_modules_manager

// LocalModuleManager Manage the cache of locally downloaded modules
type LocalModuleManager struct {
	selefraHomeWorkspace string
	projectWorkspace     string
	downloadWorkspace    string
}

func NewLocalModuleManager() *LocalModuleManager {
	return &LocalModuleManager{}
}
