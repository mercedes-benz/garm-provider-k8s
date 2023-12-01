# SPDX-License-Identifier: MIT

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

GARM_GITHUB_NAME ?= ""
GARM_GITHUB_OAUTH_TOKEN ?= ""
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

GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.53.3

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
	@mkdir -p ./hack/bin
	@cp "$(BIN_DIR)/$(BINARY_NAME)" "./hack/bin/$(BINARY_NAME)"

clean:
	@echo "Cleaning up $(BIN_DIR)..."
	@rm -rf $(BIN_DIR)

DEBUG = 0

##@ Development
.PHONY: docker-build
docker-build:
	docker build -t garm-with-k8s:latest ./hack
	docker push localhost:5000/garm-with-k8s:latest

.PHONY: docker-build-runner
docker-build-runner:
	docker build -t $(RUNNER_IMAGE) ./runner
	docker push $(RUNNER_IMAGE)

.PHONY: template
template:
ifeq ($(GARM_GITHUB_NAME),)
	$(error GARM_GITHUB_NAME is undefined)
endif
ifeq ($(GARM_GITHUB_OAUTH_TOKEN),)
	$(error GARM_GITHUB_OAUTH_TOKEN is undefined)
endif
	GARM_GITHUB_BASE_URL=$(GARM_GITHUB_BASE_URL) \
	GARM_GITHUB_API_BASE_URL=$(GARM_GITHUB_API_BASE_URL) \
	GARM_GITHUB_UPLOAD_BASE_URL=$(GARM_GITHUB_UPLOAD_BASE_URL) \
	envsubst < hack/local-development/kubernetes/configmap-envsubst.yaml > hack/local-development/kubernetes/configmap.yaml
	

.PHONY: apply
apply:
	kustomize build k8s_deployment | sed 's/<ARCH>/$(ARCH)/' | kubectl apply -f -

.PHONY: kind-cluster
kind-cluster: ## Create a new kind cluster designed for local development
	hack/scripts/kind-with-registry.sh

.PHONY: delete-kind-cluster
delete-kind-cluster:
	kind delete cluster --name garm-operator
	docker kill kind-registry && docker rm kind-registry

.PHONY: tilt-up
tilt-up: kind-cluster ## Start tilt and build kind cluster
	tilt up

##@ Release
.PHONY: release
release: goreleaser ## Create a new release
	$(GORELEASER) release --clean

##@ Lint / Verify
.PHONY: lint
lint: $(GOLANGCI_LINT) ## Run linting.
	$(GOLANGCI_LINT) run -v $(GOLANGCI_LINT_EXTRA_ARGS)

.PHONY: lint-fix
lint-fix: $(GOLANGCI_LINT) ## Lint the codebase and run auto-fixers if supported by the linte
	GOLANGCI_LINT_EXTRA_ARGS=--fix $(MAKE) lint

.PHONY: verify-license
verify-license: ## Verify license headers
	./hack//scripts/verify-license.sh
