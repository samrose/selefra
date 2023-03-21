package cloud_sdk

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/grpc/pb/cloud"
	"github.com/selefra/selefra/pkg/modules/module_loader"
	"os"
	"path/filepath"
)

func (x *CloudClient) UploadWorkspace(ctx context.Context, projectName, workspace string) *schema.Diagnostics {
	diagnostics := schema.NewDiagnostics()
	fileSlice, err := workspaceYamlFileSlice(workspace)
	if err != nil {
		return diagnostics.AddErrorMsg("make workspace file map error: %s", err.Error())
	}
	response, err := x.cloudClient.SyncWorkplace(x.BuildMetaContext(), &cloud.SyncWorkplace_Request{
		ProjectName:      projectName,
		ProjectWorkplace: fileSlice,
	})
	if err != nil {
		return schema.NewDiagnostics().AddErrorMsg("upload workspace file error: %s", err.Error())
	}
	if response.Diagnosis != nil && response.Diagnosis.Code != 0 {
		return schema.NewDiagnostics().AddErrorMsg("upload workspace file response error, code = %d, message = %s", response.Diagnosis.Code, response.Diagnosis.Msg)
	}
	return nil
}

func workspaceYamlFileSlice(dirname string) ([]*cloud.SyncWorkplace_ProjectWorkplace, error) {

	fileSlice := make([]*cloud.SyncWorkplace_ProjectWorkplace, 0)
	var fn func(dirname string) error
	fn = func(dirname string) error {
		files, e := os.ReadDir(dirname)
		if e != nil {
			return e
		}
		for _, file := range files {
			if file.IsDir() {
				if err := fn(filepath.Join(dirname, file.Name())); err != nil {
					return err
				}
			} else {
				if module_loader.IsYamlFile(file) {
					b, e := os.ReadFile(filepath.Join(dirname, file.Name()))
					if e != nil {
						return e
					}
					fileSlice = append(fileSlice, &cloud.SyncWorkplace_ProjectWorkplace{
						Path:        filepath.Join(dirname, file.Name()),
						YamlContent: string(b),
					})
				}
			}
		}
		return nil
	}

	if err := fn(dirname); err != nil {
		return nil, err
	}
	return fileSlice, nil
}

//
//type ModuleLocalDirectory struct {
//}
//
//type WorkPlaceReq struct {
//	Data        []Data `json:"data"`
//	ProjectName string `json:"project_name"`
//	Token       string `json:"token"`
//}
//
//type Data struct {
//	Path        string `json:"path"`
//	YAMLContent string `json:"yaml_content"`
//}
//
//func (x *CloudClient) UploadWorkspace(project string) error {
//	fileMap, err := config.FileMap(global.WorkSpace())
//	if err != nil {
//		return err
//	}
//	err = http_client.TryUploadWorkspace(project, fileMap)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//// TODO Do not upload sensitive information
//// TryUploadWorkspace upload downloadWorkspace to selefra cloud when use is login, else do nothing
//func (x *CloudClient) TryUploadWorkspace(project string, fileMap map[string]string) error {
//	if global.Token() == "" || project == "" {
//		return nil
//	}
//
//	var workplace WorkPlaceReq
//
//	workplace.Token = global.Token()
//	workplace.ProjectName = project
//	workplace.Data = make([]Data, 0)
//	for k, v := range fileMap {
//		workplace.Data = append(workplace.Data, Data{
//			Path:        k,
//			YAMLContent: v,
//		})
//	}
//	res, err := CliHttpClient[UploadWorkplaceRes]("POST", "/cli/upload_workplace", workplace)
//	if err != nil {
//		return err
//	}
//	if res.Code != 0 {
//		return errors.New(res.Msg)
//	}
//	return nil
//}
