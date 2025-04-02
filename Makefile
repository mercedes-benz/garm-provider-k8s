# SPDX-License-Identifier: MIT

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

GARM_GITHUB_NAME ?= "github-pat"
GARM_GITHUB_BASE_URL ?= https://github.com
GARM_GITHUB_API_BASE_URL ?= https://api.github.com
GARM_GITHUB_UPLOAD_BASE_URL ?= https://uploads.github.com

# Set binary output folder
BIN_DIR := ./bin

BINARY_NAME := garm-provider-k8s

GO_BUILD_FLAGS :=

PKG_PATH := ./cmd/garm-provider-k8s/main.go

$(eval ARCH := $(shell go env GOARCH))

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


## Tool Binaries
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
GORELEASER ?= $(LOCALBIN)/goreleaser
MDTOC ?= $(LOCALBIN)/mdtoc
NANCY ?= $(LOCALBIN)/nancy
GOVULNCHECK ?= $(LOCALBIN)/govulncheck
KIND ?= $(LOCALBIN)/kind

## Tool Versions
GOLANGCI_LINT_VERSION ?= v1.59.1
GORELEASER_VERSION ?= v1.21.0
MDTOC_VERSION ?= v1.1.0
NANCY_VERSION ?= v1.0.46
KIND_VERSION ?= v0.22.0

##@ Build
.PHONY: all build copy run clean
all: clean build copy docker-build apply

build:
	@go mod tidy
	@if [ "$(OS)" = "mac" ]; then \
		echo "Building $(BINARY_NAME) for mac..." && \
		GOOS="darwin" CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(PKG_PATH); \
	else \
	  	echo "Building $(BINARY_NAME) for linux..." && \
		GOOS="linux" CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(PKG_PATH); \
	fi

test:
	@echo "Testing App..."
	@go mod tidy
	@go test -v ./...

copy:
	@echo "copying binaries..."
	@mkdir -p ./hack/bine
	@cp "$(BIN_DIR)/$(BINARY_NAME)" "./hack/bin/$(BINARY_NAME)"

clean:
	@echo "Cleaning up $(BIN_DIR)..."
	@rm -rf $(BIN_DIR)

DEBUG = 0

##@ Local Development
.PHONY: docker-build
docker-build: ## Build a garm image with the k8s provider
	docker build -t localhost:5000/garm-with-k8s:latest ./hack
	docker push localhost:5000/garm-with-k8s:latest

.PHONY: docker-build-summerwind-runner
docker-build-summerwind-runner: ## Build the used runner image
	$(eval RUNNER_IMAGE ?= $(shell echo "localhost:5000/runner:linux-ubuntu-22.04-$(shell uname -m)"))
	docker build -t $(RUNNER_IMAGE) ./runner/summerwind
	docker push $(RUNNER_IMAGE)

.PHONY: docker-build-upstream-runner
docker-build-upstream-runner: ## Build the used runner image
	$(eval RUNNER_IMAGE ?= $(shell echo "localhost:5000/runner:upstream-linux-ubuntu-22.04-$(shell uname -m)"))
	docker build -t $(RUNNER_IMAGE) ./runner/upstream
	docker push $(RUNNER_IMAGE)

.PHONY: template
template: ## Create the necessary configmap for the local development
ifeq ($(GARM_GITHUB_TOKEN),)
	$(error GARM_GITHUB_TOKEN is undefined)
endif
ifeq ($(GARM_GITHUB_REPOSITORY),)
	$(error GARM_GITHUB_REPOSITORY is undefined)
endif
ifeq ($(GARM_GITHUB_WEBHOOK_SECRET),)
	$(error GARM_GITHUB_WEBHOOK_SECRET is undefined)
endif
ifeq ($(GARM_GITHUB_ORGANIZATION),)
	$(error GARM_GITHUB_ORGANIZATION is undefined)
endif
	GARM_GITHUB_BASE_URL=$(GARM_GITHUB_BASE_URL) \
	GARM_GITHUB_API_BASE_URL=$(GARM_GITHUB_API_BASE_URL) \
	GARM_GITHUB_UPLOAD_BASE_URL=$(GARM_GITHUB_UPLOAD_BASE_URL) \
	GARM_GITHUB_REPOSITORY=$(GARM_GITHUB_REPOSITORY) \
	GARM_GITHUB_ORGANIZATION=$(GARM_GITHUB_ORGANIZATION) \
	GARM_GITHUB_TOKEN=$(shell echo -n $(GARM_GITHUB_TOKEN) | base64) \
	GARM_GITHUB_WEBHOOK_SECRET=$(shell echo -n $(GARM_GITHUB_WEBHOOK_SECRET) | base64) \
	RUNNER_IMAGE=$(shell echo "localhost:5000/runner:linux-ubuntu-22.04-$(shell uname -m)") \
	envsubst < hack/local-development/kubernetes/garm-operator-crs-envsubst.yaml > hack/local-development/kubernetes/garm-operator-crs.yaml

