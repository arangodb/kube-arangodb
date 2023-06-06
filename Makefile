CURRENT=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
ROOT:=$(CURRENT)

SED ?= sed
REALPATH ?= realpath

ifeq ($(shell uname),Darwin)
	SED ?= gsed
	REALPATH ?= grealpath
endif

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

GOBUILDDIR := $(SCRIPTDIR)/.gobuild
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

include $(ROOT)/$(RELEASE_MODE).mk

ifndef KEEP_GOPATH
	GOPATH := $(GOBUILDDIR)
endif

GOBUILDARGS ?=
GOBASEVERSION := 1.19
GOVERSION := $(GOBASEVERSION)-alpine3.17
DISTRIBUTION := alpine:3.15

PULSAR := $(GOBUILDDIR)/bin/pulsar$(shell go env GOEXE)
GOASSETSBUILDER := $(GOBUILDDIR)/bin/go-assets-builder$(shell go env GOEXE)

BUILDTIME = $(shell go run "$(ROOT)/tools/dategen/")

HELM ?= $(shell which helm)

UPPER = $(shell echo '$1' | tr '[:lower:]' '[:upper:]')
LOWER = $(shell echo '$1' | tr '[:upper:]' '[:lower:]')
UPPER_ENV = $(shell echo '$1' | tr '[:lower:]' '[:upper:]' | tr -d '-')

.PHONY: helm
helm:
ifeq ($(HELM),)
	$(error Before templating you need to install helm in PATH or export helm binary using "export HELM=<path to helm>")
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
MANIFESTPATHDEPLOYMENT := manifests/arango-deployment$(MANIFESTSUFFIX).yaml
MANIFESTPATHDEPLOYMENTREPLICATION := manifests/arango-deployment-replication$(MANIFESTSUFFIX).yaml
MANIFESTPATHBACKUP := manifests/arango-backup$(MANIFESTSUFFIX).yaml
MANIFESTPATHAPPS := manifests/arango-apps$(MANIFESTSUFFIX).yaml
MANIFESTPATHK2KCLUSTERSYNC := manifests/arango-k2kclustersync$(MANIFESTSUFFIX).yaml
MANIFESTPATHSTORAGE := manifests/arango-storage$(MANIFESTSUFFIX).yaml
MANIFESTPATHALL := manifests/arango-all$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHCRD := manifests/kustomize/crd/arango-crd$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHDEPLOYMENT := manifests/kustomize/deployment/arango-deployment$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHDEPLOYMENTREPLICATION := manifests/kustomize/deployment-replication/arango-deployment-replication$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHBACKUP := manifests/kustomize/backup/arango-backup$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHAPPS := manifests/kustomize/apps/arango-apps$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHK2KCLUSTERSYNC := manifests/kustomize/k2kclustersync/arango-k2kclustersync$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHSTORAGE := manifests/kustomize/storage/arango-storage$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHALL := manifests/kustomize/all/arango-all$(MANIFESTSUFFIX).yaml
else
MANIFESTPATHCRD := manifests/enterprise-crd$(MANIFESTSUFFIX).yaml
MANIFESTPATHDEPLOYMENT := manifests/enterprise-deployment$(MANIFESTSUFFIX).yaml
MANIFESTPATHDEPLOYMENTREPLICATION := manifests/enterprise-deployment-replication$(MANIFESTSUFFIX).yaml
MANIFESTPATHBACKUP := manifests/enterprise-backup$(MANIFESTSUFFIX).yaml
MANIFESTPATHAPPS := manifests/enterprise-apps$(MANIFESTSUFFIX).yaml
MANIFESTPATHK2KCLUSTERSYNC := manifests/enterprise-k2kclustersync$(MANIFESTSUFFIX).yaml
MANIFESTPATHSTORAGE := manifests/enterprise-storage$(MANIFESTSUFFIX).yaml
MANIFESTPATHALL := manifests/enterprise-all$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHCRD := manifests/kustomize-enterprise/crd/enterprise-crd$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHDEPLOYMENT := manifests/kustomize-enterprise/deployment/enterprise-deployment$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHDEPLOYMENTREPLICATION := manifests/kustomize-enterprise/deployment-replication/enterprise-deployment-replication$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHBACKUP := manifests/kustomize-enterprise/backup/enterprise-backup$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHAPPS := manifests/kustomize-enterprise/apps/enterprise-apps$(MANIFESTSUFFIX).yaml
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
VBIN_LINUX_AMD64 := $(BINDIR)/$(RELEASE_MODE)/linux/amd64/$(BINNAME)
VBIN_LINUX_ARM64 := $(BINDIR)/$(RELEASE_MODE)/linux/arm64/$(BINNAME)

