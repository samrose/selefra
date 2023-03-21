package cloud_sdk

import (
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/grpc/pb/cloud"
	"os"
)

func (x *CloudClient) CreateTask(projectName string) (*cloud.CreateTask_Response, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	if !x.IsLoggedIn() {
		return nil, diagnostics.AddErrorMsg("You need login first!")
	}

	response, err := x.cloudClient.CreateTask(x.BuildMetaContext(), &cloud.CreateTask_Request{
		ProjectName: projectName,
		TaskId:      os.Getenv("SELEFRA_TASK_ID"),
		TaskSource:  os.Getenv("SELEFRA_TASK_SOURCE"),
		Name:        os.Getenv("SELEFRA_TASK_NAME"),
	})
	if err != nil {
		return nil, diagnostics.AddErrorMsg("create cloud task failed: %s", err.Error())
	}

	if response.Diagnosis != nil && response.Diagnosis.Code != 0 {
		return nil, diagnostics.AddErrorMsg("create cloud task response error, code = %d, message = %s", response.Diagnosis.Code, response.Diagnosis.Msg)
	}

	//d := x.InitTaskClientContext(response.TaskId, )
	//if diagnostics.AddDiagnostics(d).HasError() {
	//	return diagnostics
	//}

	x.taskId = response.TaskId

	return response, diagnostics
}

//// ------------------------------------------------- --------------------------------------------------------------------
//
//type TaskData struct {
//	TaskUUID string `json:"task_uuid"`
//}
//
//type CreateTaskRequest struct {
//	Token       string `json:"token"`
//	ProjectName string `json:"project_name"`
//	TaskID      string `json:"task_id"`
//	TaskSource  string `json:"task_source"`
//}
//
//// TryCreateTask create a task in selefra cloud when use is login, else do nothing
//func (x *CloudClient) TryCreateTask(ctx context.Context, projectName string) (*Response[TaskData], error) {
//
//	if !x.IsLoggedIn() {
//		return nil, ErrYouAreNotLogin
//	}
//
//	requestBody := &CreateTaskRequest{
//		Token:       x.token,
//		ProjectName: projectName,
//		TaskID:      os.Getenv("SELEFRA_TASK_ID"),
//		TaskSource:  os.Getenv("SELEFRA_TASK_SOURCE"),
//	}
//	return http_client.PostJson[*CreateTaskRequest, *Response[TaskData]](ctx, x.buildAPIURL("/cli/create_task"), requestBody)
//}
//
//// ------------------------------------------------- --------------------------------------------------------------------
