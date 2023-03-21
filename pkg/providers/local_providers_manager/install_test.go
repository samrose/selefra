package local_providers_manager

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/message"
	"testing"
)

//import (
//	"context"
//	"github.com/selefra/selefra/global"
//	"testing"
//)
//
//func TestInstall(t *testing.T) {
//	*global.WORKSPACE = "../../tests/workspace/offline"
//	ctx := context.Background()
//	err := install(ctx, []string{"aws@latest"})
//	if err != nil {
//		t.Error(err)
//	}
//}

func TestLocalProvidersManager_InstallProvider(t *testing.T) {
	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		t.Log(message.ToString())
	})
	getTestLocalProviderManager().InstallProvider(context.Background(), &InstallProvidersOptions{
		RequiredProvider: NewLocalProvider("mock", "v0.0.1"),
		MessageChannel:   messageChannel,
	})
	messageChannel.ReceiverWait()
}
