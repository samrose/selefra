package selefra_workspace

import (
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-utils/pkg/id_util"
	"github.com/selefra/selefra/pkg/logger"
	"github.com/selefra/selefra/pkg/utils"
	"path/filepath"
	"sync"
)

// DeviceInformation Stores information about the current device
// We will save a device file in the working directory of selefra.
// This file is used for coordination in case of distributed conflicts.
// Please do not manually edit or delete this file, otherwise it may cause program errors
type DeviceInformation struct {

	// This is the only ID available, This is for distributed collaboration with other nodes
	// There is no device information collection at present, and there will not be any in the future.
	// I hope to make a real open source and mutual assistance software, and my boss also thinks so
	ID string `json:"id"`
}

var deviceInformationOnce sync.Once

// GetDeviceID Gets the ID of this device
func GetDeviceID() (string, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	// Ensure that the device file exists
	deviceInformationOnce.Do(func() {
		d := EnsureDeviceIDExists()
		if utils.HasError(d) {
			logger.ErrorF("EnsureDeviceIDExists error: %s", d.String())
		}
	})

	path, err := GetDeviceInformationFilePath()
	if err != nil {
		return "", diagnostics.AddErrorMsg("get device file path error: %s", err.Error())
	}

	information, err := utils.ReadJsonFile[*DeviceInformation](path)
	if err != nil {
		return "", diagnostics.AddErrorMsg("read device file error: %s", err.Error())
	}
	if information == nil || information.ID == "" {
		return "", diagnostics.AddErrorMsg("device id not found")
	}
	return information.ID, nil
}

// EnsureDeviceIDExists Ensure that the device file exists
func EnsureDeviceIDExists() *schema.Diagnostics {

	diagnostics := schema.NewDiagnostics()

	path, err := GetDeviceInformationFilePath()
	if err != nil {
		return diagnostics.AddErrorMsg("get device file path error: %s", err.Error())
	}

	// If the device file already exists, it is not generated again
	information, err := utils.ReadJsonFile[*DeviceInformation](path)
	if err == nil || (information != nil && information.ID != "") {
		return nil
	}

	information = &DeviceInformation{
		ID: id_util.RandomId(),
	}
	err = utils.WriteJsonFile(path, information)
	if err != nil {
		return diagnostics.AddErrorMsg("write device file error: %s", err.Error())
	}
	return diagnostics
}

// GetDeviceInformationFilePath Obtain the directory for storing device files
func GetDeviceInformationFilePath() (string, error) {
	selefraHomeWorkspace, err := GetSelefraWorkspaceDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(selefraHomeWorkspace, "device.json"), nil
}
