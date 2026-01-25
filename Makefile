# sourced by https://github.com/octomation/makefiles

.DEFAULT_GOAL = check
GIT_HOOKS     = post-merge pre-commit pre-push
GO_VERSIONS   = 1.27

AT    := @
ARCH  := $(shell uname -m | tr '[:upper:]' '[:lower:]')
OS    := $(shell uname -s | tr '[:upper:]' '[:lower:]')
DATE  := $(shell date -u +%Y-%m-%dT%T%Z)
SHELL := /usr/bin/env bash -euo pipefail -c

find-todos:
	$(AT) grep \
		--exclude-dir={bin,docs} \
		--exclude-dir={node_modules,vendor} \
		--exclude-dir={.docz,.next} \
		--color \
		--text \
		-inRo -E ' TODO:.*|SkipNow' . | grep -v ' TODO:.*|SkipNow' || true
.PHONY: find-todos

help:
	@LC_ALL=C $(MAKE) -pRrq -f $(firstword $(MAKEFILE_LIST)) : 2>/dev/null \
	| awk -v RS= -F: '/(^|\n)# Files(\n|$$)/,/(^|\n)# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' \
	| sort \
	| egrep -v -e '^[^[:alnum:]]' -e '^$@$$' || true
.PHONY: help

make-verbose:
	$(eval AT :=) $(eval MAKE := $(MAKE) verbose) @true
.PHONY: make-verbose

COMMIT  := $(shell git rev-parse --short --verify HEAD)
RELEASE := $(shell git describe --tags 2>/dev/null | rev | cut -d - -f3- | rev)

git-config:
	$(AT) git config core.autocrlf input
	$(AT) git config core.hooksPath .github/hooks
.PHONY: git-config

ifneq (, $(wildcard .github/hooks/))
ifdef GIT_HOOKS

# versioned hooks take effect in place, see octomation/makefiles#81
git-hooks:
	$(AT) git config --local core.hooksPath .github/hooks
.PHONY: git-hooks

git-unhook:
	$(AT) git config --local --unset core.hooksPath || true
.PHONY: git-unhook

define hook_tpl
git-$(1):
	$$(AT) .github/hooks/$(1) < /dev/null
.PHONY: git-$(1)
endef

render_hook_tpl = $(eval $(call hook_tpl,$(hook)))
$(foreach hook,$(GIT_HOOKS),$(render_hook_tpl))

endif
else
git-hooks:
	@echo have no git hooks
.PHONY: git-hooks

git-unhook: ;
.PHONY: git-unhook
endif

git-clean:
	$(AT) git clean -fdx
.PHONY: git-clean

git-check:
	$(AT) git diff --exit-code >/dev/null
	$(AT) git diff --cached --exit-code >/dev/null
	$(AT) ! git ls-files --others --exclude-standard | grep -q ^
.PHONY: git-check

git-rmdir:
	$(AT) for dir in `git ls-files --others --exclude-standard --directory`; do \
		find $${dir%%/} -depth -type d -empty | xargs rmdir; \
	done
.PHONY: git-rmdir

GOBIN       ?= $(PWD)/bin/$(OS)/$(ARCH)
GOFLAGS     ?= -mod=
GOPRIVATE   ?= go.octolab.org
GOPROXY     ?= https://proxy.golang.org,direct
GOTEST      ?= $(shell PATH="$(PATH)" command -v testit)
GOTESTFLAGS ?=
GOTRACEBACK ?= all
LOCAL       ?= $(MODULE)
MODULE      ?= `go list $(GOFLAGS)`
PACKAGES    ?= `go list $(GOFLAGS) ./...`
PATHS       ?= $(shell echo $(PACKAGES) | sed -e "s|$(MODULE)/||g" | sed -e "s|$(MODULE)|$(PWD)/*.go|g")
TIMEOUT     ?= 2m

ifeq (, $(GOTEST))
	GOTEST = go test
else
	GOTEST := $(GOTEST) go --colored --stacked
endif

ifeq (, $(PACKAGES))
	PACKAGES = $(MODULE)
endif

ifeq (, $(PATHS))
	PATHS = .
endif

export GOBIN       := $(GOBIN)
export GOFLAGS     := $(GOFLAGS)
export GOPRIVATE   := $(GOPRIVATE)
export GOPROXY     := $(GOPROXY)
export GOTRACEBACK := $(GOTRACEBACK)

