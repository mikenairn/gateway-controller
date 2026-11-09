IMG ?= gateway-controller:latest

# kind-up builds and loads this tag rather than IMG, so a local kind cluster
# is never accidentally running an image meant for a registry.
KIND_CLUSTER_NAME ?= gateway-controller-dev
KIND_IMG ?= gateway-controller:dev

# Container engine, docker by default as kubebuilder does, which keeps the
# docker-* target names honest. Override with CONTAINER_ENGINE=podman.
# CONTAINER_TOOL is the kubebuilder name and is honoured too.
CONTAINER_ENGINE ?= docker
CONTAINER_TOOL ?= $(CONTAINER_ENGINE)

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} \
		/^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate CRD and RBAC manifests.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate deepcopy implementations.
	$(CONTROLLER_GEN) object paths="./..."

# The verify-* targets are intended for CI against a clean checkout, and are
# run individually so a failure names the step that produced it. They check
# untracked files too: new generated output would otherwise pass unnoticed.
.PHONY: verify-manifests
verify-manifests: manifests ## Fail if generated manifests are not committed.
	$(call verify-clean,config,manifests)

.PHONY: verify-generate
verify-generate: generate ## Fail if generated deepcopy code is not committed.
	$(call verify-clean,'*zz_generated*',generate)

.PHONY: verify-tidy
verify-tidy: tidy ## Fail if go.mod or go.sum are not tidy.
	$(call verify-clean,go.mod go.sum,tidy)

.PHONY: verify
verify: verify-manifests verify-generate verify-tidy ## Run all verify-* targets.

# verify-clean fails when pathspec $1 has uncommitted or untracked changes,
# naming the make target $2 that regenerates it.
define verify-clean
@out="$$(git status --porcelain --untracked-files=all -- $(1))"; \
if [ -n "$$out" ]; then \
	echo "$(1) is out of date; run 'make $(2)' and commit the result:"; \
	echo "$$out"; \
	exit 1; \
fi
endef

.PHONY: fmt
fmt: ## Format Go sources.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet.
	go vet ./...

.PHONY: lint
lint: golangci-lint ## Run golangci-lint.
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint and apply fixes.
	$(GOLANGCI_LINT) run --fix

.PHONY: lint-config
lint-config: golangci-lint ## Verify the golangci-lint configuration.
	$(GOLANGCI_LINT) config verify

.PHONY: lint-extra
lint-extra: actionlint shellcheck ## Lint things Go tooling does not cover.
	$(ACTIONLINT)
	$(SHELLCHECK) hack/*.sh

.PHONY: tidy
tidy: ## Tidy go.mod and go.sum.
	go mod tidy

##@ Test

# Unit tests are hermetic: no cluster, no envtest binaries, no network.
.PHONY: test
test: ## Run unit tests.
	go test $(GO_TEST_FLAGS) ./api/... ./cmd/... ./internal/...

.PHONY: test-integration
test-integration: manifests generate setup-envtest ## Run integration tests against envtest.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" \
		go test $(GO_TEST_FLAGS) ./test/integration/...

.PHONY: test-e2e
test-e2e: ## Run e2e tests against a kind cluster (see kind-up).
	go test $(GO_TEST_FLAGS) -timeout 30m ./test/e2e/...

.PHONY: coverage
coverage: setup-envtest ## Run unit and integration tests with coverage.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" \
		go test -coverprofile=cover.out -covermode=atomic -coverpkg=./api/...,./internal/... \
		./api/... ./cmd/... ./internal/... ./test/integration/...
	go tool cover -func=cover.out | tail -1

.PHONY: kind-up
kind-up: ## Create a kind cluster with the manager built and deployed.
	$(MAKE) docker-build IMG=$(KIND_IMG)
	IMG=$(KIND_IMG) CONTAINER_ENGINE=$(CONTAINER_ENGINE) \
		KIND_CLUSTER_NAME=$(KIND_CLUSTER_NAME) bash hack/setup-kind.sh

.PHONY: kind-down
kind-down: ## Delete the kind cluster.
	KIND_CLUSTER_NAME=$(KIND_CLUSTER_NAME) bash hack/teardown-kind.sh

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build the manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run the manager against the configured cluster.
	go run cmd/main.go

.PHONY: docker-build
docker-build: ## Build the manager image.
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push the manager image.
	$(CONTAINER_TOOL) push ${IMG}

##@ Deployment

.PHONY: install
install: manifests kustomize ## Install the CRDs into the current cluster.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Remove the CRDs from the current cluster.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy the manager to the current cluster.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: kustomize ## Remove the manager from the current cluster.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found -f -

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
ENVTEST ?= $(LOCALBIN)/setup-envtest
KUSTOMIZE ?= $(LOCALBIN)/kustomize
ACTIONLINT ?= $(LOCALBIN)/actionlint
SHELLCHECK ?= $(LOCALBIN)/shellcheck

## Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.23.0
GOLANGCI_LINT_VERSION ?= v2.15.0
KUSTOMIZE_VERSION ?= v5.8.1
ACTIONLINT_VERSION ?= v1.7.12
SHELLCHECK_VERSION ?= v0.11.0
# Resolved from go.mod so the envtest tooling cannot drift from the libraries
# it must match.
ENVTEST_VERSION ?= $(call gomodver,sigs.k8s.io/controller-runtime)
ENVTEST_K8S_VERSION ?= $(shell printf '%s' '$(call gomodver,k8s.io/api)' | sed -E 's/^v?[0-9]+\.([0-9]+).*/1.\1/')

