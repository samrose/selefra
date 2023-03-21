package cloud_sdk

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	selefraGrpc "github.com/selefra/selefra/pkg/grpc"
	"github.com/selefra/selefra/pkg/grpc/pb/cloud"
	"github.com/selefra/selefra/pkg/grpc/pb/issue"
	"github.com/selefra/selefra/pkg/grpc/pb/log"
	"github.com/selefra/selefra/pkg/message"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"time"
)

const (
	IssueUploaderName = "issue-uploader"
	LogUploaderName   = "log-uploader"
)

type CloudClient struct {
	serverUrl string

	cloudNoAuthClient cloud.CloudNoAuthClient
	cloudClient       cloud.CloudClient

	taskId string
	token  string

	// This parameter is used to upload Issues
	//IssueStreamUploader *selefraGrpc.StreamUploader[issue.Issue_UploadIssueStreamClient, int, *issue.UploadIssueStream_Request, *issue.UploadIssueStream_Response]

	// This parameter is used to upload logs
	//LogClient         log.LogClient
	//LogStreamUploader *selefraGrpc.StreamUploader[log.Log_UploadLogStreamClient, int, *log.UploadLogStream_Request, *log.UploadLogStream_Response]

	//MessageChannel *message.Channel[*schema.Diagnostics]
}

func NewCloudClient(serverUrl string) (*CloudClient, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	x := &CloudClient{
		serverUrl: serverUrl,
		//cloudNoAuthClient: cloudNoAuthClient,
		//cloudClient:       cloud.NewCloudClient(conn),
		//IssueStreamUploader: nil,
		//LogStreamUploader:   nil,
	}

	conn, err := x.DialCloudHost()
	if err != nil {
		return nil, diagnostics.AddErrorMsg("connect to cloud server %s failed: %s", serverUrl, err.Error())
	}
	cloudNoAuthClient := cloud.NewCloudNoAuthClient(conn)
	x.cloudNoAuthClient = cloudNoAuthClient

	conn, err = x.DialCloudHost()
	if err != nil {
		return nil, diagnostics.AddErrorMsg("connect to cloud server %s failed: %s", serverUrl, err.Error())
	}
	x.cloudClient = cloud.NewCloudClient(conn)

	return x, nil
}

//// InitTaskClientContext Initialize the task client context, for after report data to selefra cloud
//func (x *CloudClient) InitTaskClientContext(taskId string, messageChannel *message.Channel[*schema.Diagnostics]) *schema.Diagnostics {
//
//	diagnostics := schema.NewDiagnostics()
//
//	x.taskId = taskId
//	x.MessageChannel = messageChannel
//
//	issueStreamUploader, d := x.NewIssueStreamUploader()
//	if diagnostics.AddDiagnostics(d).HasError() {
//		return diagnostics
//	}
//	x.IssueStreamUploader = issueStreamUploader
//
//	logClient, logStreamUploader, d := x.NewLogStreamUploader()
//	if diagnostics.AddDiagnostics(d).HasError() {
//		return diagnostics
//	}
//	x.LogClient = logClient
//	x.LogStreamUploader = logStreamUploader
//
//	return diagnostics
//}

// NewIssueStreamUploader Create a component that uploads Issues
func (x *CloudClient) NewIssueStreamUploader(messageChannel *message.Channel[*schema.Diagnostics]) (*selefraGrpc.StreamUploader[issue.Issue_UploadIssueStreamClient, int, *issue.UploadIssueStream_Request, *issue.UploadIssueStream_Response], *schema.Diagnostics) {

	// new connection
	conn, err := x.DialCloudHost()
	if err != nil {
		return nil, schema.NewDiagnostics().AddErrorMsg("connect to cloud server %s failed: %s", x.serverUrl, err.Error())
	}

	// create upload issue stream client
	stream, err := issue.NewIssueClient(conn).UploadIssueStream(x.BuildMetaContext())
	if err != nil {
		return nil, schema.NewDiagnostics().AddErrorMsg("")
	}
	uploaderOptions := &selefraGrpc.StreamUploaderOptions[issue.Issue_UploadIssueStreamClient, int, *issue.UploadIssueStream_Request, *issue.UploadIssueStream_Response]{
		Name:                      IssueUploaderName,
		Client:                    stream,
		WaitSendTaskQueueBuffSize: 1000,
		MessageChannel:            messageChannel,
	}
	uploader := selefraGrpc.NewStreamUploader[issue.Issue_UploadIssueStreamClient, int, *issue.UploadIssueStream_Request, *issue.UploadIssueStream_Response](uploaderOptions)
	return uploader, nil
}

// NewLogStreamUploader Create a component that uploads logs
func (x *CloudClient) NewLogStreamUploader(messageChannel *message.Channel[*schema.Diagnostics]) (log.LogClient, *selefraGrpc.StreamUploader[log.Log_UploadLogStreamClient, int, *log.UploadLogStream_Request, *log.UploadLogStream_Response], *schema.Diagnostics) {

	// new connection
	conn, err := x.DialCloudHost()
	if err != nil {
		return nil, nil, schema.NewDiagnostics().AddErrorMsg("connect to cloud server %s failed: %s", x.serverUrl, err.Error())
	}

	// create upload
	diagnostics := schema.NewDiagnostics()
	client := log.NewLogClient(conn)
	stream, err := client.UploadLogStream(x.BuildMetaContext())
	if err != nil {
		return nil, nil, diagnostics.AddErrorMsg("create cloud log stream error: %s", err.Error())
	}
	uploaderOptions := &selefraGrpc.StreamUploaderOptions[log.Log_UploadLogStreamClient, int, *log.UploadLogStream_Request, *log.UploadLogStream_Response]{
		Name:                      LogUploaderName,
		Client:                    stream,
		WaitSendTaskQueueBuffSize: 1000,
		MessageChannel:            messageChannel,
	}
	uploader := selefraGrpc.NewStreamUploader[log.Log_UploadLogStreamClient, int, *log.UploadLogStream_Request, *log.UploadLogStream_Response](uploaderOptions)
	return client, uploader, nil
}