go-env:
	@echo "GO111MODULE: $(strip `go env GO111MODULE`)"
	@echo "GOBIN:       $(strip `go env GOBIN`)"
	@echo "GOFLAGS:     $(strip `go env GOFLAGS`)"
	@echo "GOPRIVATE:   $(strip `go env GOPRIVATE`)"
	@echo "GOPROXY:     $(strip `go env GOPROXY`)"
	@echo "GOTEST:      $(GOTEST)"
	@echo "GOTESTFLAGS: $(GOTESTFLAGS)"
	@echo "GOTRACEBACK: $(GOTRACEBACK)"
	@echo "GOVERSION:   $(strip `go version | awk '{print $$3}' | sed -e 's|^go||g'`)"
	@echo "LOCAL:       $(LOCAL)"
	@echo "MODULE:      $(MODULE)"
	@echo "PACKAGES:    $(PACKAGES)"
	@echo "PATHS:       $(strip $(PATHS))"
	@echo "TIMEOUT:     $(TIMEOUT)"
.PHONY: go-env

go-verbose:
	$(eval GOTESTFLAGS := -v) @true
.PHONY: go-verbose

go-deps-check:
	$(AT) go mod verify
	$(AT) govulncheck -tags $(GOTAGS) -test ./...
	$(AT) if command -v egg >/dev/null; then \
		egg deps check license; \
		egg deps check version; \
	fi
.PHONY: go-deps-check

go-deps-clean:
	$(AT) go clean -modcache
.PHONY: go-deps-clean

go-deps-fetch:
	$(AT) go mod download
	$(AT) if [[ "`go env GOFLAGS`" =~ -mod=vendor ]]; then go mod vendor; fi
.PHONY: go-deps-fetch

go-deps-tidy:
	$(AT) go mod tidy
	$(AT) if [[ "`go env GOFLAGS`" =~ -mod=vendor ]]; then go mod vendor; fi
.PHONY: go-deps-tidy

go-deps-update: selector = '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}'
go-deps-update:
	$(AT) if command -v egg >/dev/null; then \
		packages="`egg deps list | tr ' ' '\n'`"; \
	else \
		packages="`go list -f $(selector) -m -mod=readonly all`"; \
	fi; \
	if [ -z "$$packages" ]; then exit; fi; \
	for package in $$packages; do \
		go mod edit -require=$$package@latest; \
		go mod tidy; \
	done
	$(AT) $(MAKE) go-deps-tidy
.PHONY: go-deps-update

go-docs:
	$(AT) go doc -all .
.PHONY: go-docs

go-fmt:
	$(AT) goimports -local $(LOCAL) -w $(PATHS)
.PHONY: go-fmt

go-generate:
	$(AT) go generate $(PACKAGES)
.PHONY: go-generate

# How to work with linters
#
#  ```bash
#  $ golangci-lint linters --out-format=json | jq '.[] | {name: .name, url: .origin}'
#  $ golangci-lint linters --out-format=json | jq '.[] | select(.presets[] | contains ("bugs"))'
#  $ golangci-lint linters --out-format=json | jq '.[] | select(.origin == "")'
#  ```
#
# For more experience see https://jqplay.org, https://stedolan.github.io/jq/.
go-lint:
	$(AT) golangci-lint run --fast-only ./...
.PHONY: go-lint

go-pkg:
	$(AT) open https://pkg.go.dev/$(MODULE)@$(RELEASE)
.PHONY: go-pkg

go-check:
	$(AT) $(GOTEST) -count=1 -test.run=^$(uuidgen) ./...
.PHONY: go-check

go-test:
	$(AT) $(GOTEST) -race -timeout $(TIMEOUT) $(GOTESTFLAGS) $(PACKAGES)
.PHONY: go-test

go-test-clean:
	$(AT) go clean -testcache
.PHONY: go-test-clean

go-test-with-coverage:
	$(AT) $(GOTEST) \
		-cover \
		-covermode atomic \
		-coverprofile c.out \
		-race \
		-timeout $(TIMEOUT) \
		$(GOTESTFLAGS) \
		$(PACKAGES)
.PHONY: go-test-with-coverage

go-test-with-coverage-report: go-test-with-coverage
	$(AT) go tool cover -html c.out
.PHONY: go-test-with-coverage-report

go-integration-test: GOTAGS := $(GOTAGS) integration
go-integration-test:
	$(AT) $(GOTEST) -race -tags $(GOTAGS) -timeout $(TIMEOUT) $(GOTESTFLAGS) $(PACKAGES)
.PHONY: go-test-integration-quick

go-integration-test-with-coverage: GOTAGS := $(GOTAGS) integration
go-integration-test-with-coverage:
	$(AT) $(GOTEST) \
		-cover \
		-covermode atomic \
		-coverprofile c.out \
		-race \
		-tags $(GOTAGS) \
		-timeout $(TIMEOUT) \
		$(GOTESTFLAGS) \
		$(PACKAGES)
.PHONY: go-integration-test-with-coverage

go-integration-test-with-coverage-report: go-integration-test-with-coverage
	$(AT) go tool cover -html c.out
.PHONY: go-integration-test-with-coverage-report

go-fuzzing-test:
	$(AT) $(GOTEST) -fuzz $(PACKAGES)
