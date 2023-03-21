package cloud_sdk

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/grpc/pb/cloud"
	"github.com/selefra/selefra/pkg/grpc/pb/common"
	"time"
)

// ------------------------------------------------- --------------------------------------------------------------------

// IsLoggedIn Check whether the login status is current
func (x *CloudClient) IsLoggedIn() bool {
	return x.token != ""
}

func (x *CloudClient) SetToken(token string) {
	x.token = token
}

// ------------------------------------------------- --------------------------------------------------------------------

// Login Try to log in with the given token
func (x *CloudClient) Login(token string) (*CloudCredentials, *schema.Diagnostics) {
	diagnostics := schema.NewDiagnostics()
	if token == "" {
		return nil, diagnostics.AddErrorMsg("Token can not be empty for login")
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()
	response, err := x.cloudNoAuthClient.Login(ctx, &cloud.Login_Request{
		Token: token,
	})
	if err != nil {
		return nil, diagnostics.AddErrorMsg("Login failed: %s", err.Error())
	}
	if response.Diagnosis != nil && response.Diagnosis.Code != 0 {
		switch response.Diagnosis.Code {
		case common.Diagnosis_IllegalToken:
			return nil, diagnostics.AddErrorMsg("Login failed, The Selefra Cloud recognizes that the token you entered is not a valid Token")
		default:
			return nil, diagnostics.AddErrorMsg("Login response error, code = %d, message = %s", response.Diagnosis.Code, response.Diagnosis.Msg)
		}
	}
	x.token = token
	credentials := &CloudCredentials{
		TokenName:   response.TokenName,
		Token:       token,
		UserName:    response.UserName,
		OrgName:     response.OrgName,
		ServerHost:  response.ServerHost,
		LoginTime:   time.Now(),
		LastUseTime: time.Now(),
	}
	d := x.SaveCredentials(credentials)

	return credentials, diagnostics.AddDiagnostics(d)
}

// Logout Log out the current token
func (x *CloudClient) Logout() *schema.Diagnostics {

	diagnostics := schema.NewDiagnostics()

	if x.token == "" {
		return diagnostics.AddErrorMsg("You need login first!")
	}

	response, err := x.cloudClient.Logout(x.BuildMetaContext(), &cloud.RequestEmpty{})
	if err != nil {
		return diagnostics.AddErrorMsg("Logout failed: %s", err.Error())
	}
	if response.Diagnosis != nil && response.Diagnosis.Code != 0 {
		return diagnostics.AddErrorMsg("Logout failed, cloud response code error, code = %d, message = %s", response.Diagnosis.Code, response.Diagnosis.Msg)
	}

	// clear current client
	x.token = ""
	//x.LogClient = nil
	//x.LogStreamUploader = nil
	//x.IssueStreamUploader = nil

	// remove local save credentials
	d := x.SaveCredentials(&CloudCredentials{})
	if diagnostics.AddDiagnostics(d).HasError() {
		return diagnostics
	}

	return diagnostics
}

//type LoginRequest struct {
//	Token string `json:"token"`
//}
//
//type LoginData struct {
//	UserName  string `json:"user_name"`
//	TokenName string `json:"token_name"`
//	OrgName   string `json:"org_name"`
//}
//
//func (x *CloudClient) Login(ctx context.Context, token string) (*Response[LoginData], error) {
//	response, err := http_client.PostJson[*LoginRequest, *Response[LoginData]](ctx, x.buildAPIURL("/cli/login"), &LoginRequest{
//		Token: token,
//	})
//	if err != nil {
//		return nil, err
//	}
//	if err := response.Check(); err != nil {
//		return nil, err
//	}
//	x.token = token
//	return response, nil
//}
//
//// ------------------------------------------------- --------------------------------------------------------------------
//
//type logoutData struct {
//}
//
//type LogoutRequest struct {
//	Token string `json:"token"`
//}
//
//type LogoutResponse struct {
//}
//
//func (x *CloudClient) Logout(ctx context.Context) error {
//
//	var info = make(map[string]string)
//	info["token"] = token
//	res, err := CliHttpClient[logoutData]("POST", "/cli/logout", info)
//	if err != nil {
//		return err
//	}
//	if res.Code != 0 {
//		return fmt.Errorf(res.Msg)
//	}
//	return nil
//}

//func (x *CloudClient) Logout() error {
//
//	if x.token == "" {
//		return fmt.Errorf("not login status")
//	}
//
//	err := http_client.Logout(token)
//	if err != nil {
//		ui.Errorln("Logout error:" + err.Error())
//		return nil
//	}
//
//	err = utils.SaveCredentials("")
//	if err != nil {
//		ui.Errorln(err.Error())
//	}
//
//	return nil
//}
//
//// ------------------------------------------------- --------------------------------------------------------------------

//// ShouldLogin should login to selefra cloud
//// if login successfully, global token will be set, else return an error
//func (x *CloudClient) ShouldLogin(tokens ...string) error {
//	var err error
//	var token string
//	if len(tokens) > 0 {
//		token = tokens[0]
//	}
//
//	if token == "" {
//		token, err = utils.GetCredentialsToken()
//		if err != nil {
//			ui.Errorln(err.Error())
//			return err
//		}
//	}
//
//	res, err := http_client.Login(token)
//	if err != nil {
//		return ErrLoginFailed
//	}
//	displayLoginSuccess(res.Data.OrgName, res.Data.TokenName, token)
//
//	return nil
//}

//// MustLogin unless the user enters wrong token, login is guaranteed
//func MustLogin(token string) error {
//	var err error
//
//	if err := ShouldLogin(token); err == nil {
//		return nil
//	}
//
//	token, err = getInputToken()
//	if err != nil {
//		return errors.New("input token failed")
//	}
//	if err = ShouldLogin(token); err == nil {
//		return nil
//	}
//
//	return ErrLoginFailed
//}
