package tools

import (
	"context"
	"github.com/selefra/selefra/config"
	"github.com/selefra/selefra/global"
	"github.com/selefra/selefra/pkg/registry"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"testing"
)

func getProviderAndConfig() (registry.ProviderBinary, *config.RootConfig, error) {
	global.Init("", global.WithWorkspace("../../tests/workspace/offline"))
	ctx := context.Background()
	cof, err := config.GetConfig()
	if err != nil {
		return registry.ProviderBinary{}, nil, err
	}
	pr := registry.Provider{
		Name:    "aws",
		Version: "latest",
		Source:  "",
	}
	namespace, _, err := utils.Home()
	if err != nil {
		return registry.ProviderBinary{}, nil, err
	}
	provider := registry.NewProviderRegistry(namespace)
	p, err := provider.Download(ctx, pr, true)
	return p, cof, err
}

func TestGetProviders(t *testing.T) {
	global.Init("", global.WithWorkspace("../../tests/workspace/offline"))
	s, err := config.GetConfig()
	if err != nil {
		t.Error(err)
	}
	provider, err := ProviderConfigStrs(s, "aws")
	require.NoError(t, err)
	require.Equal(t, 1, len(provider))
	require.Equal(t, "aws_01", provider[0])
}

func TestSetProviders(t *testing.T) {
	p, cof, err := getProviderAndConfig()

	var node yaml.Node
	yamls := `providers:
    - name: aws_01
      cache: 1d
      provider: aws
      resources:`
	if err := yaml.Unmarshal([]byte(yamls), &node); err != nil {
		t.Fatal(err)
	}

	s := `
      ##  Optional, Repeated. Add an accounts block for every account you want to assume-role into and fetch data from.
      #accounts:
      #    #     Optional. User identification
      #  - account_name: <UNIQUE ACCOUNT IDENTIFIER>
      #    #    Optional. Named profile in config or credential file from where Selefra should grab credentials
      #    shared_config_profile: < PROFILE_NAME >
      #    #    Optional. Location of shared configuration files
      #    shared_config_files:
      #      - <FILE_PATH>
      #    #   Optional. Location of shared credentials files
      #    shared_credentials_files:
      #      - <FILE_PATH>
      #    #    Optional. Role ARN we want to assume when accessing this account
      #    role_arn: < YOUR_ROLE_ARN >
      #    #    Optional. Named role session to grab specific operation under the assumed role
      #    role_session_name: <SESSION_NAME>
      #    #    Optional. Any outside of the org account id that has additional control
      #    external_id: <ID>
      #    #    Optional. Designated region of servers
      #    default_region: <REGION_CODE>
      #    #    Optional. by default assumes all regions
      #    regions:
      #      - us-east-1
      #      - us-west-2
      ##    The maximum number of times that a request will be retried for failures. Defaults to 10 retry attempts.
      #max_attempts: 10
      ##    The maximum back off delay between attempts. The backoff delays exponentially with a jitter based on the number of attempts. Defaults to 30 seconds.
      #max_backoff: 30
`
	err = SetProviderTmpl(s, p, cof)
	b, err := yaml.Marshal(cof.Providers)
	t.Log(string(b))
	if err != nil {
		t.Error(err)
	}
}

func TestSetSelefraProvider(t *testing.T) {
	p, cof, err := getProviderAndConfig()

	if err != nil {
		t.Error(err)
	}
	err = AppendProviderDecl(p, cof, "latest")
	if err != nil {
		t.Error(err)
	}
}