.PHONY: prepare-operator
prepare-operator: ## Prepare garm-operator for local development
	curl -L https://github.com/mercedes-benz/garm-operator/releases/download/v0.3.2/garm_operator_all.yaml | \
	GARM_SERVER_URL=http://garm-server.garm-server.svc:9997 \
	GARM_SERVER_USERNAME=admin \
	GARM_SERVER_PASSWORD=LmrBG1KcBOsDfNKq4cQTGpc0hJ0kejkk \
	OPERATOR_WATCH_NAMESPACE=garm-operator-system \
	envsubst > hack/local-development/kubernetes/garm-operator-all.yaml

.PHONY: kind-cluster
kind-cluster: $(KIND) ## Create a new kind cluster designed for local development
	hack/scripts/kind-with-registry.sh

.PHONY: delete-kind-cluster
delete-kind-cluster:
	$(KIND) delete cluster --name garm
	docker kill kind-registry && docker rm kind-registry

.PHONY: tilt-up
tilt-up: kind-cluster ## Start tilt and build kind cluster
	tilt up

##@ Release
.PHONY: release
release: goreleaser ## Create a new release
	$(GORELEASER) release --clean

##@ Lint / Verify
ALL_VERIFY_CHECKS = doctoc license security

.PHONY: verify
verify: $(addprefix verify-,$(ALL_VERIFY_CHECKS)) ## Run all verify-* targets

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Run linting.
	$(GOLANGCI_LINT) run -v $(GOLANGCI_LINT_EXTRA_ARGS)

.PHONY: lint-fix
lint-fix: $(GOLANGCI_LINT) ## Lint the codebase and run auto-fixers if supported by the linte
	GOLANGCI_LINT_EXTRA_ARGS=--fix $(MAKE) lint

.PHONY: verify-license
verify-license: ## Verify license headers
	./hack/scripts/verify-license.sh

.PHONY: verify-doctoc
verify-doctoc: generate-doctoc
	@if !(git diff --quiet HEAD); then \
		git diff; \
		echo "doctoc is out of date, run make generate-doctoc"; exit 1; \
	fi

.PHONY: verify-security
verify-security: govulncheck-scan nancy-scan ## Verify security by running govulncheck and nancy
	@echo "Security checks passed"

.PHONY: govulncheck-scan
govulncheck-scan: govulncheck ## Perform govulncheck scan
	$(GOVULNCHECK) ./...

.PHONY: nancy-scan
nancy-scan: nancy ## Perform nancy scan
	go list -json -deps ./... | $(NANCY) sleuth	

##@ Build Dependencies

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary. If wrong version is installed, it will be overwritten.
$(GOLANGCI_LINT): $(LOCALBIN)
	test -s $(LOCALBIN)/golangci-lint && $(LOCALBIN)/golangci-lint --version | grep -q $(GOLANGCI_LINT_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: goreleaser
goreleaser: $(GORELEASER) ## Download goreleaser locally if necessary. If wrong version is installed, it will be overwritten.
$(GORELEASER): $(LOCALBIN)
	test -s $(LOCALBIN)/goreleaser && $(LOCALBIN)/goreleaser --version | grep -q $(GORELEASER_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION)

.PHONY: mdtoc
mdtoc: $(MDTOC) ## Download mdtoc locally if necessary. If wrong version is installed, it will be overwritten.
$(MDTOC): $(LOCALBIN)
	test -s $(LOCALBIN)/mdtoc || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/mdtoc@$(MDTOC_VERSION)

.PHONY: nancy
nancy: $(NANCY) ## Download nancy locally if necessary. If wrong version is installed, it will be overwritten.
$(NANCY): $(LOCALBIN)
	test -s $(LOCALBIN)/nancy && $(LOCALBIN)/nancy --version | grep -q $(NANCY_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/sonatype-nexus-community/nancy@$(NANCY_VERSION)

.PHONY: govulncheck
govulncheck: $(GOVULNCHECK) ## Download govulncheck locally if necessary. If wrong version is installed, it will be overwritten.
$(GOVULNCHECK): $(LOCALBIN)
	test -s $(LOCALBIN)/govulncheck || \
	GOBIN=$(LOCALBIN) go install golang.org/x/vuln/cmd/govulncheck@latest

.PHONY: kind
kind: $(KIND) ## Download kind locally if necessary. If wrong version is installed, it will be overwritten.
$(KIND): $(LOCALBIN)
	test -s $(LOCALBIN)/kind && $(LOCALBIN)/kind version | grep -q $(KIND_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/kind@$(KIND_VERSION)

##@ Documentation
.PHONY: generate-doctoc
generate-doctoc: mdtoc ## Generate documentation table of contents
	grep --include='*.md' -rl -e '<!-- toc -->' $(git rev-parse --show-toplevel) | xargs $(MDTOC) -inplace -max-depth 3