BIN_OPS_NAME := $(PROJECT)_ops
BIN_OPS := $(BINDIR)/$(BIN_OPS_NAME)
VBIN_OPS_LINUX_AMD64 := $(BINDIR)/$(RELEASE_MODE)/linux/amd64/$(BIN_OPS_NAME)
VBIN_OPS_LINUX_ARM64 := $(BINDIR)/$(RELEASE_MODE)/linux/arm64/$(BIN_OPS_NAME)

ifdef VERBOSE
	TESTVERBOSEOPTIONS := -v
endif

EXCLUDE_DIRS := vendor .gobuild deps tools pkg/generated/clientset pkg/generated/informers pkg/generated/listers
EXCLUDE_FILES := *generated.deepcopy.go
SOURCES_QUERY := find ./ -type f -name '*.go' ! -name '*.pb.go' $(foreach EXCLUDE_DIR,$(EXCLUDE_DIRS), ! -path "*/$(EXCLUDE_DIR)/*") $(foreach EXCLUDE_FILE,$(EXCLUDE_FILES), ! -path "*/$(EXCLUDE_FILE)")
SOURCES := $(shell $(SOURCES_QUERY))

YAML_EXCLUDE_DIRS := vendor .gobuild deps tools pkg/generated/clientset pkg/generated/informers pkg/generated/listers chart/kube-arangodb/templates chart/kube-arangodb-crd/templates chart/arangodb-ingress-proxy/templates
YAML_EXCLUDE_FILES :=
YAML_QUERY := find ./ -type f -name '*.yaml' $(foreach EXCLUDE_DIR,$(YAML_EXCLUDE_DIRS), ! -path "*/$(EXCLUDE_DIR)/*") $(foreach EXCLUDE_FILE,$(YAML_EXCLUDE_FILES), ! -path "*/$(EXCLUDE_FILE)")
YAMLS := $(shell $(YAML_QUERY))

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
all: check-vars verify-generated build

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
	@$(GOPATH)/bin/addlicense -f "./tools/codegen/license-header.txt" -check $(SOURCES) $(PROTOSOURCES)

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
	@$(GOPATH)/bin/yamlfmt -w $(YAMLS)
	@$(GOPATH)/bin/yamlfmt -w $(YAMLS)

.PHONY: license
license:
	@echo ">> Ensuring license of files"
	@$(GOPATH)/bin/addlicense -f "./tools/codegen/license-header.txt" $(SOURCES) $(PROTOSOURCES)

.PHONY: fmt-verify
fmt-verify: license-verify
	@echo ">> Verify files style"
	@if [ X"$$($(GOPATH)/bin/goimports -l $(SOURCES) | wc -l)" != X"0" ]; then echo ">> Style errors"; $(GOPATH)/bin/goimports -l $(SOURCES); exit 1; fi

.PHONY: linter
linter:
	@$(GOPATH)/bin/golangci-lint run --build-tags "$(RELEASE_MODE)" $(foreach LINT_EXCLUDE,$(LINT_EXCLUDES),--exclude '$(LINT_EXCLUDE)') ./...

