package utils

import (
	"errors"
	"github.com/mitchellh/go-homedir"
	"os"
	"path/filepath"
)

// Home return selefra home, config in selefra home, an error
// selefra is in ~/.selefra, it store tokens, downloaded binary files, database files, and other configuration files, etc.
// configPath is ~/.selefra/.path/config.json, in config.json, the absolute path of the provider binary is declared
func Home() (homeDir string, configPath string, err error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", "", err
	}
	registryPath := filepath.Join(home, ".selefra")
	_, err = os.Stat(registryPath)
	if errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(registryPath, 0755)
		if err != nil {
			return "", "", err
		}
	}

	// provider binary file store in providerPath
	providerPath := filepath.Join(home, ".selefra", ".path")

	_, err = os.Stat(providerPath)
	if errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(providerPath, 0755)
		if err != nil {
			return "", "", err
		}
	}

	config := filepath.Join(home, ".selefra", ".path", "config.json")

	_, err = os.Stat(config)
	if errors.Is(err, os.ErrNotExist) {
		err = os.WriteFile(config, []byte("{}"), 0644)
		if err != nil {
			return "", "", err
		}
	}
	return registryPath, config, nil
}

func GetTempPath() (string, error) {
	path, _, err := Home()
	if err != nil {
		return "", err
	}
	ociPath := filepath.Join(path, "temp")
	_, err = os.Stat(ociPath)
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(ociPath, 0755)
		if err != nil {
			return "", err
		}
	}
	return ociPath, nil
}

func CreateSource(path, version, latest string) (string, string) {
	if latest == "latest" {
		return "selefra/" + path + "@" + version, "selefra/" + path + "@latest"
	}
	return "selefra/" + path + "@" + version, ""
}
