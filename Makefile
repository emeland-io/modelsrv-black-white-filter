SHELL := /bin/bash

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

GOVERSION ?= go1.26.1

GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.11.3

MOCKGEN = $(LOCALBIN)/mockgen
MOCKGEN_VERSION ?= v0.6.0

.PHONY: all
all: fmt vet test lint ## Format, vet, test, and lint.

.PHONY: build
build: ## Build the filtered modelserver binary.
	GOTOOLCHAIN=$(GOVERSION) go build -o $(LOCALBIN)/modelsrv-bw-filter ./cmd/main.go

.PHONY: fmt
fmt: ## Run go fmt.
	GOTOOLCHAIN=$(GOVERSION) go fmt ./...

.PHONY: vet
vet: ## Run go vet.
	GOTOOLCHAIN=$(GOVERSION) go vet ./...

.PHONY: generate
generate: ## Run go generate (no-op until //go:generate directives exist).
	GOTOOLCHAIN=$(GOVERSION) go generate ./...

.PHONY: gen
gen: generate ## Alias for generate.

.PHONY: test
test: fmt vet generate ## Run tests with coverage profile (cover.out).
	GOTOOLCHAIN=$(GOVERSION) go test ./... -coverprofile=cover.out -covermode=atomic

.PHONY: test-race
test-race: fmt vet generate ## Run tests with the race detector.
	GOTOOLCHAIN=$(GOVERSION) go test ./... -race

.PHONY: lint
lint: golangci-lint ## Run golangci-lint.
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint with fixes.
	$(GOLANGCI_LINT) run --fix

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: tools
tools: $(MOCKGEN) ## Install dev tools (mockgen) into ./bin.
$(MOCKGEN): $(LOCALBIN)
	$(call go-install-tool,$(MOCKGEN),go.uber.org/mock/mockgen,$(MOCKGEN_VERSION))

.PHONY: clean
clean: ## Remove build artifacts and coverage profile.
	rm -f cover.out
	rm -f $(LOCALBIN)/modelsrv-bw-filter

define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOTOOLCHAIN=$(GOVERSION) GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef

.PHONY: help
help: ## Show targets.
	@grep -E '^[a-zA-Z0-9_.-]+:.*?##' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'