.PHONY: linter-fix
linter-fix:
	@$(GOPATH)/bin/golangci-lint run --fix --build-tags "$(RELEASE_MODE)" $(foreach LINT_EXCLUDE,$(LINT_EXCLUDES),--exclude '$(LINT_EXCLUDE)') ./...

.PHONY: vulncheck
vulncheck:
	@echo ">> Checking for known vulnerabilities"
	@-$(GOPATH)/bin/govulncheck --tags $(RELEASE_MODE) ./...

.PHONY: build
build: docker manifests

ifndef IGNORE_UBI
build: docker-ubi
endif

.PHONY: clean
clean:
	rm -Rf $(BIN) $(BINDIR) $(DASHBOARDDIR)/build $(DASHBOARDDIR)/node_modules $(VBIN_LINUX_AMD64) $(VBIN_LINUX_ARM64) $(VBIN_OPS_LINUX_AMD64) $(VBIN_OPS_LINUX_ARM64)

.PHONY: check-vars
check-vars:
ifndef DOCKERNAMESPACE
	@echo "DOCKERNAMESPACE must be set"
	@exit 1
endif
	@echo "Using docker namespace: $(DOCKERNAMESPACE)"

.PHONY: update-vendor
update-vendor:
	@rm -Rf $(VENDORDIR)/k8s.io/code-generator
	@git clone --branch kubernetes-1.22.15 https://github.com/kubernetes/code-generator.git $(VENDORDIR)/k8s.io/code-generator
	@rm -Rf $(VENDORDIR)/k8s.io/code-generator/.git


.PHONY: update-generated
update-generated:
	@rm -fr $(ORGDIR)
	@mkdir -p $(ORGDIR)
	@ln -s -f $(SCRIPTDIR) $(ORGDIR)/kube-arangodb
	@$(SED) -e 's/^/\/\/ /' -e 's/ *$$//' $(ROOTDIR)/tools/codegen/license-header.txt > $(ROOTDIR)/tools/codegen/boilerplate.go.txt
	GOPATH=$(GOBUILDDIR) $(VENDORDIR)/k8s.io/code-generator/generate-groups.sh  \
			"all" \
			"github.com/arangodb/kube-arangodb/pkg/generated" \
			"github.com/arangodb/kube-arangodb/pkg/apis" \
			"deployment:v1 replication:v1 storage:v1alpha backup:v1 deployment:v2alpha1 replication:v2alpha1 apps:v1" \
			--go-header-file "./tools/codegen/boilerplate.go.txt" \
			$(VERIFYARGS)
	GOPATH=$(GOBUILDDIR) $(VENDORDIR)/k8s.io/code-generator/generate-groups.sh  \
			"deepcopy" \
			"github.com/arangodb/kube-arangodb/pkg/generated" \
			"github.com/arangodb/kube-arangodb/pkg/apis" \
			"shared:v1" \
			--go-header-file "./tools/codegen/boilerplate.go.txt" \
			$(VERIFYARGS)

.PHONY: verify-generated
verify-generated:
	@${MAKE} -B -s VERIFYARGS=--verify-only update-generated

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

.PHONY: bin bin-all
bin: $(BIN)
bin-all: $(BIN) $(VBIN_LINUX_AMD64) $(VBIN_LINUX_ARM64)

$(VBIN_LINUX_AMD64): $(SOURCES) dashboard/assets.go VERSION
	@mkdir -p $(BINDIR)/$(RELEASE_MODE)/linux/amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${GOBUILDARGS} --tags "$(RELEASE_MODE)" $(COMPILE_DEBUG_FLAGS) -installsuffix netgo -ldflags "-X $(REPOPATH)/pkg/version.version=$(VERSION) -X $(REPOPATH)/pkg/version.buildDate=$(BUILDTIME) -X $(REPOPATH)/pkg/version.build=$(COMMIT)" -o $(VBIN_LINUX_AMD64) ./cmd/main
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${GOBUILDARGS} --tags "$(RELEASE_MODE)" $(COMPILE_DEBUG_FLAGS) -installsuffix netgo -ldflags "-X $(REPOPATH)/pkg/version.version=$(VERSION) -X $(REPOPATH)/pkg/version.buildDate=$(BUILDTIME) -X $(REPOPATH)/pkg/version.build=$(COMMIT)" -o $(VBIN_OPS_LINUX_AMD64) ./cmd/main-ops

