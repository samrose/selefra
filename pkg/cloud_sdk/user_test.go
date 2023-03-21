package cloud_sdk

import (
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCloudClient_IsLoggedIn(t *testing.T) {

	// unauth
	unAuthCloudClient := getUnAuthSDKClientForTest()
	assert.False(t, unAuthCloudClient.IsLoggedIn())

	// auth
	authCloudClient := getAuthedSDKClientForTest()
	assert.True(t, authCloudClient.IsLoggedIn())

}

func TestCloudClient_Login(t *testing.T) {
	client := getAuthedSDKClientForTest()
	assert.NotNil(t, client)
}

func TestCloudClient_Logout(t *testing.T) {
	client := getAuthedSDKClientForTest()
	assert.True(t, client.IsLoggedIn())
	d := client.Logout()
	if utils.IsNotEmpty(d) {
		t.Log(d.ToString())
	}
	assert.False(t, utils.HasError(d))
	assert.False(t, client.IsLoggedIn())
}
