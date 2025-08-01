CURRENT=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
ROOT:=$(CURRENT)

SED ?= sed
REALPATH ?= realpath

K3D ?= k3d

ifeq ($(shell uname),Darwin)
	SED ?= gsed
	REALPATH ?= grealpath
endif

KUBERNETES_VERSION_MINOR:=31
KUBERNETES_VERSION_PATCH:=8

ENVOY_IMAGE=envoyproxy/envoy:v1.32.5

PROJECT := arangodb_operator
SCRIPTDIR := $(shell pwd)
ROOTDIR := $(shell cd $(SCRIPTDIR) && pwd)
VERSION := $(shell cat $(ROOTDIR)/VERSION)
VERSION_MAJOR_MINOR_PATCH := $(shell echo $(VERSION) | cut -f 1 -d '+')
VERSION_MAJOR_MINOR := $(shell echo $(VERSION_MAJOR_MINOR_PATCH) | cut -f 1,2 -d '.')
VERSION_MAJOR := $(shell echo $(VERSION_MAJOR_MINOR) | cut -f 1 -d '.')
COMMIT := $(shell git rev-parse --short HEAD)
DOCKERCLI := $(shell which docker)
RELEASE_MODE ?= community

MAIN_DIR := $(ROOT)/pkg/entry/$(RELEASE_MODE)

ifndef KEEP_GOPATH
	GOBUILDDIR := $(SCRIPTDIR)/.gobuild
	GOPATH := $(GOBUILDDIR)
else
	GOBUILDDIR := $(GOPATH)
endif

SRCDIR := $(SCRIPTDIR)
CACHEVOL := $(PROJECT)-gocache
BINDIR := $(ROOTDIR)/bin
VBINDIR := $(BINDIR)/$(RELEASE_MODE)
VENDORDIR := $(ROOTDIR)/deps
DASHBOARDDIR := $(ROOTDIR)/dashboard
LOCALDIR := $(ROOT)/local

ORGPATH := github.com/arangodb
ORGDIR := $(GOBUILDDIR)/src/$(ORGPATH)
REPONAME := kube-arangodb
REPODIR := $(ORGDIR)/$(REPONAME)
REPOPATH := $(ORGPATH)/$(REPONAME)

K3D_KUBECONFIG = $(GOBUILDDIR)/.kubeconfig
K3D_CLUSTER ?= $(REPONAME)

include $(ROOT)/$(RELEASE_MODE).mk

TEST_BUILD ?= 0
GOBUILDARGS ?=
GOBASEVERSION := 1.22.3
GOVERSION := $(GOBASEVERSION)-alpine3.18
DISTRIBUTION := alpine:3.15
GOCOMPAT := $(shell sed -En 's/^go (.*)$$/\1/p' go.mod)

GOBUILDTAGS := $(RELEASE_MODE)

ifeq ($(TEST_BUILD),1)
GOBUILDTAGS := $(GOBUILDTAGS),test_build
endif

PULSAR := $(GOBUILDDIR)/bin/pulsar$(shell go env GOEXE)
GOASSETSBUILDER := $(GOBUILDDIR)/bin/go-assets-builder$(shell go env GOEXE)

BUILDTIME = $(shell go run "$(ROOT)/tools/dategen/")

GOBUILDLDFLAGS := -X $(REPOPATH)/pkg/version.version=$(VERSION) -X $(REPOPATH)/pkg/version.buildDate=$(BUILDTIME) -X $(REPOPATH)/pkg/version.build=$(COMMIT)
GOBUILDGCFLAGS :=

# Go Strip Section
GOBUILDSTRIP ?= 1
ifeq ($(GOBUILDSTRIP),1)
GOBUILDLDFLAGS += -w -s
endif

# Go Disable function inlining
GOBUILDDISABLEFUNCTIONINLINING ?= 1
ifeq ($(GOBUILDDISABLEFUNCTIONINLINING),1)
GOBUILDGCFLAGS += -l
endif

# Go Disable bound checks
GOBUILDDISABLEBOUNDCHECKS ?= 1
ifeq ($(GOBUILDDISABLEBOUNDCHECKS),1)
GOBUILDGCFLAGS += -B
endif

HELM ?= $(shell which helm)

UPPER = $(shell echo '$1' | tr '[:lower:]' '[:upper:]')
LOWER = $(shell echo '$1' | tr '[:upper:]' '[:lower:]')
UPPER_ENV = $(shell echo '$1' | tr '[:lower:]' '[:upper:]' | tr -d '-')

.PHONY: helm
helm:
ifeq ($(HELM),)
	$(error "Before templating you need to install helm in PATH or export helm binary using 'export HELM=<path to helm>'")
endif

HELM_OPTIONS = --set "operator.image=$(OPERATORIMAGE)" \
	--set "operator.imagePullPolicy=Always" \
	--set "operator.resources=null" \
	--set "operator.debug=$(DEBUG)"

ifeq ($(shell $(HELM) version --client --template '{{.Version}}' 2> /dev/null | cut -f 1 -d '.'),v3)
	# Using helm v3
	HELM_PACKAGE_CMD = $(HELM) package "$(ROOTDIR)/chart/$(CHART_NAME)" -d "$(ROOTDIR)/bin/charts" \
		--version "$(VERSION_MAJOR_MINOR_PATCH)"

	HELM_CMD = $(HELM) template $(NAME) "$(ROOTDIR)/chart/$(CHART_NAME)" $(HELM_OPTIONS) --namespace "$(DEPLOYMENTNAMESPACE)"
else
	# Using helm v2
	HELM_PACKAGE_CMD = $(HELM) package "$(ROOTDIR)/chart/$(CHART_NAME)" -d "$(ROOTDIR)/bin/charts" \
		--save=false --version "$(VERSION_MAJOR_MINOR_PATCH)"

	HELM_CMD = $(HELM) template "$(ROOTDIR)/chart/$(CHART_NAME)" --name "$(NAME)" $(HELM_OPTIONS) \
		--namespace "$(DEPLOYMENTNAMESPACE)"
endif

ifndef LOCALONLY
	PUSHIMAGES := 1
endif

BUILD_SKIP_UPDATE ?= false

ifdef IMAGETAG
	IMAGESUFFIX := :$(IMAGETAG)
else
	IMAGESUFFIX := :dev
endif

ifeq ($(DEBUG),true)
	DEBUG := true
	DOCKERFILE := Dockerfile.debug
	# required by DLV https://github.com/go-delve/delve/blob/master/Documentation/usage/dlv_exec.md
	COMPILE_DEBUG_FLAGS := -gcflags="all=-N -l" -ldflags "-extldflags '-static'" 
else
	DEBUG := false
	DOCKERFILE := Dockerfile
	COMPILE_DEBUG_FLAGS :=
endif

PROTOC_VERSION := 21.1
ifeq ($(shell uname),Darwin)
	PROTOC_ARCHIVE_SUFFIX := osx-universal_binary
else
	PROTOC_ARCHIVE_SUFFIX := linux-x86_64
endif
PROTOC_URL := https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-${PROTOC_ARCHIVE_SUFFIX}.zip

ifeq ($(MANIFESTSUFFIX),-)
	# Release setting
	MANIFESTSUFFIX :=
else
ifndef MANIFESTSUFFIX
	MANIFESTSUFFIX := -dev
endif
endif