$(VBIN_LINUX_ARM64): $(SOURCES) dashboard/assets.go VERSION
	@mkdir -p $(BINDIR)/$(RELEASE_MODE)/linux/arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ${GOBUILDARGS} --tags "$(RELEASE_MODE)" $(COMPILE_DEBUG_FLAGS) -installsuffix netgo -ldflags "-X $(REPOPATH)/pkg/version.version=$(VERSION) -X $(REPOPATH)/pkg/version.buildDate=$(BUILDTIME) -X $(REPOPATH)/pkg/version.build=$(COMMIT)" -o $(VBIN_LINUX_ARM64) ./cmd/main
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ${GOBUILDARGS} --tags "$(RELEASE_MODE)" $(COMPILE_DEBUG_FLAGS) -installsuffix netgo -ldflags "-X $(REPOPATH)/pkg/version.version=$(VERSION) -X $(REPOPATH)/pkg/version.buildDate=$(BUILDTIME) -X $(REPOPATH)/pkg/version.build=$(COMMIT)" -o $(VBIN_OPS_LINUX_ARM64) ./cmd/main-ops

$(BIN): $(VBIN_LINUX_AMD64)
	@cp "$(VBIN_LINUX_AMD64)" "$(BIN)"
	@cp "$(VBIN_OPS_LINUX_AMD64)" "$(BIN_OPS)"

.PHONY: docker
docker: check-vars $(VBIN_LINUX_AMD64) $(VBIN_LINUX_ARM64)
ifdef PUSHIMAGES
	docker buildx build --no-cache -f $(DOCKERFILE) --build-arg GOVERSION=$(GOVERSION) --build-arg DISTRIBUTION=$(DISTRIBUTION) \
		--build-arg "VERSION=${VERSION_MAJOR_MINOR_PATCH}" --build-arg "RELEASE_MODE=$(RELEASE_MODE)" \
		--platform linux/amd64,linux/arm64 --push -t $(OPERATORIMAGE) .
else
	docker buildx build --no-cache -f $(DOCKERFILE) --build-arg GOVERSION=$(GOVERSION) --build-arg DISTRIBUTION=$(DISTRIBUTION) \
		--build-arg "VERSION=${VERSION_MAJOR_MINOR_PATCH}" --build-arg "RELEASE_MODE=$(RELEASE_MODE)" \
		--platform linux/amd64,linux/arm64 -t $(OPERATORIMAGE) .
endif

.PHONY: docker-ubi
docker-ubi: check-vars $(VBIN_LINUX_AMD64)
	docker build --no-cache -f "$(DOCKERFILE).ubi" --build-arg "VERSION=${VERSION_MAJOR_MINOR_PATCH}" --build-arg "RELEASE_MODE=$(RELEASE_MODE)" --build-arg "IMAGE=$(BASEUBIIMAGE)" -t $(OPERATORUBIIMAGE)-local-only-build .
	docker build --no-cache -f $(DOCKERFILE) --build-arg "VERSION=${VERSION_MAJOR_MINOR_PATCH}" --build-arg "TARGETARCH=amd64" --build-arg "RELEASE_MODE=$(RELEASE_MODE)" --build-arg "IMAGE=$(OPERATORUBIIMAGE)-local-only-build" -t $(OPERATORUBIIMAGE) .
