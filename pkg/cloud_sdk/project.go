package cloud_sdk

import (
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/grpc/pb/cloud"
	"github.com/selefra/selefra/pkg/grpc/pb/common"
)

// CreateProject Returns the name of the project if the given project name already exists,
// otherwise creates the project and returns information about the project
func (x *CloudClient) CreateProject(projectName string) (*cloud.CreateProject_Response, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	if !x.IsLoggedIn() {
		return nil, diagnostics.AddErrorMsg("You need login first!")
	}

	response, err := x.cloudClient.CreateProject(x.BuildMetaContext(), &cloud.CreateProject_Request{
		Name: projectName,
	})
	if err != nil {
		return nil, diagnostics.AddErrorMsg("create cloud project error: %s", err.Error())
	}
	if response.Diagnosis != nil && response.Diagnosis.Code != 0 {
		switch response.Diagnosis.Code {
		case common.Diagnosis_NoAuthority:
			errorMsg := `Free users can only create a project, you can pay in this upgrade at https://app.selefra.io/Settings/planBilling
Alternatively, you can logout the currently logged user using the command selefra logout, which will not be synchronized to the cloud.`
			return nil, diagnostics.AddErrorMsg(errorMsg)
		default:
			return nil, diagnostics.AddErrorMsg("create cloud project response error, code = %d, message = %s", response.Diagnosis.Code, response.Diagnosis.Msg)
		}
	}

	return response, nil
}

//// ------------------------------------------------- --------------------------------------------------------------------
//
//type CreateProjectRequest struct {
//	Token       string `json:"token"`
//	ProjectName string `json:"name"`
//}
//
//type CreateProjectData struct {
//	Name    string `json:"name"`
//	OrgName string `json:"org_name"`
//}
//
//// CreateProject create a project in selefra cloud when use is login, else do nothing
//func (x *CloudClient) CreateProject(ctx context.Context, projectName string) (orgName string, err error) {
//
//	if !x.IsLoggedIn() {
//		return
//	}
//
//	response, err := http_client.PostJson[*CreateProjectRequest, *Response[CreateProjectData]](ctx, x.buildAPIURL("/cli/create_project"), &CreateProjectRequest{})
//	if err != nil {
//		return "", err
//	}
//	if err := response.Check(); err != nil {
//		return "", err
//	}
//	return response.Data.OrgName, nil
//}
//
//// ------------------------------------------------- --------------------------------------------------------------------
//
//type Stage string
//
//const (
//	Creating = "creating"
//	Testing  = "testing"
//	Failed   = "failed"
//	Success  = "success"
//)
//
//type SetupProjectStageDataRequest struct {
//	Token       string `json:"token"`
//	ProjectName string `json:"project_name"`
//	Stag        string `json:"stag"`
//}
//
//type SetupProjectStageData struct{}
//
//// UploadSetupStage sync project stage to selefra cloud when use is login, else do nothing
//func (x *CloudClient) UploadSetupStage(ctx context.Context, projectName string, stage Stage) error {
//
//	if !x.IsLoggedIn() {
//		return ErrYouAreNotLogin
//	}
//
//	response, err := http_client.PostJson[*SetupProjectStageDataRequest, *Response[SetupProjectStageData]](ctx, x.buildAPIURL("/cli/update_setup_stag"), &SetupProjectStageDataRequest{
//		Token:       x.token,
//		ProjectName: projectName,
//		Stag:        string(stage),
//	})
//	if err != nil {
//		return err
//	}
//	if err := response.Check(); err != nil {
//		return err
//	}
//	return nil
//}
//
//// ------------------------------------------------- --------------------------------------------------------------------