ifeq ($(RELEASE_MODE),community)
MANIFESTPATHCRD := manifests/arango-crd$(MANIFESTSUFFIX).yaml
MANIFESTPATHCRDBASIC := manifests/arango-crd-basic$(MANIFESTSUFFIX).yaml
MANIFESTPATHCRDALL := manifests/arango-crd-all$(MANIFESTSUFFIX).yaml
MANIFESTPATHDEPLOYMENT := manifests/arango-deployment$(MANIFESTSUFFIX).yaml
MANIFESTPATHDEPLOYMENTREPLICATION := manifests/arango-deployment-replication$(MANIFESTSUFFIX).yaml
MANIFESTPATHBACKUP := manifests/arango-backup$(MANIFESTSUFFIX).yaml
MANIFESTPATHAPPS := manifests/arango-apps$(MANIFESTSUFFIX).yaml
MANIFESTPATHML := manifests/arango-ml$(MANIFESTSUFFIX).yaml
MANIFESTPATHK2KCLUSTERSYNC := manifests/arango-k2kclustersync$(MANIFESTSUFFIX).yaml
MANIFESTPATHSTORAGE := manifests/arango-storage$(MANIFESTSUFFIX).yaml
MANIFESTPATHALL := manifests/arango-all$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHCRD := manifests/kustomize/crd/arango-crd$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHCRDBASIC := manifests/kustomize/crd/arango-crd-basic$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHCRDALL := manifests/kustomize/crd/arango-crd-all$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHDEPLOYMENT := manifests/kustomize/deployment/arango-deployment$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHDEPLOYMENTREPLICATION := manifests/kustomize/deployment-replication/arango-deployment-replication$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHBACKUP := manifests/kustomize/backup/arango-backup$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHAPPS := manifests/kustomize/apps/arango-apps$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHML := manifests/kustomize/apps/arango-ml$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHK2KCLUSTERSYNC := manifests/kustomize/k2kclustersync/arango-k2kclustersync$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHSTORAGE := manifests/kustomize/storage/arango-storage$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHALL := manifests/kustomize/all/arango-all$(MANIFESTSUFFIX).yaml
else
MANIFESTPATHCRD := manifests/enterprise-crd$(MANIFESTSUFFIX).yaml
MANIFESTPATHCRDBASIC := manifests/enterprise-crd-basic$(MANIFESTSUFFIX).yaml
MANIFESTPATHCRDALL := manifests/enterprise-crd-all$(MANIFESTSUFFIX).yaml
MANIFESTPATHDEPLOYMENT := manifests/enterprise-deployment$(MANIFESTSUFFIX).yaml
MANIFESTPATHDEPLOYMENTREPLICATION := manifests/enterprise-deployment-replication$(MANIFESTSUFFIX).yaml
MANIFESTPATHBACKUP := manifests/enterprise-backup$(MANIFESTSUFFIX).yaml
MANIFESTPATHAPPS := manifests/enterprise-apps$(MANIFESTSUFFIX).yaml
MANIFESTPATHML := manifests/enterprise-ml$(MANIFESTSUFFIX).yaml
MANIFESTPATHK2KCLUSTERSYNC := manifests/enterprise-k2kclustersync$(MANIFESTSUFFIX).yaml
MANIFESTPATHSTORAGE := manifests/enterprise-storage$(MANIFESTSUFFIX).yaml
MANIFESTPATHALL := manifests/enterprise-all$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHCRD := manifests/kustomize-enterprise/crd/enterprise-crd$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHCRDBASIC := manifests/kustomize-enterprise/crd/enterprise-crd-basic$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHCRDALL := manifests/kustomize-enterprise/crd/enterprise-crd-all$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHDEPLOYMENT := manifests/kustomize-enterprise/deployment/enterprise-deployment$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHDEPLOYMENTREPLICATION := manifests/kustomize-enterprise/deployment-replication/enterprise-deployment-replication$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHBACKUP := manifests/kustomize-enterprise/backup/enterprise-backup$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHAPPS := manifests/kustomize-enterprise/apps/enterprise-apps$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHML := manifests/kustomize/apps/enterprise-ml$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHK2KCLUSTERSYNC := manifests/kustomize-enterprise/k2kclustersync/enterprise-k2kclustersync$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHSTORAGE := manifests/kustomize-enterprise/storage/enterprise-storage$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHALL := manifests/kustomize-enterprise/all/enterprise-all$(MANIFESTSUFFIX).yaml
endif

ifndef DEPLOYMENTNAMESPACE
	DEPLOYMENTNAMESPACE := default
endif

BASEUBIIMAGE ?= registry.access.redhat.com/ubi8/ubi-minimal:8.4

OPERATORIMAGENAME ?= $(REPONAME)

ifndef OPERATORIMAGE
	OPERATORIMAGE := $(DOCKERNAMESPACE)/$(OPERATORIMAGENAME)$(IMAGESUFFIX)
endif
ifndef OPERATORUBIIMAGE
	OPERATORUBIIMAGE := $(DOCKERNAMESPACE)/$(OPERATORIMAGENAME)$(IMAGESUFFIX)-ubi
endif
ifndef ENTERPRISEIMAGE
	ENTERPRISEIMAGE := $(DEFAULTENTERPRISEIMAGE)
endif
ifndef ENTERPRISELICENSE
	ENTERPRISELICENSE := $(DEFAULTENTERPRISELICENSE)
endif
DASHBOARDBUILDIMAGE := kube-arangodb-dashboard-builder

ifndef ALLOWCHAOS
	ALLOWCHAOS := true
endif

BINNAME := $(PROJECT)
BIN := $(BINDIR)/$(BINNAME)

BIN_OPS_NAME := $(PROJECT)_ops
BIN_OPS := $(BINDIR)/$(BIN_OPS_NAME)

BIN_INT_NAME := $(PROJECT)_integration
BIN_INT := $(BINDIR)/$(BIN_INT_NAME)

BIN_PLATFORM_NAME := $(PROJECT)_platform
BIN_PLATFORM := $(BINDIR)/$(BIN_PLATFORM_NAME)

define binary
$(eval $(call binary_operator,$1,$2,$3))
$(eval $(call binary_ops,$1,$2,$3))
$(eval $(call binary_int,$1,$2,$3))
$(eval $(call binary_platform,$1,$2,$3))
endef

define binary_operator
$(eval _OS:=$(call UPPER_ENV,$1))
$(eval _ARCH:=$(call UPPER_ENV,$2))

VBIN_OPERATOR_$(_OS)_$(_ARCH) := $(BINDIR)/$(RELEASE_MODE)/$1/$2/$(BINNAME)$3

.PHONY: $$(VBIN_OPERATOR_$(_OS)_$(_ARCH))

$$(VBIN_OPERATOR_$(_OS)_$(_ARCH)): $$(SOURCES) dashboard/assets.go VERSION
	@mkdir -p $(BINDIR)/$(RELEASE_MODE)/$1/$2
	CGO_ENABLED=0 GOOS=$1 GOARCH=$2 go build $${GOBUILDARGS} --tags "$$(GOBUILDTAGS)" $$(COMPILE_DEBUG_FLAGS) -installsuffix netgo -gcflags=all="$$(GOBUILDGCFLAGS)" -ldflags "$$(GOBUILDLDFLAGS)" -o $$@ ./cmd/main

bin-all: $$(VBIN_OPERATOR_$(_OS)_$(_ARCH))
endef

define binary_int
$(eval _OS:=$(call UPPER_ENV,$1))
$(eval _ARCH:=$(call UPPER_ENV,$2))

