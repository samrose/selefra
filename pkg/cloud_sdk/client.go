package cloud_sdk

//import (
//	"context"
//	"fmt"
//	"github.com/selefra/selefra/global"
//	"github.com/selefra/selefra/pkg/grpc_client/proto/issue"
//	"github.com/selefra/selefra/pkg/utils"
//	"google.golang.org/grpc"
//	"google.golang.org/grpc/credentials/insecure"
//	"google.golang.org/protobuf/types/known/timestamppb"
//	"strings"
//	"sync"
//)
//
//type client struct {
//	ctx context.Context
//
//	// conn is a grpc connection
//	conn *grpc.ClientConn
//
//	// openedIssueStreamClient is an opened issue upload stream client
//	openedIssueStreamClient issue.Issue_UploadIssueStreamClient
//
//	// openedLogStreamClient is an opened log upload stream client
//	openedLogStreamClient logPb.Log_UploadLogStreamClient
//
//	// logClient is a client for upload Log
//	logClient logPb.LogClient
//
//	taskID    string
//	token     string
//	statusMap map[string]string
//}
//
//var o = sync.Once{}
//var c *client
//
//func shouldClient() {
//	o.Do(func() {
//		// when user not login, do nothing
//		if global.Token() == "" {
//			return
//		}
//
//		ctx := context.Background()
//		var conn *grpc.ClientConn
//		var err error
//		conn, err = grpc.Dial(getDial(), grpc.WithTransportCredentials(insecure.NewCredentials()))
//		if err != nil {
//		}
//
//		innerClient := client{
//			ctx:       ctx,
//			conn:      conn,
//			statusMap: make(map[string]string),
//		}
//
//		var openedLogStreamClient logPb.Log_UploadLogStreamClient
//		logClient := logPb.NewLogClient(conn)
//		innerClient.logClient = logClient
//
//		openedLogStreamClient, err = logClient.UploadLogStream(ctx)
//		if err != nil {
//			return
//		}
//		innerClient.openedLogStreamClient = openedLogStreamClient
//
//		var openedIssueStreamClient issue.Issue_UploadIssueStreamClient
//		issueStreamClient := issue.NewIssueClient(conn)
//		openedIssueStreamClient, err = issueStreamClient.UploadIssueStream(ctx)
//		if err != nil {
//			return
//		}
//		innerClient.openedIssueStreamClient = openedIssueStreamClient
//
//		utils.MultiRegisterClose(map[string]func(){
//			"grpc conn": func() {
//				_ = conn.Close()
//			},
//			"log stream": func() {
//				_ = openedLogStreamClient.CloseSend()
//			},
//			"issue stream": func() {
//				_ = openedIssueStreamClient.CloseSend()
//			},
//		})
//
//		c = &innerClient
//	})
//}
//
//func IssueStreamSend(req *issue.Req) error {
//	if c == nil {
//		shouldClient()
//		if c == nil {
//			return nil
//		}
//	}
//
//	return c.openedIssueStreamClient.Send(req)
//}
//
//func IssueStreamClose() error {
//	if c == nil {
//		shouldClient()
//		if c == nil {
//			return nil
//		}
//	}
//
//	return c.openedIssueStreamClient.CloseSend()
//}
//
//func LogStreamSend(req *logPb.ConnectMsg) error {
//	if c == nil {
//		shouldClient()
//		if c == nil {
//			return nil
//		}
//	}
//
//	return c.openedLogStreamClient.Send(req)
//}
//
//func LogStreamClose() error {
//	if c == nil {
//		shouldClient()
//		if c == nil {
//			return nil
//		}
//	}
//
//	return c.openedLogStreamClient.CloseSend()
//}
//
//func SetStatus(status string) {
//	if c == nil {
//		shouldClient()
//		if c == nil {
//			return
//		}
//	}
//
//	c.statusMap[global.Stage()] = status
//}
//
//func GetStatus() string {
//	if c == nil {
//		shouldClient()
//		if c == nil {
//			return "success"
//		}
//	}
//
//	return c.statusMap[global.Stage()]
//}
//
//func SetTaskID(taskId string) {
//	if c == nil {
//		shouldClient()
//		if c == nil {
//			return
//		}
//	}
//
//	c.taskID = taskId
//}
//
//func TaskID() string {
//	if c == nil {
//		shouldClient()
//		if c == nil {
//			return ""
//		}
//	}
//
//	return c.taskID
//}
//
//func Token() string {
//	if c == nil {
//		shouldClient()
//		if c == nil {
//			return ""
//		}
//	}
//
//	return c.token
//}
//
//func UploadLogStatus() (*logPb.Res, error) {
//	if c == nil {
//		shouldClient()
//		if c == nil {
//			return nil, nil
//		}
//	}
//
//	statusInfo := &logPb.StatusInfo{
//		BaseInfo: &logPb.BaseConnectionInfo{
//			Token:  c.token,
//			TaskId: c.taskID,
//		},
//		Stag:   global.Stage(),
//		Status: c.statusMap[global.Stage()],
//		Time:   timestamppb.Now(),
//	}
//	res, err := c.logClient.UploadLogStatus(c.ctx, statusInfo)
//	if err != nil {
//		return nil, fmt.Errorf("Fail to upload log status:%s", err.Error())
//	}
//
//	return res, nil
//}
//
//func getDial() string {
//	var dialMap = make(map[string]string)
//	dialMap["dev-api.selefra.io"] = "dev-tcp.selefra.io:1234"
//	dialMap["main-api.selefra.io"] = "main-tcp.selefra.io:1234"
//	dialMap["pre-api.selefra.io"] = "pre-tcp.selefra.io:1234"
//	if dialMap[global.SERVER] != "" {
//		return dialMap[global.SERVER]
//	}
//	arr := strings.Split(global.SERVER, ":")
//	return arr[0] + ":1234"
//}
