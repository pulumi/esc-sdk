# The ESC SDKs now live in github.com/pulumi/pulumi. This repository only keeps
# the Go re-export shim at sdk/go so that github.com/pulumi/esc-sdk/sdk/go keeps
# working for existing consumers. See sdk/go/shim_gen.go.
GO := go

.PHONY: build
build::
	cd sdk && $(GO) build ./...

.PHONY: lint
lint::
	cd sdk && $(GO) vet ./...

.PHONY: test
test::
	cd sdk && $(GO) test ./...