VBIN_INT_$(_OS)_$(_ARCH) := $(BINDIR)/$(RELEASE_MODE)/$1/$2/$(BIN_INT_NAME)$3

.PHONY: $$(VBIN_INT_$(_OS)_$(_ARCH))

$$(VBIN_INT_$(_OS)_$(_ARCH)): $$(SOURCES) dashboard/assets.go VERSION
	@mkdir -p $(BINDIR)/$(RELEASE_MODE)/$1/$2
	CGO_ENABLED=0 GOOS=$1 GOARCH=$2 go build $${GOBUILDARGS} --tags "$$(GOBUILDTAGS)" $$(COMPILE_DEBUG_FLAGS) -installsuffix netgo -gcflags=all="$$(GOBUILDGCFLAGS)" -ldflags "$$(GOBUILDLDFLAGS)" -o $$@ ./cmd/main-int

bin-all: $$(VBIN_INT_$(_OS)_$(_ARCH))
endef

define binary_ops
$(eval _OS:=$(call UPPER_ENV,$1))
$(eval _ARCH:=$(call UPPER_ENV,$2))

VBIN_OPS_$(_OS)_$(_ARCH) := $(BINDIR)/$(RELEASE_MODE)/$1/$2/$(BIN_OPS_NAME)$3

.PHONY: $$(VBIN_OPS_$(_OS)_$(_ARCH))

$$(VBIN_OPS_$(_OS)_$(_ARCH)): $$(SOURCES) dashboard/assets.go VERSION
	@mkdir -p $(BINDIR)/$(RELEASE_MODE)/$1/$2
	CGO_ENABLED=0 GOOS=$1 GOARCH=$2 go build $${GOBUILDARGS} --tags "$$(GOBUILDTAGS)" $$(COMPILE_DEBUG_FLAGS) -installsuffix netgo -gcflags=all="$$(GOBUILDGCFLAGS)" -ldflags "$$(GOBUILDLDFLAGS)" -o $$@ ./cmd/main-ops

bin-all: $$(VBIN_OPS_$(_OS)_$(_ARCH))
endef

define binary_platform
$(eval _OS:=$(call UPPER_ENV,$1))
$(eval _ARCH:=$(call UPPER_ENV,$2))

VBIN_PLATFORM_$(_OS)_$(_ARCH) := $(BINDIR)/$(RELEASE_MODE)/$1/$2/$(BIN_PLATFORM_NAME)$3

.PHONY: $$(VBIN_PLATFORM_$(_OS)_$(_ARCH))

$$(VBIN_PLATFORM_$(_OS)_$(_ARCH)): $$(SOURCES) dashboard/assets.go VERSION
	@mkdir -p $(BINDIR)/$(RELEASE_MODE)/$1/$2
	CGO_ENABLED=0 GOOS=$1 GOARCH=$2 go build $${GOBUILDARGS} --tags "$$(GOBUILDTAGS)" $$(COMPILE_DEBUG_FLAGS) -installsuffix netgo -gcflags=all="$$(GOBUILDGCFLAGS)" -ldflags "$$(GOBUILDLDFLAGS)" -o $$@ ./cmd/main-platform

bin-all: $$(VBIN_PLATFORM_$(_OS)_$(_ARCH))
endef

$(eval $(call binary,linux,amd64))
$(eval $(call binary,linux,arm64))
$(eval $(call binary,darwin,amd64))
$(eval $(call binary,darwin,arm64))
$(eval $(call binary_platform,windows,amd64,.exe))
$(eval $(call binary_platform,windows,arm64,.exe))

ifdef VERBOSE
	TESTVERBOSEOPTIONS := -v
endif

EXCLUDE_DIRS := vendor .gobuild deps tools pkg/generated/clientset pkg/generated/informers pkg/generated/listers
EXCLUDE_FILES := *generated.deepcopy.go
SOURCES_QUERY := find ./ -type f -name '*.go' ! -name '*.pb.gw.go' ! -name '*.pb.go' $(foreach EXCLUDE_DIR,$(EXCLUDE_DIRS), ! -path "*/$(EXCLUDE_DIR)/*") $(foreach EXCLUDE_FILE,$(EXCLUDE_FILES), ! -path "*/$(EXCLUDE_FILE)")
SOURCES := $(shell $(SOURCES_QUERY))

NON_EE_SOURCES_QUERY := $(SOURCES_QUERY) ! -name '*.enterprise.go'
NON_EE_SOURCES := $(shell $(NON_EE_SOURCES_QUERY))

YAML_EXCLUDE_DIRS := vendor .gobuild deps tools pkg/generated/clientset pkg/generated/informers pkg/generated/listers \
                     chart/kube-arangodb/templates chart/kube-arangodb-arm64/templates chart/kube-arangodb-enterprise/templates chart/kube-arangodb-enterprise-arm64/templates  \
                     chart/kube-arangodb-crd/templates
YAML_EXCLUDE_FILES :=
YAML_QUERY := find ./ -type f -name '*.yaml' $(foreach EXCLUDE_DIR,$(YAML_EXCLUDE_DIRS), ! -path "*/$(EXCLUDE_DIR)/*") $(foreach EXCLUDE_FILE,$(YAML_EXCLUDE_FILES), ! -path "*/$(EXCLUDE_FILE)")
YAMLS := $(shell $(YAML_QUERY))


DOCS_EXCLUDE_DIRS := 
DOCS_EXCLUDE_FILES :=
DOCS_QUERY := find ./docs/ -type f -name '*.md' $(foreach EXCLUDE_DIR,$(DOCS_EXCLUDE_DIRS), ! -path "*/$(EXCLUDE_DIR)/*") $(foreach EXCLUDE_FILE,$(DOCS_EXCLUDE_FILES), ! -path "*/$(EXCLUDE_FILE)")
DOCS := $(shell $(DOCS_QUERY))

DASHBOARDSOURCES := $(shell find $(DASHBOARDDIR)/src -name '*.js') $(DASHBOARDDIR)/package.json
LINT_EXCLUDES:=
ifeq ($(RELEASE_MODE),enterprise)
LINT_EXCLUDES+=.*\.community\.go$$
else
LINT_EXCLUDES+=.*\.enterprise\.go$$
endif

PROTOSOURCES := $(shell find ./ -type f  -name '*.proto' $(foreach EXCLUDE_DIR,$(EXCLUDE_DIRS), ! -path "*/$(EXCLUDE_DIR)/*") | sort)

.DEFAULT_GOAL := all
.PHONY: all
all: check-vars build

.PHONY: compile
compile: check-vars build

# allall  is now obsolete
.PHONY: allall
allall: all

#
# Tip: Run `eval $(minikube docker-env)` before calling make if you're developing on minikube.
#

.PHONY: license-verify
license-verify:
	@echo ">> Verify license of files"
	@$(GOPATH)/bin/addlicense -f "./tools/codegen/license-header.txt" -check $(NON_EE_SOURCES) $(PROTOSOURCES)

.PHONY: license-range-verify
license-range-verify:
	@GOBIN=$(GOPATH)/bin go run "$(ROOT)/tools/license/" $(SOURCES)

.PHONY: license-range
license-range:
	@GOBIN=$(GOPATH)/bin go run "$(ROOT)/tools/license/" -w $(SOURCES)

.PHONY: fmt
fmt:
	@echo ">> Ensuring style of files"
	@$(GOPATH)/bin/goimports -w $(SOURCES)
	@$(GOPATH)/bin/gci write -s "standard" -s "default" -s "prefix(github.com/arangodb)" -s "prefix(github.com/arangodb/kube-arangodb)" $(SOURCES)