func (x *CloudClient) DialCloudHost() (*grpc.ClientConn, error) {
	return grpc.Dial(x.serverUrl,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                3 * time.Second,
			Timeout:             3 * time.Minute,
			PermitWithoutStream: true}))
	//return grpc.Dial(x.serverUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// ------------------------------------------------- --------------------------------------------------------------------

func (x *CloudClient) BuildMetaContext() context.Context {
	return metadata.AppendToOutgoingContext(context.Background(), "taskUUID", x.taskId, "token", x.token)
}

// ------------------------------------------------- --------------------------------------------------------------------
//
//type OutputReq struct {
//	Name     string              `json:"name"`
//	Query    string              `json:"query"`
//	Labels   map[string][]string `json:"labels"`
//	Metadata Metadata            `json:"metadata"`
//}
//
//type Metadata struct {
//	Id           string   `json:"id"`
//	Severity     string   `json:"severity"`
//	Provider     string   `json:"provider"`
//	Tags         []string `json:"tags"`
//	SrcTableName []string `json:"src_table_name"`
//	Remediation  string   `yaml:"remediation" json:"remediation"`
//	Author       string   `json:"author"`
//	Title        string   `json:"title"`
//	Description  string   `json:"description"`
//	Output       string   `json:"output"`
//}
//
//type OutputRes struct {
//}
//
//type UploadWorkplaceRes struct {
//}
//
//// ------------------------------------------------- --------------------------------------------------------------------
//
//type Response[T any] struct {
//	Code int    `json:"code"`
//	Data T      `json:"data"`
//	Msg  string `json:"msg"`
//}
//
//func (x *Response[T]) IsResponseCodeOk() bool {
//	return x.Code == 200
//}
//
//func (x *Response[T]) Check() error {
//	if !x.IsResponseCodeOk() {
//		// TODO error
//		return errors.New("")
//	}
//	return nil
//}
//
//// ------------------------------------------------- --------------------------------------------------------------------

//func (x *CloudClient) CliHttpClient[T any](method, url string, info interface{}) (*Response[T], error) {
//	var client http.Client
//	httpLogger.Info("request info: %s , %s", url, info)
//	bytesData, err := json.Marshal(info)
//	if err != nil {
//		return nil, err
//	}
//	req, err := http.NewRequest(method, "https://"+global.SERVER+url, bytes.NewReader(bytesData))
//	if err != nil {
//		return nil, err
//	}
//	req.Header.Set("Content-Type", "application/json")
//	resp, err := client.Do(req)
//	if err != nil {
//		fmt.Println(err.Error())
//		return nil, err
//	}
//	defer resp.Body.Close()
//	if resp.StatusCode == http.StatusNotFound {
//		return nil, errors.New("404 not found")
//	}
//	respBytes, err := io.ReadAll(resp.Body)
//	httpLogger.Info("resp info: %s , %s", url, string(respBytes))
//	if err != nil {
//		return nil, err
//	}
//	var res Response[T]
//	err = json.Unmarshal(respBytes, &res)
//	if err != nil {
//		return nil, err
//	}
//	return &res, err
//}

// ------------------------------------------------- --------------------------------------------------------------------

//func (x *CloudClient) buildAPIURL(apiRequestPath string) string {
//	return path.Join(x.serverUrl, apiRequestPath)
//}
//
//// ------------------------------------------------- --------------------------------------------------------------------
//
//func f() {
//	ctx := context.Background()
//	var conn *grpc.ClientConn
//	var err error
//	grpc.DialContext()
//	conn, err = grpc.Dial(getDial(), grpc.WithTransportCredentials(insecure.NewCredentials()))
//	if err != nil {
//	}
//
//	innerClient := client{
//		ctx:       ctx,
//		conn:      conn,
//		statusMap: make(map[string]string),
//	}
//
//	var openedLogStreamClient logPb.Log_UploadLogStreamClient
//	logClient := logPb.NewLogClient(conn)
//	innerClient.logClient = logClient
//
//	openedLogStreamClient, err = logClient.UploadLogStream(ctx)
//	if err != nil {
//		return
//	}
//	innerClient.openedLogStreamClient = openedLogStreamClient
//
//	var openedIssueStreamClient issue.Issue_UploadIssueStreamClient
//	issueStreamClient := issue.NewIssueClient(conn)
//	openedIssueStreamClient, err = issueStreamClient.UploadIssueStream(ctx)
//	if err != nil {
//		return
//	}
//	innerClient.openedIssueStreamClient = openedIssueStreamClient
//
//	utils.MultiRegisterClose(map[string]func(){
//		"grpc conn": func() {
//			_ = conn.Close()
//		},
//		"log stream": func() {
//			_ = openedLogStreamClient.CloseSend()
//		},
//		"issue stream": func() {
//			_ = openedIssueStreamClient.CloseSend()
//		},
//	})
//
//	c = &innerClient
//})
//}

// ------------------------------------------------- --------------------------------------------------------------------