.PHONY: go-fuzzing-test

go-fuzzing-test-clean:
	$(AT) go clean -fuzzcache
.PHONY: go-fuzzing-test-clean

go-performance-test:
	$(AT) $(GOTEST) -bench=. -run=^$(uuidgen) $(GOTESTFLAGS) $(PACKAGES)
.PHONY: go-performance-test

BINARY  ?= $(GOBIN)/$(shell basename $(MAIN))
LDFLAGS ?= -ldflags "-s -w -X main.version=$(RELEASE) -X main.commit=$(COMMIT) -X main.date=$(DATE)"
MAIN    ?= $(MODULE)

go-build-env:
	@echo "DATE:        $(DATE)"
	@echo "COMMIT:      $(COMMIT)"
	@echo "RELEASE:     $(RELEASE)"
	@echo "BINARY:      $(BINARY)"
	@echo "LDFLAGS:     $(LDFLAGS)"
	@echo "MAIN:        $(MAIN)"
.PHONY: go-build-env

go-build:
	$(AT) go build -o $(BINARY) $(LDFLAGS) $(MAIN)
.PHONY: go-build

go-build-with-race:
	$(AT) go build -race -o $(BINARY) $(LDFLAGS) $(MAIN)
.PHONY: go-build-with-race

go-install:
	$(AT) go install $(LDFLAGS) $(MAIN)
.PHONY: go-install

go-install-clean:
	$(AT) go clean -cache
.PHONY: go-install-clean

go-dist-check:
	$(AT) goreleaser --clean --skip=publish --snapshot
.PHONY: go-dist-check

config-vet:
	$(AT) cue vet -c .github/settings.cue .github/settings.json
.PHONY: config-vet

doctor:
	$(AT) node .github/scripts/release.mjs doctor
.PHONY: doctor

# runs the doctor workflow on main and waits for its result: unlike doctor, it
# mints the tap token of the OctoLab Releaser App, which proves it can publish
workflow-doctor:
	$(AT) last() { gh run list -w doctor.yml -e workflow_dispatch -L 1 --json databaseId -q '.[0].databaseId'; }; \
	prev="$$(last)"; \
	gh workflow run doctor.yml --ref main -f reason="$(or $(REASON),make workflow-doctor)"; \
	until run="$$(last)" && [ -n "$$run" ] && [ "$$run" != "$$prev" ]; do sleep 3; done; \
	gh run watch --exit-status "$$run"
.PHONY: workflow-doctor

release-check: config-vet
	$(AT) test -n "$(TAG)" || { echo 'usage: git tag vX.Y.Z && make release-check TAG=vX.Y.Z'; exit 2; }
	$(AT) node .github/scripts/release.mjs check $(TAG) --local
	$(AT) $(MAKE) go-deps-tidy go-tools-tidy git-check
	$(AT) $(MAKE) release-config-check
.PHONY: release-check

go-dist-clean:
	$(AT) rm -rf dist
.PHONY: go-dist-clean

go-dist-installer:
	$(AT) godownloader .goreleaser.yml >bin/install
.PHONY: go-dist-installer

TOOLFLAGS ?= -mod=

go-tools-env:
	@echo "GOBIN:       `go env GOBIN`"
	@echo "TOOLFLAGS:   $(TOOLFLAGS)"
.PHONY: go-tools-env

ifneq (, $(wildcard ./tools/))
go-tools-check: GOFLAGS = $(TOOLFLAGS)
go-tools-check:
	$(AT) cd tools; \
	go mod verify; \
	govulncheck tool; \
	if command -v egg >/dev/null; then \
		egg deps check license; \
		egg deps check version; \
	fi
.PHONY: go-tools-check

go-tools-fetch: GOFLAGS = $(TOOLFLAGS)
go-tools-fetch:
	$(AT) cd tools; go mod download; \
	if [[ "`go env GOFLAGS`" =~ -mod=vendor ]]; then go mod vendor; fi
.PHONY: go-tools-fetch

go-tools-install: GOFLAGS = $(TOOLFLAGS)
go-tools-install: go-tools-fetch
	$(AT) cd tools; go install tool
.PHONY: go-tools-install

go-tools-tidy: GOFLAGS = $(TOOLFLAGS)
go-tools-tidy:
	$(AT) cd tools; go mod tidy; \
	if [[ "`go env GOFLAGS`" =~ -mod=vendor ]]; then go mod vendor; fi
.PHONY: go-tools-tidy

# modules behind the tool directive are required as indirect,
# so the tool pattern is the only way to select them
go-tools-update: GOFLAGS = $(TOOLFLAGS)
go-tools-update:
	$(AT) cd tools; go get tool
	$(AT) $(MAKE) go-tools-tidy go-tools-install