.PHONY: yamlfmt
yamlfmt:
	@echo ">> Ensuring style of yaml files"
	@$(GOPATH)/bin/yamlfmt -quiet $(YAMLS)

.PHONY: yamlfmt-verify
yamlfmt-verify:
	@echo ">> Verifying style of yaml files"
	@$(GOPATH)/bin/yamlfmt -lint -quiet $(YAMLS)

.PHONY: license
license:
	@echo ">> Ensuring license of files"
	@$(GOPATH)/bin/addlicense -f "./tools/codegen/license-header.txt" $(NON_EE_SOURCES) $(PROTOSOURCES)

.PHONY: fmt-verify
fmt-verify: license-verify
	@echo ">> Verify files style"
	@if $(GOPATH)/bin/goimports -l $(SOURCES) | grep -v "^$$"; then echo ">> Style errors"; exit 1; fi

.PHONY: linter
linter:
	@$(GOPATH)/bin/golangci-lint run --build-tags "testing,$(GOBUILDTAGS)" $(foreach LINT_EXCLUDE,$(LINT_EXCLUDES),--exclude '$(LINT_EXCLUDE)') ./...

.PHONY: linter-fix
linter-fix:
	@$(GOPATH)/bin/golangci-lint run --fix --build-tags "testing,$(GOBUILDTAGS)" $(foreach LINT_EXCLUDE,$(LINT_EXCLUDES),--exclude '$(LINT_EXCLUDE)') ./...

.PHONY: protolint protolint-fix

protolint:
	@$(GOPATH)/bin/protolint lint $(PROTOSOURCES)

protolint-fix:
	@$(GOPATH)/bin/protolint lint --fix $(PROTOSOURCES)

.PHONY: vulncheck vulncheck-optional
vulncheck:
	@echo ">> Checking for known vulnerabilities (required)"
	@$(GOPATH)/bin/govulncheck --tags $(GOBUILDTAGS) ./...

vulncheck-optional:
	@echo ">> Checking for known vulnerabilities (optional)"
	@-$(GOPATH)/bin/govulncheck --tags $(GOBUILDTAGS) ./...

.PHONY: build
build: docker manifests

ifndef IGNORE_UBI
build: docker-ubi
endif

.PHONY: clean
clean:
	rm -Rf $(BIN) $(BINDIR) $(DASHBOARDDIR)/build $(DASHBOARDDIR)/node_modules $(VBIN_OPERATOR_LINUX_AMD64) $(VBIN_OPERATOR_LINUX_ARM64) $(VBIN_OPS_LINUX_AMD64) $(VBIN_OPS_LINUX_ARM64) $(VBIN_OPS_DARWIN_AMD64) $(VBIN_OPS_DARWIN_ARM64) $(VBIN_OPS_WIN_AMD64)

.PHONY: check-vars
check-vars:
ifndef DOCKERNAMESPACE
	@echo "DOCKERNAMESPACE must be set"
	@exit 1
endif
	@echo "Using docker namespace: $(DOCKERNAMESPACE)"

KUBERNETES_APIS := k8s.io/api \
					k8s.io/apiextensions-apiserver \
					k8s.io/apimachinery \
					k8s.io/apiserver \
					k8s.io/client-go \
					k8s.io/cloud-provider \
					k8s.io/cluster-bootstrap \
					k8s.io/code-generator \
					k8s.io/component-base \
					k8s.io/kubernetes \
					k8s.io/metrics

KUBERNETES_MODS := k8s.io/api \
					k8s.io/apiextensions-apiserver \
					k8s.io/apimachinery \
					k8s.io/apiserver \
					k8s.io/client-go \
					k8s.io/cloud-provider \
					k8s.io/cluster-bootstrap \
					k8s.io/code-generator \
					k8s.io/component-base \
					k8s.io/metrics

.PHONY: update-kubernetes-version
update-kubernetes-version:
	@$(foreach API,$(KUBERNETES_APIS), sed -i 's#$(API) => $(API) .*#$(API) => $(API) v0.$(KUBERNETES_VERSION_MINOR).$(KUBERNETES_VERSION_PATCH)#g' '$(ROOT)/go.mod' &&) echo "Replaced to K8S 1.$(KUBERNETES_VERSION_MINOR).$(KUBERNETES_VERSION_PATCH)"

update-kubernetes-version-go:
	@$(foreach API,$(KUBERNETES_MODS), go get '$(API)@v0.$(KUBERNETES_VERSION_MINOR).$(KUBERNETES_VERSION_PATCH)' &&) echo "Go Upgraded to K8S 1.$(KUBERNETES_VERSION_MINOR).$(KUBERNETES_VERSION_PATCH)"

.PHONY: update-vendor
update-vendor:
	@rm -Rf $(VENDORDIR)/k8s.io/code-generator
	@git clone --branch "kubernetes-1.$(KUBERNETES_VERSION_MINOR).$(KUBERNETES_VERSION_PATCH)" https://github.com/kubernetes/code-generator.git $(VENDORDIR)/k8s.io/code-generator
	@rm -Rf $(VENDORDIR)/k8s.io/code-generator/.git
	@(cd "$(VENDORDIR)/k8s.io/code-generator"; go mod download; go mod vendor)


.PHONY: update-generated
update-generated:
	@$(SED) -e 's/^/\/\/ /' -e 's/ *$$//' $(ROOTDIR)/tools/codegen/license-header.txt > $(ROOTDIR)/tools/codegen/boilerplate.go.txt
	bash "${ROOTDIR}/scripts/codegen.sh" "${ROOTDIR}"

dashboard/assets.go:
	cd $(DASHBOARDDIR) && docker build -t $(DASHBOARDBUILDIMAGE) -f Dockerfile.build $(DASHBOARDDIR)
	@mkdir -p $(DASHBOARDDIR)/build
	docker run --rm \
		-u $(shell id -u):$(shell id -g) \
		-v $(DASHBOARDDIR)/build:/usr/code/build \
		-v $(DASHBOARDDIR)/public:/usr/code/public:ro \
		-v $(DASHBOARDDIR)/src:/usr/code/src:ro \
		$(DASHBOARDBUILDIMAGE)
	$(GOASSETSBUILDER) -s /dashboard/build/ -o dashboard/assets.go -p dashboard dashboard/build

# Binaries

.PHONY: bin
bin: $(BIN)

$(BIN): $(VBIN_OPERATOR_LINUX_AMD64) $(VBIN_OPS_LINUX_AMD64) $(VBIN_INT_LINUX_AMD64) $(VBIN_PLATFORM_LINUX_AMD64)
	@cp "$(VBIN_OPERATOR_LINUX_AMD64)" "$(BIN)"
	@cp "$(VBIN_OPS_LINUX_AMD64)" "$(BIN_OPS)"

.PHONY: docker
docker: clean check-vars $(VBIN_OPERATOR_LINUX_AMD64) $(VBIN_OPERATOR_LINUX_ARM64)
ifdef PUSHIMAGES
	docker buildx build --no-cache -f $(DOCKERFILE) --build-arg "ENVOY_IMAGE=$(ENVOY_IMAGE)" --build-arg GOVERSION=$(GOVERSION) --build-arg DISTRIBUTION=$(DISTRIBUTION) \
		--build-arg "VERSION=${VERSION_MAJOR_MINOR_PATCH}" --build-arg "RELEASE_MODE=$(RELEASE_MODE)" --build-arg "BUILD_SKIP_UPDATE=${BUILD_SKIP_UPDATE}" \
		--platform linux/amd64,linux/arm64 --push -t $(OPERATORIMAGE) .
