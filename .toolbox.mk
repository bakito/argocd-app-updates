## toolbox - start
## Generated with https://github.com/bakito/toolbox

## Current working directory
TB_LOCALDIR ?= $(shell which cygpath > /dev/null 2>&1 && cygpath -m $$(pwd) || pwd)
## Location to install dependencies to
TB_LOCALBIN ?= $(TB_LOCALDIR)/bin
$(TB_LOCALBIN):
	mkdir -p $(TB_LOCALBIN)

## Tool Binaries
TB_DEEPCOPY_GEN ?= $(TB_LOCALBIN)/deepcopy-gen
TB_GOLANGCI_LINT ?= $(TB_LOCALBIN)/golangci-lint
TB_SEMVER ?= $(TB_LOCALBIN)/semver

## Tool Versions
TB_DEEPCOPY_GEN_VERSION ?= v0.31.1
TB_GOLANGCI_LINT_VERSION ?= v1.61.0
TB_SEMVER_VERSION ?= v1.1.3

## Tool Installer
.PHONY: tb.deepcopy-gen
tb.deepcopy-gen: $(TB_DEEPCOPY_GEN) ## Download deepcopy-gen locally if necessary.
$(TB_DEEPCOPY_GEN): $(TB_LOCALBIN)
	test -s $(TB_LOCALBIN)/deepcopy-gen || GOBIN=$(TB_LOCALBIN) go install k8s.io/code-generator/cmd/deepcopy-gen@$(TB_DEEPCOPY_GEN_VERSION)
.PHONY: tb.golangci-lint
tb.golangci-lint: $(TB_GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(TB_GOLANGCI_LINT): $(TB_LOCALBIN)
	test -s $(TB_LOCALBIN)/golangci-lint || GOBIN=$(TB_LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(TB_GOLANGCI_LINT_VERSION)
.PHONY: tb.semver
tb.semver: $(TB_SEMVER) ## Download semver locally if necessary.
$(TB_SEMVER): $(TB_LOCALBIN)
	test -s $(TB_LOCALBIN)/semver || GOBIN=$(TB_LOCALBIN) go install github.com/bakito/semver@$(TB_SEMVER_VERSION)

## Reset Tools
.PHONY: tb.reset
tb.reset:
	@rm -f \
		$(TB_LOCALBIN)/deepcopy-gen \
		$(TB_LOCALBIN)/golangci-lint \
		$(TB_LOCALBIN)/semver

## Update Tools
.PHONY: tb.update
tb.update: tb.reset
	toolbox makefile -f $(TB_LOCALDIR)/Makefile \
		k8s.io/code-generator/cmd/deepcopy-gen@github.com/kubernetes/code-generator \
		github.com/golangci/golangci-lint/cmd/golangci-lint \
		github.com/bakito/semver
## toolbox - end