package cli_ui

import (
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/cloud_sdk"
	"github.com/selefra/selefra/pkg/utils"
	"strings"
)

// ------------------------------------------------- --------------------------------------------------------------------

// CloudTokenRequestPath What is the request path to obtain the cloud token
// If there is a change in the address of the cloud side, synchronize it here
const CloudTokenRequestPath = "/Settings/accessTokens"

// InputCloudToken Guide the user to enter a cloud token
func InputCloudToken(serverUrl string) (string, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	tipsTemplate := `selefra will login https://app.selefra.io in your default browser.
if login is successful, the token will be stored as a plain text file for future usage:

   Enter your access token from https://app.selefra.io{{.CloudTokenRequestPath}}
   or hit <ENTER> to complete login in browser:`

	// Render display tips
	data := make(map[string]string)
	data["CloudTokenRequestPath"] = CloudTokenRequestPath
	inputCloudTokenTips, err := utils.RenderingTemplate("input-token-tips-template", tipsTemplate, data)
	if err != nil {
		return "", diagnostics.AddErrorMsg("input-token-tips-template render error: %s", err.Error())
	}
	fmt.Println(inputCloudTokenTips)

	// Open a browser window to allow the user to log in
	_, _, _ = utils.OpenBrowser("https://app.selefra.io" + CloudTokenRequestPath)

	// Read the token entered by the user
	var rawToken string
	_, err = fmt.Scanln(&rawToken)
	//reader := bufio.NewReader(os.Stdin)
	//rawToken, err := reader.ReadString('\n')
	if err != nil {
		return "", diagnostics.AddErrorMsg("Input cloud token error: %s", err.Error())
	}
	cloudToken := strings.TrimSpace(strings.Replace(rawToken, "\n", "", -1))
	if cloudToken == "" {
		return "", diagnostics.AddErrorMsg("No token provided")
	}

	return cloudToken, diagnostics
}

// ShowLoginSuccess The CLI prompt indicating successful login is displayed
func ShowLoginSuccess(serverUrl string, cloudCredentials *cloud_sdk.CloudCredentials) {
	loginSuccessTemplate := `
Retrieved token for user: {{.UserName}}.
Welcome to Selefra CloudClient!
Logged in to selefra as {{.UserName}} (https://{{.ServerHost}}/{{.OrgName}})
`
	template, err := utils.RenderingTemplate("login-success-tips-template", loginSuccessTemplate, cloudCredentials)
	if err != nil {
		Errorf("render login success message error: %s\n", err.Error())
		return
	}
	Successf(template)
}

// ShowLoginFailed Displays a login failure message
func ShowLoginFailed(cloudToken string) {
	Errorf("You input token %s login failed \n", cloudToken)
}

// ------------------------------------------------- --------------------------------------------------------------------

// ShowRetrievedCloudCredentials Displays the results of the local retrieval of login credentials
func ShowRetrievedCloudCredentials(cloudCredentials *cloud_sdk.CloudCredentials) {
	if cloudCredentials == nil {
		return
	}
	Successf(fmt.Sprintf("Auto login with user %s \n", cloudCredentials.UserName))
}

// ------------------------------------------------- --------------------------------------------------------------------

// ShowLogout Display the logout success prompt
func ShowLogout(cloudCredentials *cloud_sdk.CloudCredentials) {
	if cloudCredentials == nil {
		return
	}
	Successf(fmt.Sprintf("User %s logout success \n", cloudCredentials.UserName))
}

// ------------------------------------------------- --------------------------------------------------------------------