else
	docker buildx build --no-cache -f $(DOCKERFILE) --build-arg "ENVOY_IMAGE=$(ENVOY_IMAGE)" --build-arg GOVERSION=$(GOVERSION) --build-arg DISTRIBUTION=$(DISTRIBUTION) \
		--build-arg "VERSION=${VERSION_MAJOR_MINOR_PATCH}" --build-arg "RELEASE_MODE=$(RELEASE_MODE)" --build-arg "BUILD_SKIP_UPDATE=${BUILD_SKIP_UPDATE}" \
		--platform linux/amd64,linux/arm64 -t $(OPERATORIMAGE) .
endif

.PHONY: docker-ubi
docker-ubi: check-vars $(VBIN_OPERATOR_LINUX_AMD64)
ifdef PUSHIMAGES
	docker buildx build --no-cache -f "$(DOCKERFILE).ubi" --build-arg "ENVOY_IMAGE=$(ENVOY_IMAGE)" --build-arg GOVERSION=$(GOVERSION) --build-arg DISTRIBUTION=$(DISTRIBUTION) \
		--build-arg "VERSION=${VERSION_MAJOR_MINOR_PATCH}" --build-arg "RELEASE_MODE=$(RELEASE_MODE)" \
		--build-arg "IMAGE=$(BASEUBIIMAGE)" \
		--platform linux/amd64 --push -t $(OPERATORUBIIMAGE) .
else
	docker buildx build --no-cache -f "$(DOCKERFILE).ubi" --build-arg "ENVOY_IMAGE=$(ENVOY_IMAGE)" --build-arg GOVERSION=$(GOVERSION) --build-arg DISTRIBUTION=$(DISTRIBUTION) \
		--build-arg "VERSION=${VERSION_MAJOR_MINOR_PATCH}" --build-arg "RELEASE_MODE=$(RELEASE_MODE)" \
		--build-arg "IMAGE=$(BASEUBIIMAGE)" \
		--platform linux/amd64 -t $(OPERATORUBIIMAGE) .
endif

# Manifests

define manifest-generator
$(eval _TARGET:=$(call LOWER,$1))
$(eval _ENV:=$(call UPPER_ENV,$1))
.PHONY: manifests-$(_TARGET)-file
manifests-$(_TARGET)-file: export CHART_NAME := $2
manifests-$(_TARGET)-file: export NAME := $(_TARGET)
manifests-$(_TARGET)-file: helm
	@echo Building manifests for $(_ENV) - $$(MANIFESTPATH$(_ENV))
	@$$(HELM_CMD) $3 > "$$(MANIFESTPATH$(_ENV))"
manifests: manifests-$(_TARGET)-file

.PHONY: manifests-$(_TARGET)-kustomize
manifests-$(_TARGET)-kustomize: export CHART_NAME := $2
manifests-$(_TARGET)-kustomize: export NAME := $(_TARGET)
manifests-$(_TARGET)-kustomize: helm
	@echo Building kustomize manifests for $(_ENV) - $$(KUSTOMIZEPATH$(_ENV))
	@$$(HELM_CMD) $3 > "$$(KUSTOMIZEPATH$(_ENV))"
manifests: manifests-$(_TARGET)-kustomize
endef

.PHONY: manifests
manifests:

.PHONY: manifests-crd-file
manifests-crd-file:
	@echo Building manifests for CRD - $(MANIFESTPATHCRD)
	@go run ${GOBUILDARGS} --tags "$(GOBUILDTAGS)" '$(ROOT)/cmd/main-ops/' crd generate --crd.validation-schema 'all=false' --crd.skip arangolocalstorages.storage.arangodb.com > $(MANIFESTPATHCRD)
manifests-crd: manifests-crd-file

.PHONY: manifests-crd-all-file
manifests-crd-all-file:
	@echo Building manifests for CRD with schemas - $(MANIFESTPATHCRDALL)
	@go run ${GOBUILDARGS} --tags "$(GOBUILDTAGS)" '$(ROOT)/cmd/main-ops/' crd generate --crd.validation-schema 'all=true' > $(MANIFESTPATHCRDALL)
manifests-crd: manifests-crd-all-file

.PHONY: manifests-crd-basic-file
manifests-crd-basic-file:
	@echo Building manifests for CRD with basic schemas - $(MANIFESTPATHCRDBASIC)
	@go run ${GOBUILDARGS} --tags "$(GOBUILDTAGS)" '$(ROOT)/cmd/main-ops/' crd generate > $(MANIFESTPATHCRDBASIC)
manifests-crd: manifests-crd-basic-file

.PHONY: manifests-crd-kustomize
manifests-crd-kustomize: manifests-crd-file
	@echo Building manifests for CRD - $(KUSTOMIZEPATHCRD)
	@cp "$(MANIFESTPATHCRD)" "$(KUSTOMIZEPATHCRD)"
manifests-crd: manifests-crd-kustomize

.PHONY: manifests-crd-basic-kustomize
manifests-crd-basic-kustomize: manifests-crd-basic-file
	@echo Building manifests for CRD with schemas - $(KUSTOMIZEPATHCRDBASIC)
	@cp "$(MANIFESTPATHCRDBASIC)" "$(KUSTOMIZEPATHCRDBASIC)"
manifests-crd: manifests-crd-basic-kustomize

.PHONY: manifests-crd-all-kustomize
manifests-crd-all-kustomize: manifests-crd-all-file
	@echo Building manifests for CRD with schemas - $(KUSTOMIZEPATHCRDALL)
	@cp "$(MANIFESTPATHCRDALL)" "$(KUSTOMIZEPATHCRDALL)"
manifests-crd: manifests-crd-all-kustomize

manifests: manifests-crd

$(eval $(call manifest-generator, deployment, kube-arangodb, \
		--set "operator.features.deployment=true" \
		--set "operator.features.deploymentReplications=false" \
		--set "operator.features.storage=false" \
		--set "operator.features.backup=false" \
		--set "operator.features.apps=false" \
		--set "operator.features.k8sToK8sClusterSync=false" \
		--set "operator.features.ml=false" \
		--set "operator.features.analytics=false" \
		--set "operator.features.networking=true" \
		--set "operator.features.scheduler=true" \
		--set "operator.features.platform=true"))

$(eval $(call manifest-generator, deployment-replication, kube-arangodb, \
		--set "operator.features.deployment=false" \
		--set "operator.features.deploymentReplications=true" \
		--set "operator.features.storage=false" \
		--set "operator.features.backup=false" \
		--set "operator.features.apps=false" \
		--set "operator.features.k8sToK8sClusterSync=false" \
		--set "operator.features.ml=false" \
		--set "operator.features.analytics=false" \
		--set "operator.features.networking=false" \
		--set "operator.features.scheduler=false" \
		--set "operator.features.platform=false"))

$(eval $(call manifest-generator, storage, kube-arangodb, \
		--set "operator.features.deployment=false" \
		--set "operator.features.deploymentReplications=false" \
		--set "operator.features.storage=true" \
		--set "operator.features.backup=false" \
		--set "operator.features.apps=false" \
		--set "operator.features.k8sToK8sClusterSync=false" \
		--set "operator.features.ml=false" \
		--set "operator.features.analytics=false" \
		--set "operator.features.networking=false" \
		--set "operator.features.scheduler=false" \
 		--set "operator.features.platform=false"))

