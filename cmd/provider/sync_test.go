package provider

//import (
//	"context"
//	"github.com/selefra/selefra/config"
//	"github.com/selefra/selefra/global"
//	"github.com/stretchr/testify/require"
//	"testing"
//)
//
//func Test_effectiveDecls(t *testing.T) {
//	ctx := context.Background()
//	global.Init("", global.WithWorkspace("../../tests/workspace/offline"))
//	rootConfig, err := config.GetConfig()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	decls, _ := effectiveDecls(ctx, rootConfig.Selefra.ProviderDecls)
//
//	require.Equal(t, 1, len(decls))
//
//	require.Equal(t, "aws", decls[0].Name)
//	require.Equal(t, "v0.0.9", decls[0].Version)
//}