ifdef PUSHIMAGES
	docker push $(OPERATORUBIIMAGE)
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
	@echo -n > $(MANIFESTPATHCRD)
	@$(foreach FILE,$(CRDS),echo '---\n# File: chart/kube-arangodb/crds/$(FILE).yaml' >> $(MANIFESTPATHCRD) && \
                           cat '$(ROOT)/chart/kube-arangodb/crds/$(FILE).yaml' >> $(MANIFESTPATHCRD) && \
                           echo '\n' >> $(MANIFESTPATHCRD);)
manifests: manifests-crd-file

.PHONY: manifests-crd-kustomize
manifests-crd-kustomize: manifests-crd-file
	@echo Building manifests for CRD - $(KUSTOMIZEPATHCRD)
	@cp "$(MANIFESTPATHCRD)" "$(KUSTOMIZEPATHCRD)"
manifests: manifests-crd-kustomize

$(eval $(call manifest-generator, deployment, kube-arangodb, \
       --set "operator.features.deployment=true" \
	   --set "operator.features.deploymentReplications=false" \
	   --set "operator.features.storage=false" \
	   --set "operator.features.apps=false" \
	   --set "operator.features.k8sToK8sClusterSync=false" \
	   --set "operator.features.backup=false"))

$(eval $(call manifest-generator, deployment-replication, kube-arangodb, \
       --set "operator.features.deployment=false" \
       --set "operator.features.deploymentReplications=true" \
       --set "operator.features.storage=false" \
       --set "operator.features.apps=false" \
       --set "operator.features.k8sToK8sClusterSync=false" \
       --set "operator.features.backup=false"))

$(eval $(call manifest-generator, storage, kube-arangodb, \
       --set "operator.features.deployment=false" \
       --set "operator.features.deploymentReplications=false" \
       --set "operator.features.storage=true" \
       --set "operator.features.apps=false" \
       --set "operator.features.k8sToK8sClusterSync=false" \
       --set "operator.features.backup=false"))

$(eval $(call manifest-generator, backup, kube-arangodb, \
       --set "operator.features.deployment=false" \
       --set "operator.features.deploymentReplications=false" \
       --set "operator.features.storage=false" \
       --set "operator.features.apps=false" \
       --set "operator.features.k8sToK8sClusterSync=false" \
       --set "operator.features.backup=true"))

$(eval $(call manifest-generator, apps, kube-arangodb, \
       --set "operator.features.deployment=false" \
       --set "operator.features.deploymentReplications=false" \
       --set "operator.features.storage=false" \
       --set "operator.features.apps=true" \
       --set "operator.features.k8sToK8sClusterSync=false" \
       --set "operator.features.backup=false"))

$(eval $(call manifest-generator, k2kclustersync, kube-arangodb, \
       --set "operator.features.deployment=false" \
       --set "operator.features.deploymentReplications=false" \
       --set "operator.features.storage=false" \
       --set "operator.features.apps=false" \
       --set "operator.features.k8sToK8sClusterSync=true" \
       --set "operator.features.backup=false"))

$(eval $(call manifest-generator, all, kube-arangodb, \
       --set "operator.features.deployment=true" \
       --set "operator.features.deploymentReplications=true" \
       --set "operator.features.storage=true" \
       --set "operator.features.apps=true" \
       --set "operator.features.k8sToK8sClusterSync=true" \
       --set "operator.features.backup=true"))

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

# Testing

.PHONY: run-unit-tests
run-unit-tests: $(SOURCES)
	go test --count=1 --tags "$(RELEASE_MODE)" $(TESTVERBOSEOPTIONS) \
		$(REPOPATH)/pkg/apis/backup/... \
		$(REPOPATH)/pkg/apis/deployment/... \
		$(REPOPATH)/pkg/apis/replication/... \
		$(REPOPATH)/pkg/apis/storage/... \
		$(REPOPATH)/pkg/deployment/... \
		$(REPOPATH)/pkg/storage \
	    $(REPOPATH)/pkg/crd/... \
		$(REPOPATH)/pkg/util/... \
		$(REPOPATH)/pkg/handlers/...