$(eval $(call manifest-generator, backup, kube-arangodb, \
		--set "operator.features.deployment=false" \
		--set "operator.features.deploymentReplications=false" \
		--set "operator.features.storage=false" \
		--set "operator.features.backup=true" \
		--set "operator.features.apps=false" \
		--set "operator.features.k8sToK8sClusterSync=false" \
		--set "operator.features.ml=false" \
		--set "operator.features.analytics=false" \
		--set "operator.features.networking=false" \
		--set "operator.features.scheduler=false" \
 		--set "operator.features.platform=false"))

$(eval $(call manifest-generator, apps, kube-arangodb, \
		--set "operator.features.deployment=false" \
		--set "operator.features.deploymentReplications=false" \
		--set "operator.features.storage=false" \
		--set "operator.features.backup=false" \
		--set "operator.features.apps=true" \
		--set "operator.features.k8sToK8sClusterSync=false" \
		--set "operator.features.ml=false" \
		--set "operator.features.analytics=false" \
		--set "operator.features.networking=false" \
		--set "operator.features.scheduler=false" \
 		--set "operator.features.platform=false"))

$(eval $(call manifest-generator, ml, kube-arangodb, \
		--set "operator.features.deployment=false" \
		--set "operator.features.deploymentReplications=false" \
		--set "operator.features.storage=false" \
		--set "operator.features.backup=false" \
		--set "operator.features.apps=false" \
		--set "operator.features.k8sToK8sClusterSync=false" \
		--set "operator.features.ml=true" \
		--set "operator.features.analytics=false" \
		--set "operator.features.networking=false" \
		--set "operator.features.scheduler=false" \
 		--set "operator.features.platform=false"))

$(eval $(call manifest-generator, k2kclustersync, kube-arangodb, \
		--set "operator.features.deployment=false" \
		--set "operator.features.deploymentReplications=false" \
		--set "operator.features.storage=false" \
		--set "operator.features.backup=false" \
		--set "operator.features.apps=false" \
		--set "operator.features.k8sToK8sClusterSync=true" \
		--set "operator.features.ml=false" \
		--set "operator.features.analytics=false" \
		--set "operator.features.networking=false" \
		--set "operator.features.scheduler=false" \
 		--set "operator.features.platform=false"))

$(eval $(call manifest-generator, all, kube-arangodb, \
		--set "operator.features.deployment=true" \
		--set "operator.features.deploymentReplications=true" \
		--set "operator.features.storage=true" \
		--set "operator.features.backup=true" \
		--set "operator.features.apps=true" \
		--set "operator.features.k8sToK8sClusterSync=true" \
		--set "operator.features.ml=true" \
		--set "operator.features.analytics=true" \
		--set "operator.features.networking=true" \
		--set "operator.features.scheduler=true" \
		--set "operator.features.platform=true"))

.PHONY: chart-crd
chart-crd: export CHART_NAME := kube-arangodb-crd
chart-crd: helm
	@mkdir -p "$(ROOTDIR)/bin/charts"
	@$(HELM_PACKAGE_CMD)
manifests: chart-crd

.PHONY: chart-operator
chart-operator: export CHART_NAME := kube-arangodb
chart-operator: helm
	@mkdir -p "$(ROOTDIR)/bin/charts"
	@$(HELM_PACKAGE_CMD)

manifests: chart-operator

.PHONY: chart-operator-enterprise
chart-operator-enterprise: export CHART_NAME := kube-arangodb-enterprise
chart-operator-enterprise: helm
	@mkdir -p "$(ROOTDIR)/bin/charts"
	@$(HELM_PACKAGE_CMD)

manifests: chart-operator-enterprise

.PHONY: chart-operator-arm64
chart-operator-arm64: export CHART_NAME := kube-arangodb-arm64
chart-operator-arm64: helm
	@mkdir -p "$(ROOTDIR)/bin/charts"
	@$(HELM_PACKAGE_CMD)

manifests: chart-operator-arm64

.PHONY: chart-operator-enterprise-arm64
chart-operator-enterprise-arm64: export CHART_NAME := kube-arangodb-enterprise-arm64
chart-operator-enterprise-arm64: helm
	@mkdir -p "$(ROOTDIR)/bin/charts"
	@$(HELM_PACKAGE_CMD)

manifests: chart-operator-enterprise-arm64

.PHONY: manifests-verify
manifests-verify:
	$(MAKE) manifest-verify-plain-ce
	$(MAKE) manifest-verify-plain-ee
	$(MAKE) manifest-verify-kustomize-ce
	$(MAKE) manifest-verify-kustomize-ee
	$(MAKE) manifest-verify-helm-ce
	$(MAKE) manifest-verify-helm-ee

manifests-verify-env-reset:
	@minikube delete && minikube start

manifest-verify-plain-ce: manifests-verify-env-reset
	@echo "Trying to install via plain manifests"
	kubectl apply -f ./manifests/arango-all.yaml
	kubectl apply -f ./manifests/arango-apps.yaml
	kubectl apply -f ./manifests/arango-backup.yaml
	kubectl apply -f ./manifests/arango-crd.yaml
	kubectl apply -f ./manifests/arango-deployment.yaml
	kubectl apply -f ./manifests/arango-deployment-replication.yaml
	kubectl apply -f ./manifests/arango-k2kclustersync.yaml
	kubectl apply -f ./manifests/arango-storage.yaml

manifest-verify-plain-ee: manifests-verify-env-reset
	kubectl apply -f ./manifests/enterprise-all.yaml
	kubectl apply -f ./manifests/enterprise-apps.yaml
	kubectl apply -f ./manifests/enterprise-backup.yaml
	kubectl apply -f ./manifests/enterprise-crd.yaml
	kubectl apply -f ./manifests/enterprise-deployment.yaml
	kubectl apply -f ./manifests/enterprise-deployment-replication.yaml
	kubectl apply -f ./manifests/enterprise-k2kclustersync.yaml
	kubectl apply -f ./manifests/enterprise-storage.yaml

define KUSTOMIZE_YAML =
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ./all
  - ./apps
  - ./backup
  - ./crd
  - ./deployment
  - ./deployment-replication
  - ./k2kclustersync
endef
export KUSTOMIZE_YAML

manifest-verify-kustomize-ce: manifests-verify-env-reset
	@echo "Trying to install via Kustomize"
	@-rm -rf ./kustomize_test
	@cp -r ./manifests/kustomize ./kustomize_test
	@echo "$$KUSTOMIZE_YAML" > ./kustomize_test/kustomization.yaml
	@kubectl create -k ./kustomize_test/

manifest-verify-kustomize-ee: manifests-verify-env-reset
	@echo "Trying to install via Kustomize"
	@-rm -rf ./kustomize_test
	@cp -r ./manifests/kustomize-enterprise ./kustomize_test
	@echo "$$KUSTOMIZE_YAML" > ./kustomize_test/kustomization.yaml
	@kubectl create -k ./kustomize_test/

manifest-verify-helm-ce: manifests-verify-env-reset
	@echo "Trying to install via Helm charts"
	helm install --generate-name --set "operator.features.storage=true" \
		./bin/charts/kube-arangodb-$(VERSION_MAJOR_MINOR_PATCH).tgz

manifest-verify-helm-ee: manifests-verify-env-reset
	@echo "Trying to install via Helm charts"
	helm install --generate-name --set "operator.image=arangodb/kube-arangodb-enterprise:$(VERSION_MAJOR_MINOR_PATCH)" --set "operator.features.storage=true" \
		./bin/charts/kube-arangodb-$(VERSION_MAJOR_MINOR_PATCH).tgz

