VERSION := $(if ${PULUMI_VERSION},${PULUMI_VERSION},$(shell ./scripts/pulumi-version.sh))
PYTHON_SDK_VERSION := $(shell echo "$(VERSION)" | sed 's/-/./g')

CONCURRENCY := 10
SHELL := sh

GO := go

.phony: .EXPORT_ALL_VARIABLES
.EXPORT_ALL_VARIABLES:

default: ensure build_go

install::
	${GO} install ./cmd/...

clean::
	rm -f ./bin/*

ensure_go::
	cd sdk && ${GO} mod download

.phony: lint
lint:: lint-copyright lint-golang lint-python
lint-golang:
	cd sdk && golangci-lint run
lint-python:
	flake8 ./sdk/python/pulumi_esc_sdk/esc_client.py --count --exit-zero --max-complexity=10 --max-line-length=127 --statistics
	flake8 ./sdk/python/test/ --count --exit-zero --max-complexity=10 --max-line-length=127 --statistics
lint-copyright:
	pulumictl copyright

.phony: format
format:
	find . -iname "*.go" -print0 | xargs -r0 gofmt -s -w

build_go:: ensure_go
	cd sdk && ${GO} build -ldflags "-X github.com/pulumi/esc/cmd/internal/version.Version=${VERSION}" ./...

build_debug:: ensure_go
	cd sdk && ${GO} build -gcflags="all=-N -l" -ldflags "-X github.com/pulumi/esc/cmd/internal/version.Version=${VERSION}" ./...

build_python::
	PYPI_VERSION=$(VERSION) ./scripts/build_python_sdk.sh

test_go:: build_go
	cd sdk && ${GO} test --timeout 30m -short -count 1 -parallel ${CONCURRENCY} ./...

test_go_cover:: build_go
	cd sdk && ${GO} test --timeout 30m -count 1 -coverpkg=github.com/pulumi/esc-sdk/... -race -coverprofile=coverage.out -parallel ${CONCURRENCY} ./...

test_typescript:: 
	cd sdk/typescript && npm i && npm run test

test_python:: 
	cd sdk/python && rm -rf ./bin/ && pytest

.PHONY: generate_go_client_sdk
generate_go_client_sdk:
	GO_POST_PROCESS_FILE="/usr/local/bin/gofmt -w" openapi-generator-cli generate -i ./sdk/swagger.yaml -p packageName=esc_sdk,withGoMod=false,isGoSubmodule=true,userAgent=esc-sdk/go/${VERSION} -t ./sdk/templates/go -g go -o ./sdk/go --git-repo-id esc --git-user-id pulumi

.PHONY: generate_ts_client_sdk
generate_ts_client_sdk:
	TS_POST_PROCESS_FILE="/usr/local/bin/prettier --write" openapi-generator-cli generate -i ./sdk/swagger.yaml -p npmName=@pulumi/esc-sdk,userAgent=esc-sdk/ts/${VERSION} -t ./sdk/templates/typescript --enable-post-process-file -g typescript-axios -o ./sdk/typescript/esc/raw  --git-repo-id esc --git-user-id pulumi

.PHONY: generate_python_client_sdk
generate_python_client_sdk:
	PYTHON_POST_PROCESS_FILE="/usr/local/bin/yapf -i" openapi-generator-cli generate -i ./sdk/swagger.yaml -p packageName=pulumi_esc_sdk,httpUserAgent=esc-sdk/python/${VERSION},packageVersion=${PYTHON_SDK_VERSION} -t ./sdk/templates/python -g python -o ./sdk/python --git-repo-id esc --git-user-id pulumi

.phony: generate_sdks
generate_sdks:: generate_go_client_sdk generate_ts_client_sdk generate_python_client_sdk