# Release building

.PHONY: patch-readme
patch-readme:
	$(ROOTDIR)/scripts/patch_readme.sh $(VERSION_MAJOR_MINOR_PATCH)

.PHONY: patch-examples
patch-examples:
	$(ROOTDIR)/scripts/patch_examples.sh $(VERSION_MAJOR_MINOR_PATCH)

.PHONY: patch-release
patch-release: patch-readme patch-examples patch-chart

.PHONY: patch-chart
patch-chart:
	$(ROOTDIR)/scripts/patch_chart.sh "$(VERSION_MAJOR_MINOR_PATCH)" "$(OPERATORIMAGE)"

.PHONY: tidy
tidy:
	@go mod tidy -compat=$(GOBASEVERSION)

.PHONY: deps-reload
deps-reload: tidy init

.PHONY: init
init: vendor tools update-generated $(BIN)

.PHONY: tools-min
tools-min: update-vendor
	@echo ">> Fetching golangci-lint linter"
	@GOBIN=$(GOPATH)/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2
	@echo ">> Fetching goimports"
	@GOBIN=$(GOPATH)/bin go install golang.org/x/tools/cmd/goimports@0bb7e5c47b1a31f85d4f173edc878a8e049764a5
	@echo ">> Fetching license check"
	@GOBIN=$(GOPATH)/bin go install github.com/google/addlicense@6d92264d717064f28b32464f0f9693a5b4ef0239
	@echo ">> Fetching yamlfmt"
	@GOBIN=$(GOPATH)/bin go install github.com/UltiRequiem/yamlfmt@v1.3.0

.PHONY: tools
tools: tools-min
	@echo ">> Fetching gci"
	@GOBIN=$(GOPATH)/bin go install github.com/daixiang0/gci@v0.3.0
	@echo ">> Fetching yamlfmt"
	@GOBIN=$(GOPATH)/bin go install github.com/UltiRequiem/yamlfmt@v1.3.0
	@echo ">> Downloading protobuf compiler..."
	@curl -L ${PROTOC_URL} -o $(GOPATH)/protoc.zip
	@echo ">> Unzipping protobuf compiler..."
	@unzip -o $(GOPATH)/protoc.zip -d $(GOPATH)/
	@chmod +x $(GOPATH)/bin/protoc
	@echo ">> Fetching protoc go plugins..."
	@GOBIN=$(GOPATH)/bin go install github.com/golang/protobuf/protoc-gen-go@v1.5.2
	@GOBIN=$(GOPATH)/bin go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	@echo ">> Fetching govulncheck"
	@GOBIN=$(GOPATH)/bin go install golang.org/x/vuln/cmd/govulncheck@v0.1.0

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
	      "$(ROOT)/pkg/operator/" \
	      "$(ROOT)/pkg/server/" \
	      "$(ROOT)/pkg/util/" \
	      "$(ROOT)/pkg/handlers/" \
	      "$(ROOT)/pkg/apis/backup/" \
	      "$(ROOT)/pkg/upgrade/" \
	  | cut -d ':' -f 1 | sort | uniq \
	  | xargs -n 1 $(SED) -i "s#github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/$*/v[A-Za-z0-9]\+#github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/$*/v$(API_VERSION)#g"