# K3D

.PHONY: _k3d_cluster_start _k3d_cluster_stop _k3d_cluster_config _k3d_cluster

_k3d_cluster_config:
	@$(K3D) kubeconfig get "$(K3D_CLUSTER)" > "$(K3D_KUBECONFIG)"

_k3d_cluster_start:
	@$(K3D) cluster delete "$(K3D_CLUSTER)"
	@$(K3D) cluster create "$(K3D_CLUSTER)"

_k3d_cluster_stop:
	@$(K3D) cluster delete "$(K3D_CLUSTER)"

# Testing

.PHONY: run-unit-tests
ifdef K3D_ENABLED
.PHONY: _run-unit-tests

run-unit-tests: _k3d_cluster_start _k3d_cluster_config _run-unit-tests _k3d_cluster_stop

_run-unit-tests: $(SOURCES)
_run-unit-tests: export TEST_KUBECONFIG=$(K3D_KUBECONFIG)
_run-unit-tests:
	@echo "Running in the internal test scope -> $(TEST_KUBECONFIG)"
else
run-unit-tests: $(SOURCES)
endif
	go test --count=1 --tags "testing,$(GOBUILDTAGS)" $(TESTVERBOSEOPTIONS) \
		$(REPOPATH)/pkg/... \
		$(REPOPATH)/cmd/... \
		$(REPOPATH)/integrations/...

# Release building

.PHONY: patch-readme
patch-readme:
	@$(ROOTDIR)/scripts/patch_readme.sh $(VERSION_MAJOR_MINOR_PATCH)

.PHONY: patch-examples
patch-examples:
	@$(ROOTDIR)/scripts/patch_examples.sh $(VERSION_MAJOR_MINOR_PATCH)

.PHONY: patch-docs
patch-docs:
	@$(ROOTDIR)/scripts/patch_docs.sh $(VERSION_MAJOR_MINOR_PATCH) $(DOCS)

.PHONY: patch-release
patch-release: patch-readme patch-examples patch-chart

.PHONY: patch-chart
patch-chart:
	@$(ROOTDIR)/scripts/patch_chart.sh $(VERSION_MAJOR_MINOR_PATCH)

.PHONY: patch
patch: patch-chart patch-release patch-examples patch-readme

.PHONY: tidy
tidy:
	@go mod tidy -v -compat=$(GOCOMPAT)

.PHONY: deps-reload
deps-reload: tidy init

.PHONY: init
init: vendor tools update-generated $(BIN)

.PHONY: tools-min
tools-min: update-vendor
	@echo ">> Fetching golangci-lint linter"
	@GOBIN=$(GOPATH)/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
	@echo ">> Fetching goimports"
	@GOBIN=$(GOPATH)/bin go install golang.org/x/tools/cmd/goimports@v0.19.0
	@echo ">> Fetching license check"
	@GOBIN=$(GOPATH)/bin go install github.com/google/addlicense@v1.1.1
	@echo ">> Fetching yamlfmt"
	@GOBIN=$(GOPATH)/bin go install github.com/google/yamlfmt/cmd/yamlfmt@v0.10.0
	@echo ">> Fetching protolinter"
	@GOBIN=$(GOPATH)/bin go install github.com/yoheimuta/protolint/cmd/protolint@v0.47.5

.PHONY: tools
tools: tools-min
	@echo ">> Fetching gci"
	@GOBIN=$(GOPATH)/bin go install github.com/daixiang0/gci@v0.13.4
	@echo ">> Fetching yamlfmt"
	@GOBIN=$(GOPATH)/bin go install github.com/google/yamlfmt/cmd/yamlfmt@v0.10.0
	@echo ">> Downloading protobuf compiler..."
	@curl -L ${PROTOC_URL} -o $(GOPATH)/protoc.zip
	@echo ">> Unzipping protobuf compiler..."
	@unzip -o $(GOPATH)/protoc.zip -d $(GOPATH)/
	@chmod +x $(GOPATH)/bin/protoc
	@echo ">> Download proto deps"
	@rm -Rf $(GOPATH)/include/googleapis
	@git clone --branch "master" --depth 1 https://github.com/googleapis/googleapis.git $(GOPATH)/include/googleapis
	@rm -Rf $(VENDORDIR)/include/googleapis/.git
	@echo ">> Fetching protoc go plugins..."
	@GOBIN=$(GOPATH)/bin go install github.com/golang/protobuf/protoc-gen-go@v1.5.2
	@GOBIN=$(GOPATH)/bin go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	@GOBIN=$(GOPATH)/bin go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.23.0
	@echo ">> Fetching govulncheck"
	@GOBIN=$(GOPATH)/bin go install golang.org/x/vuln/cmd/govulncheck@v1.1.4

.PHONY: vendor
vendor:
	@echo ">> Updating vendor"
	@go mod vendor -e

set-deployment-api-version-v2alpha1: export API_VERSION=2alpha1
set-deployment-api-version-v2alpha1: set-api-version/deployment set-api-version/replication

set-deployment-api-version-v1: export API_VERSION=1
set-deployment-api-version-v1: set-api-version/deployment set-api-version/replication

set-typed-api-version/%:
	@grep -rHn "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/$*/v[A-Za-z0-9]\+" \
	      "$(ROOT)/pkg/deployment/" \
	      "$(ROOT)/pkg/replication/" \
	      "$(ROOT)/pkg/integrations/" \
	      "$(ROOT)/pkg/operator/" \
	      "$(ROOT)/pkg/operatorV2/" \
	      "$(ROOT)/pkg/server/" \
	      "$(ROOT)/pkg/util/" \
	      "$(ROOT)/pkg/handlers/" \
	      "$(ROOT)/pkg/apis/backup/" \
	      "$(ROOT)/pkg/apis/networking/" \
	      "$(ROOT)/pkg/apis/scheduler/" \
	      "$(ROOT)/pkg/upgrade/" \
	      "$(ROOT)/integrations/" \
	  | cut -d ':' -f 1 | sort | uniq \
	  | xargs -n 1 $(SED) -i "s#github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/$*/v[A-Za-z0-9]\+#github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/$*/v$(API_VERSION)#g"


