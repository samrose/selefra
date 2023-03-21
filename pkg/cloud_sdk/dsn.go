package cloud_sdk

import (
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/grpc/pb/cloud"
)

// ------------------------------------------------- --------------------------------------------------------------------

// FetchOrgDSN Getting a user-configured database connection from the selefra cloud may not be configured
func (x *CloudClient) FetchOrgDSN() (string, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	if !x.IsLoggedIn() {
		return "", diagnostics.AddErrorMsg("You need login first!")
	}

	response, err := x.cloudClient.FetchOrgDsn(x.BuildMetaContext(), &cloud.RequestEmpty{})
	if err != nil {
		return "", schema.NewDiagnostics().AddErrorMsg("Request DSN from selefra cloud failed: %s", err.Error())
	}
	if response.Diagnosis != nil && response.Diagnosis.Code != 0 {
		return "", schema.NewDiagnostics().AddErrorMsg("Request DSN from selefra cloud response error, code = %d, msg = %s", response.Diagnosis.Code, response.Diagnosis.Msg)
	}
	return response.Dsn, nil
}