.PHONY: go-tools-update
else
go-tools-disabled:
	@echo have no tools
.PHONY: go-tools-disabled

go-tools-check: go-tools-disabled
	@true
.PHONY: go-tools-check

go-tools-fetch: go-tools-disabled
	@true
.PHONY: go-tools-fetch

go-tools-install: go-tools-disabled
	@true
.PHONY: go-tools-install

go-tools-tidy: go-tools-disabled
	@true
.PHONY: go-tools-tidy

go-tools-update: go-tools-disabled
	@true
.PHONY: go-tools-update
endif

ifneq (, $(shell PATH="$(PATH)" command -v docker))
ifdef GO_VERSIONS

define go_tpl
go-$(1):
	$$(AT) docker run \
		--rm -it \
		-v $(PWD):/src \
		-w /src \
		golang:$(1) bash
.PHONY: go-$(1)
endef

render_go_tpl = $(eval $(call go_tpl,$(version)))
$(foreach version,$(GO_VERSIONS),$(render_go_tpl))

endif
endif

export PATH := $(GOBIN):$(PATH)

setup: git-config git-hooks go-deps-fetch go-tools-install
.PHONY: setup

clean: go-deps-clean go-dist-clean go-install-clean go-test-clean go-fuzzing-test-clean
.PHONY: clean

drop: clean git-clean git-rmdir
.PHONY: drop

env: go-env go-build-env go-tools-env
env:
	@echo "PATH:        $(PATH)"
.PHONY: env

deps: go-deps-fetch
.PHONY: deps

tools: go-tools-install
.PHONY: tools

update: go-deps-update go-tools-update
.PHONY: update

format: go-fmt
.PHONY: format

generate: go-generate go-fmt
.PHONY: generate

docs:
	@true # TODO:implement go-docs go-pkg
.PHONY: docs

lint: go-lint
.PHONY: lint

test: go-test
.PHONY: test

test-with-coverage: go-test-with-coverage
.PHONY: test-with-coverage

check: lint test
.PHONY: check

fast-check: find-todos
fast-check: go-check go-deps-check
fast-check: go-deps-tidy go-generate git-check
.PHONY: fast-check

full-check: fast-check go-dist-check check
.PHONY: full-check

full-test: go-test go-fuzzing-test go-integration-test go-performance-test
.PHONY: full-test

install: go-install
.PHONY: install

verbose: make-verbose go-verbose
.PHONY: verbose

# Repository checks shared with GitHub Actions.
source-check:
	$(AT) go vet ./...
	$(AT) test -z "$$(gofmt -l $$(git ls-files '*.go' ':!:tools/**'))"
.PHONY: source-check

release-config-check:
	$(AT) node .github/scripts/goreleaser-check.mjs
.PHONY: release-config-check

# Tools-module vulnerability checks run in tools.yml, independently of releases.
release-verify: config-vet source-check go-deps-check go-deps-tidy go-generate git-check
.PHONY: release-verify

verify: go-deps-check generate check git-check
.PHONY: verify

# Names before the go-* scheme: the v0.2.0 release tasks still use them, drop them after it.
build: go-build
build-env: go-build-env
build-with-race: go-build-with-race
deps-check: go-deps-check
deps-clean: go-deps-clean
deps-fetch: go-deps-fetch
deps-tidy: go-deps-tidy
deps-update: go-deps-update
dist-check: go-dist-check
dist-installer: go-dist-installer
hooks: git-hooks
init: setup check
install-clean: go-install-clean
post-merge: git-post-merge
pre-commit: git-pre-commit
pre-push: git-pre-push
refresh: go-deps-tidy update deps tools generate check go-build
test-clean: go-test-clean
test-integration: go-integration-test-with-coverage
test-integration-quick: go-integration-test
test-integration-report: go-integration-test-with-coverage-report
test-with-coverage-report: go-test-with-coverage-report
todo: find-todos
tools-check: go-tools-check
tools-env: go-tools-env
tools-fetch: go-tools-fetch
tools-install: go-tools-install
tools-tidy: go-tools-tidy
tools-update: go-tools-update
unhook: git-unhook
$(foreach version,$(GO_VERSIONS),$(eval go$(version): go-$(version)))
.PHONY: build build-env build-with-race deps-check deps-clean deps-fetch deps-tidy
.PHONY: deps-update dist-check dist-installer hooks init install-clean post-merge
.PHONY: pre-commit pre-push refresh test-clean test-integration test-integration-quick
.PHONY: test-integration-report test-with-coverage-report todo tools-check tools-env
.PHONY: tools-fetch tools-install tools-tidy tools-update unhook $(GO_VERSIONS:%=go%)

test-quick:
	$(AT) $(GOTEST) -timeout $(TIMEOUT) $(GOTESTFLAGS) $(PACKAGES)
.PHONY: test-quick
