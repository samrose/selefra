package cloud_sdk

import (
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/config"
	"github.com/selefra/selefra/pkg/utils"
	"path/filepath"
	"time"
)

const CredentialsFileName = "credentials.json"

// ------------------------------------------------- --------------------------------------------------------------------

// CloudCredentials Credentials to connect to the selefra cloud
type CloudCredentials struct {

	// The name of the token above is easy to remember
	// This name is set when the token is created on the cloud side
	TokenName string `json:"token_name"`

	// token used for authentication
	Token string `json:"token"`

	UserName string `json:"user_name"`

	OrgName string `json:"org_name"`

	ServerHost string `json:"server_host"`

	// Time of the last login
	LoginTime time.Time `json:"login_time"`

	// The last time the token was used
	LastUseTime time.Time `json:"last_login_time"`
}

// ------------------------------------------------- --------------------------------------------------------------------

// GetCredentialsWorkspacePath get the Credentials save directory
func (x *CloudClient) GetCredentialsWorkspacePath() (string, *schema.Diagnostics) {
	path, err := config.GetSelefraHomeWorkspacePath()
	if err != nil {
		return "", schema.NewDiagnostics().AddError(err)
	}
	return filepath.Join(path, CredentialsFileName), nil
}

// SaveCredentials Save the login credentials to the local directory
func (x *CloudClient) SaveCredentials(credentials *CloudCredentials) *schema.Diagnostics {

	selefraHomeDirectory, err := config.GetSelefraHomeWorkspacePath()
	if err != nil {
		return schema.NewDiagnostics().AddErrorMsg("get selefra home directory error: %s", err.Error())
	}
	err = utils.EnsureDirectoryExists(selefraHomeDirectory)
	if err != nil {
		return schema.NewDiagnostics().AddErrorMsg("create directory %s error: %s", selefraHomeDirectory, err.Error())
	}

	path, d := x.GetCredentialsWorkspacePath()
	if utils.HasError(d) {
		return d
	}

	err = utils.WriteJsonFile[*CloudCredentials](path, credentials)
	if err != nil {
		return schema.NewDiagnostics().AddError(err)
	}
	return nil
}

// GetCredentials Read the credentials stored in the local directory
func (x *CloudClient) GetCredentials() (*CloudCredentials, *schema.Diagnostics) {
	path, d := x.GetCredentialsWorkspacePath()
	if utils.HasError(d) {
		return nil, d
	}
	credentials, err := utils.ReadJsonFile[*CloudCredentials](path)
	if err != nil {
		return nil, schema.NewDiagnostics().AddError(err)
	}
	if credentials.Token == "" {
		return nil, nil
	}
	return credentials, nil
}

// ------------------------------------------------- --------------------------------------------------------------------
