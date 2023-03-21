package cli_env

import (
	"os"
	"strings"
)

const SelefraCloudFlag = "SELEFRA_CLOUD_FLAG"

// IsCloudEnv Check whether the system is running in the cloud environment
func IsCloudEnv() bool {
	flag := strings.ToLower(os.Getenv(SelefraCloudFlag))
	return flag == "true" || flag == "enable"
}

//func GetTaskID() {
//
//}

// ------------------------------------------------- --------------------------------------------------------------------

const SelefraServerHost = "SELEFRA_CLOUD_HOST"

const DefaultCloudHost = "main-grpc.selefra.io:1234"

// GetServerHost Gets the address of the server
func GetServerHost() string {

	// read from env
	if os.Getenv(SelefraServerHost) != "" {
		return os.Getenv(SelefraServerHost)
	}

	return DefaultCloudHost
}

// ------------------------------------------------- --------------------------------------------------------------------

const SelefraCloudToken = "SELEFRA_CLOUD_TOKEN"

func GetCloudToken() string {
	return os.Getenv(SelefraCloudToken)
}

// ------------------------------------------------- --------------------------------------------------------------------

const SelefraCloudHttpHost = "SELEFRA_CLOUD_HTTP_HOST"

const DefaultSelefraCloudHttpHost = "https://www.selefra.io"

func GetSelefraCloudHttpHost() string {

	if os.Getenv(SelefraCloudHttpHost) != "" {
		return os.Getenv(SelefraCloudHttpHost)
	}

	return DefaultSelefraCloudHttpHost
}

// ------------------------------------------------ ---------------------------------------------------------------------

const SelefraTelemetryEnable = "SELEFRA_TELEMETRY_ENABLE"

func GetSelefraTelemetryEnable() string {
	return strings.ToLower(os.Getenv(SelefraTelemetryEnable))
}

// ------------------------------------------------ ---------------------------------------------------------------------

// SelefraTelemetryToken Compile-time injection
var SelefraTelemetryToken = ""

const SelefraTelemetryTokenEnvName = "SELEFRA_TELEMETRY_TOKEN"

func GetSelefraTelemetryToken() string {
	if SelefraTelemetryToken != "" {
		return SelefraTelemetryToken
	}
	return os.Getenv(SelefraTelemetryTokenEnvName)
}

// ------------------------------------------------ ---------------------------------------------------------------------