set-api-version/%:
	@grep -rHn "github.com/arangodb/kube-arangodb/pkg/apis/$*/v[A-Za-z0-9]\+" \
	      "$(ROOT)/pkg/deployment/" \
	      "$(ROOT)/pkg/replication/" \
	      "$(ROOT)/pkg/operator/" \
	      "$(ROOT)/pkg/server/" \
	      "$(ROOT)/pkg/util/" \
	      "$(ROOT)/pkg/handlers/" \
	      "$(ROOT)/pkg/apis/backup/" \
	      "$(ROOT)/pkg/upgrade/" \
	  | cut -d ':' -f 1 | sort | uniq \
	  | xargs -n 1 $(SED) -i "s#github.com/arangodb/kube-arangodb/pkg/apis/$*/v[A-Za-z0-9]\+#github.com/arangodb/kube-arangodb/pkg/apis/$*/v$(API_VERSION)#g"
	@grep -rHn "DatabaseV[A-Za-z0-9]\+()" \
		  "$(ROOT)/pkg/deployment/" \
	      "$(ROOT)/pkg/replication/" \
	      "$(ROOT)/pkg/operator/" \
	      "$(ROOT)/pkg/server/" \
		  "$(ROOT)/pkg/util/" \
	      "$(ROOT)/pkg/handlers/" \
	      "$(ROOT)/pkg/apis/backup/" \
	      "$(ROOT)/pkg/upgrade/" \
	  | cut -d ':' -f 1 | sort | uniq \
	  | xargs -n 1 $(SED) -i "s#DatabaseV[A-Za-z0-9]\+()\.#DatabaseV$(API_VERSION)().#g"
	@grep -rHn "ReplicationV[A-Za-z0-9]\+()" \
		  "$(ROOT)/pkg/deployment/" \
		  "$(ROOT)/pkg/replication/" \
		  "$(ROOT)/pkg/operator/" \
		  "$(ROOT)/pkg/server/" \
		  "$(ROOT)/pkg/util/" \
		  "$(ROOT)/pkg/handlers" \
		  "$(ROOT)/pkg/apis/backup/" \
	      "$(ROOT)/pkg/upgrade/" \
	  | cut -d ':' -f 1 | sort | uniq \
	  | xargs -n 1 $(SED) -i "s#ReplicationV[A-Za-z0-9]\+()\.#ReplicationV$(API_VERSION)().#g"

synchronize-v2alpha1-with-v1:
	@rm -f pkg/apis/deployment/v1/zz_generated.deepcopy.go pkg/apis/deployment/v2alpha1/zz_generated.deepcopy.go
	@for file in $$(find "$(ROOT)/pkg/apis/deployment/v1/" -type f -exec $(REALPATH) --relative-to "$(ROOT)/pkg/apis/deployment/v1/" {} \;); do if [ ! -d "$(ROOT)/pkg/apis/deployment/v2alpha1/$$(dirname $${file})" ]; then mkdir -p "$(ROOT)/pkg/apis/deployment/v2alpha1/$$(dirname $${file})"; fi; done
	@for file in $$(find "$(ROOT)/pkg/apis/deployment/v1/" -type f -exec $(REALPATH) --relative-to "$(ROOT)/pkg/apis/deployment/v1/" {} \;); do cat "$(ROOT)/pkg/apis/deployment/v1/$${file}" | $(SED) "s#package v1#package v2alpha1#g" | $(SED) 's#ArangoDeploymentVersion = "v1"#ArangoDeploymentVersion = "v2alpha1"#g' > "$(ROOT)/pkg/apis/deployment/v2alpha1/$${file}"; done
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
	@$(MAKE) fmt yamlfmt license-verify linter run-unit-tests bin vulncheck

generate: generate-internal generate-proto fmt

generate-internal:
	ROOT=$(ROOT) go test --count=1 "$(REPOPATH)/internal/..."

generate-proto:
	PATH="$(PATH):$(GOBUILDDIR)/bin" $(GOBUILDDIR)/bin/protoc -I.:$(GOBUILDDIR)/include/ \
			--go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			$(PROTOSOURCES)

.PHONY: fix
fix: license-range fmt license yamlfmt

CRDS:=apps-job \
      backups-backup backups-backuppolicy \
      database-clustersynchronization database-deployment database-member database-task \
      replication-deploymentreplication

.PHONY: sync-crds
sync-crds:
	@cp $(foreach FILE,$(CRDS),"$(ROOT)/chart/kube-arangodb/crds/$(FILE).yaml" ) "$(ROOT)/pkg/crd/crds/"