set-api-version/%:
	@grep -rHn "github.com/arangodb/kube-arangodb/pkg/apis/$*/v[A-Za-z0-9]\+" \
	      "$(ROOT)/pkg/deployment/" \
	      "$(ROOT)/pkg/replication/" \
	      "$(ROOT)/pkg/integrations/" \
	      "$(ROOT)/pkg/operator/" \
	      "$(ROOT)/pkg/operatorV2/" \
	      "$(ROOT)/pkg/server/" \
	      "$(ROOT)/pkg/util/" \
	      "$(ROOT)/pkg/handlers/" \
	      "$(ROOT)/pkg/apis/backup/" \
	      "$(ROOT)/pkg/apis/networking/" \
	      "$(ROOT)/pkg/apis/scheduler/" \
	      "$(ROOT)/pkg/apis/platform/" \
	      "$(ROOT)/pkg/upgrade/" \
	      "$(ROOT)/integrations/" \
	  | cut -d ':' -f 1 | sort | uniq \
	  | xargs -n 1 $(SED) -i "s#github.com/arangodb/kube-arangodb/pkg/apis/$*/v[A-Za-z0-9]\+#github.com/arangodb/kube-arangodb/pkg/apis/$*/v$(API_VERSION)#g"
	@grep -rHn "DatabaseV[A-Za-z0-9]\+()" \
		  "$(ROOT)/pkg/deployment/" \
	      "$(ROOT)/pkg/replication/" \
	      "$(ROOT)/pkg/integrations/" \
	      "$(ROOT)/pkg/operator/" \
	      "$(ROOT)/pkg/operatorV2/" \
	      "$(ROOT)/pkg/server/" \
		  "$(ROOT)/pkg/util/" \
	      "$(ROOT)/pkg/handlers/" \
	      "$(ROOT)/pkg/apis/backup/" \
	      "$(ROOT)/pkg/apis/networking/" \
	      "$(ROOT)/pkg/apis/scheduler/" \
	      "$(ROOT)/pkg/apis/platform/" \
	      "$(ROOT)/pkg/upgrade/" \
	      "$(ROOT)/integrations/" \
	  | cut -d ':' -f 1 | sort | uniq \
	  | xargs -n 1 $(SED) -i "s#DatabaseV[A-Za-z0-9]\+()\.#DatabaseV$(API_VERSION)().#g"
	@grep -rHn "ReplicationV[A-Za-z0-9]\+()" \
		  "$(ROOT)/pkg/deployment/" \
		  "$(ROOT)/pkg/replication/" \
	      "$(ROOT)/pkg/integrations/" \
		  "$(ROOT)/pkg/operator/" \
	      "$(ROOT)/pkg/operatorV2/" \
		  "$(ROOT)/pkg/server/" \
		  "$(ROOT)/pkg/util/" \
		  "$(ROOT)/pkg/handlers" \
		  "$(ROOT)/pkg/apis/backup/" \
	      "$(ROOT)/pkg/apis/networking/" \
	      "$(ROOT)/pkg/apis/scheduler/" \
	      "$(ROOT)/pkg/apis/platform/" \
	      "$(ROOT)/pkg/upgrade/" \
	      "$(ROOT)/integrations/" \
	  | cut -d ':' -f 1 | sort | uniq \
	  | xargs -n 1 $(SED) -i "s#ReplicationV[A-Za-z0-9]\+()\.#ReplicationV$(API_VERSION)().#g"

synchronize: synchronize-v2alpha1-with-v1

synchronize-v2alpha1-with-v1:
	@echo ">> Please use only COMMUNITY mode! Current RELEASE_MODE=$(RELEASE_MODE)"
	@rm -f pkg/apis/deployment/v1/zz_generated.deepcopy.go pkg/apis/deployment/v2alpha1/zz_generated.deepcopy.go
	@for file in $$(find "$(ROOT)/pkg/apis/deployment/v1/" -type f -exec $(REALPATH) --relative-to "$(ROOT)/pkg/apis/deployment/v1/" {} \;); do if [ ! -d "$(ROOT)/pkg/apis/deployment/v2alpha1/$$(dirname $${file})" ]; then mkdir -p "$(ROOT)/pkg/apis/deployment/v2alpha1/$$(dirname $${file})"; fi; done
	@for file in $$(find "$(ROOT)/pkg/apis/deployment/v1/" -type f -exec $(REALPATH) --relative-to "$(ROOT)/pkg/apis/deployment/v1/" {} \;); do cat "$(ROOT)/pkg/apis/deployment/v1/$${file}" | $(SED) "s#package v1#package v2alpha1#g" | $(SED) 's#ArangoDeploymentVersion = string(utilConstants.VersionV1)#ArangoDeploymentVersion = string(utilConstants.VersionV2Alpha1)#g' > "$(ROOT)/pkg/apis/deployment/v2alpha1/$${file}"; done
	@make update-generated
	@make set-deployment-api-version-v2alpha1 bin
	@make set-deployment-api-version-v1 bin

.PHONY: check-all check-enterprise check-community _check

check-all: check-enterprise check-community license-range-verify

check-enterprise:
	@$(MAKE) _check RELEASE_MODE=enterprise

check-community:
	@$(MAKE) _check RELEASE_MODE=community

_check: sync-crds
	@$(MAKE) fmt yamlfmt license-verify linter run-unit-tests bin-all vulncheck-optional

generate: generate-internal generate-proto fmt yamlfmt license

generate-internal:
	ROOT=$(ROOT) go test --count=1 "$(REPOPATH)/internal/..."

generate-proto:
	PATH="$(PATH):$(GOBUILDDIR)/bin" $(GOBUILDDIR)/bin/protoc -I.:$(GOBUILDDIR)/include/ -I.:$(GOBUILDDIR)/include/googleapis/ \
			--go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
			$(PROTOSOURCES)

.PHONY: fix
fix: license-range fmt license yamlfmt

CRDS:=apps-job \
      backups-backup backups-backuppolicy \
      database-clustersynchronization database-deployment database-member database-task \
      replication-deploymentreplication \
      ml-storage ml-extension ml-job-batch ml-job-cron \
      scheduler-profile scheduler-pod scheduler-deployment scheduler-batchjob scheduler-cronjob \
      analytics-graphanalyticsengine \
      networking-route \
      platform-storage platform-chart platform-service

.PHONY: sync
sync:

.PHONY: sync-crds
sync-crds:
	@cp $(foreach FILE,$(CRDS),"$(ROOT)/chart/kube-arangodb/crds/$(FILE).yaml" ) "$(ROOT)/pkg/crd/crds/"

sync: sync-crds

.PHONY: sync-charts
sync-charts:
	@(cd "$(ROOT)/chart/kube-arangodb"; find . -type d -not -name values.yaml -and -not -name README.md -exec mkdir -p "$(ROOT)/chart/kube-arangodb-enterprise/{}" \;)
	@(cd "$(ROOT)/chart/kube-arangodb"; find . -type f -not -name values.yaml -and -not -name README.md -not -name Chart.yaml -exec cp "$(ROOT)/chart/kube-arangodb/{}" "$(ROOT)/chart/kube-arangodb-enterprise/{}" \;)

	@(cd "$(ROOT)/chart/kube-arangodb"; find . -type d -not -name values.yaml -and -not -name README.md -exec mkdir -p "$(ROOT)/chart/kube-arangodb-enterprise-arm64/{}" \;)
	@(cd "$(ROOT)/chart/kube-arangodb"; find . -type f -not -name values.yaml -and -not -name README.md -not -name Chart.yaml -exec cp "$(ROOT)/chart/kube-arangodb/{}" "$(ROOT)/chart/kube-arangodb-enterprise-arm64/{}" \;)

	@(cd "$(ROOT)/chart/kube-arangodb"; find . -type d -not -name values.yaml -and -not -name README.md -exec mkdir -p "$(ROOT)/chart/kube-arangodb-arm64/{}" \;)
	@(cd "$(ROOT)/chart/kube-arangodb"; find . -type f -not -name values.yaml -and -not -name README.md -not -name Chart.yaml -exec cp "$(ROOT)/chart/kube-arangodb/{}" "$(ROOT)/chart/kube-arangodb-arm64/{}" \;)

sync: sync-charts

ci-check:
	@$(MAKE) tidy vendor generate update-generated synchronize-v2alpha1-with-v1 sync fmt yamlfmt license protolint
	@git checkout -- go.sum # ignore changes in go.sum
	@if [ ! -z "$$(git status --porcelain)" ]; then echo "There are uncommited changes!"; git status; exit 1; fi

.PHONY: reset
reset:
	@git checkout origin/master -- go.mod go.sum
