.DEFAULT_GOAL: selefra

GOPATH := $(shell go env GOPATH)
ifeq ($(origin GOBIN),undefined)
    GOBIN := $(GOPATH)/bin
endif

# if env variable SELEFRA_VERSION is not set, use git id as selefra version
ifeq ($(origin SELEFRA_VERSION),undefined)
	SELEFRA_VERSION := $(shell git rev-parse HEAD)
endif

build:
	@echo "GOBIN=${GOBIN}"
	@echo "SELEFRA_VERSION=$(SELEFRA_VERSION)"
	@#go build -o $(GOBIN)  -ldflags "-X 'github.com/selefra/selefra/cmd/version.Version=$(SELEFRA_VERSION)'" main.go
	@go build -o selefra -ldflags "-X 'github.com/selefra/selefra/cmd/version.Version=$(SELEFRA_VERSION)' -X 'github.com/selefra/selefra/pkg/cli_env.SelefraTelemetryToken=${SELEFRA_TELEMETRY_TOKEN}'" main.go

PROTO_FILES ?= issue log
.PHONY: protoc
protoc: $(addprefix protoc-, $(PROTO_FILES))

.PHONY: protoc-%
protoc-%:
	@protoc  -I./pkg/grpcClient/proto/third-party/ -I. \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/grpcClient/proto/$*/$*.proto