define gomodver
$(shell go list -m -f '{{if .Replace}}{{.Replace.Version}}{{else}}{{.Version}}{{end}}' $(1) 2>/dev/null)
endef

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Install kustomize.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

# shellcheck ships per-OS/arch tarballs rather than as a Go module, so it
# cannot use go-install-tool.
.PHONY: shellcheck
shellcheck: $(SHELLCHECK) ## Install shellcheck.
$(SHELLCHECK): $(LOCALBIN)
	@[ -f "$(SHELLCHECK)-$(SHELLCHECK_VERSION)" ] || { \
		set -e ;\
		os="$$(uname -s | tr '[:upper:]' '[:lower:]')" ;\
		arch="$$(uname -m)"; [ "$$arch" = "arm64" ] && arch=aarch64 ;\
		echo "Downloading shellcheck $(SHELLCHECK_VERSION) ($$os.$$arch)" ;\
		curl -sSL "https://github.com/koalaman/shellcheck/releases/download/$(SHELLCHECK_VERSION)/shellcheck-$(SHELLCHECK_VERSION).$$os.$$arch.tar.xz" \
			| tar xJ -C "$(LOCALBIN)" --strip-components=1 "shellcheck-$(SHELLCHECK_VERSION)/shellcheck" ;\
		mv "$(LOCALBIN)/shellcheck" "$(SHELLCHECK)-$(SHELLCHECK_VERSION)" ;\
	}
	@ln -sf "$(SHELLCHECK)-$(SHELLCHECK_VERSION)" "$(SHELLCHECK)"

.PHONY: actionlint
actionlint: $(ACTIONLINT) ## Install actionlint.
$(ACTIONLINT): $(LOCALBIN)
	$(call go-install-tool,$(ACTIONLINT),github.com/rhysd/actionlint/cmd/actionlint,$(ACTIONLINT_VERSION))


.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Install controller-gen.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Install golangci-lint.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: setup-envtest
setup-envtest: envtest ## Download the binaries required for ENVTEST.
	@echo "Setting up envtest binaries for Kubernetes version $(ENVTEST_K8S_VERSION)..."
	@"$(ENVTEST)" use $(ENVTEST_K8S_VERSION) --bin-dir "$(LOCALBIN)" -p path || { \
		echo "Error: Failed to set up envtest binaries for version $(ENVTEST_K8S_VERSION)."; \
		exit 1; \
	}

.PHONY: envtest
envtest: $(ENVTEST) ## Install setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

# go-install-tool installs $2@$3 as $1-$3 and links $1 to it, so that changing
# a pinned version reinstalls rather than silently reusing the old binary.
define go-install-tool
@[ -f "$(1)-$(3)" ] && [ "$$(readlink -- "$(1)" 2>/dev/null)" = "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f "$(1)" ;\
GOBIN="$(LOCALBIN)" go install $${package} ;\
mv "$(LOCALBIN)/$$(basename "$(1)")" "$(1)-$(3)" ;\
} ;\
ln -sf "$(1)-$(3)" "$(1)"
endef
