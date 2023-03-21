package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func Exists(filepath string) bool {
	_, err := os.Stat(filepath)
	if err != nil {
		return false
	}
	return true
}

func ExistsFile(filepath string) bool {
	info, err := os.Stat(filepath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func ExistsDirectory(directoryPath string) bool {
	info, err := os.Stat(directoryPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// EnsureDirectoryExists Make sure the directory exists, and create it if it does not
func EnsureDirectoryExists(directoryPath string) error {
	_, err := os.Stat(directoryPath)
	if errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(directoryPath, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func EnsureDirectoryNotExists(directoryPath string) error {
	_, err := os.Stat(directoryPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return os.RemoveAll(directoryPath)
}

// EnsureFileExists Make sure the file exists, and if it does not, use the given content to create the file
func EnsureFileExists(fileFullPath string, initBytes []byte) error {

	_, err := os.Stat(fileFullPath)

	if err == nil {
		return nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	err = EnsureDirectoryExists(filepath.Dir(fileFullPath))
	if err != nil {
		return err
	}

	return os.WriteFile(fileFullPath, initBytes, os.ModePerm)
}

func ReadYamlFile[T any](yamlFilePath string) (T, error) {
	//var r T
	//yamlFileReader, err := os.Open(yamlFilePath)
	//if err != nil {
	//	return r, fmt.Errorf("open file %s error: %s", yamlFilePath, err.Error())
	//}
	//config := viper.New()
	//config.AddConfigPath(yamlFilePath)
	//configType := strings.Replace(path.Ext(yamlFilePath), ".", "", 1)
	//config.SetConfigType(configType)
	//err = config.ReadConfig(yamlFileReader)
	//if err != nil {
	//	return r, fmt.Errorf("read yaml file %s error: %s", yamlFilePath, err.Error())
	//}
	//err = config.Unmarshal(&r)
	//if err != nil {
	//	return r, fmt.Errorf("unmarshal yaml file %s error: %s", yamlFilePath, err.Error())
	//}
	//return r, nil

	var r T
	yamlFileBytes, err := os.ReadFile(yamlFilePath)
	if err != nil {
		return r, fmt.Errorf("open file %s error: %s", yamlFilePath, err.Error())
	}
	err = yaml.Unmarshal(yamlFileBytes, &r)
	if err != nil {
		return r, fmt.Errorf("unmarshal yaml file %s error: %s", yamlFilePath, err.Error())
	}
	return r, nil
}

func ReadJsonFile[T any](jsonFilePath string) (T, error) {
	var r T
	jsonBytes, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return r, fmt.Errorf("open file %s error: %s", jsonFilePath, err.Error())
	}

	err = json.Unmarshal(jsonBytes, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}

func WriteJsonFile[T any](jsonFilePath string, v T) error {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(jsonFilePath, jsonBytes, os.ModePerm)
}

func AbsPath(path string) string {
	abs, _ := filepath.Abs(path)
	return abs
}
