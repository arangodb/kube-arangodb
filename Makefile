CURRENT=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
ROOT:=$(CURRENT)

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

GOBUILDDIR := $(SCRIPTDIR)/.gobuild
SRCDIR := $(SCRIPTDIR)
CACHEVOL := $(PROJECT)-gocache
BINDIR := $(ROOTDIR)/bin
VBINDIR := $(BINDIR)/$(RELEASE_MODE)
VENDORDIR := $(ROOTDIR)/deps
DASHBOARDDIR := $(ROOTDIR)/dashboard

ORGPATH := github.com/arangodb
ORGDIR := $(GOBUILDDIR)/src/$(ORGPATH)
REPONAME := kube-arangodb
REPODIR := $(ORGDIR)/$(REPONAME)
REPOPATH := $(ORGPATH)/$(REPONAME)

GOPATH := $(GOBUILDDIR)
GOVERSION := 1.10.0-alpine

PULSAR := $(GOBUILDDIR)/bin/pulsar$(shell go env GOEXE)
GOASSETSBUILDER := $(GOBUILDDIR)/bin/go-assets-builder$(shell go env GOEXE)

BUILDTIME = $(shell go run "$(ROOT)/tools/dategen/")

DOCKERFILE := Dockerfile

HELM ?= $(shell which helm)

UPPER = $(shell echo '$1' | tr '[:lower:]' '[:upper:]')
LOWER = $(shell echo '$1' | tr '[:upper:]' '[:lower:]')
UPPER_ENV = $(shell echo '$1' | tr '[:lower:]' '[:upper:]' | tr -d '-')

.PHONY: helm
helm:
ifeq ($(HELM),)
	$(error Before templating you need to install helm in PATH or export helm binary using "export HELM=<path to helm>")
endif

HELM_PACKAGE_CMD = $(HELM) package "$(ROOTDIR)/chart/$(CHART_NAME)" \
                           -d "$(ROOTDIR)/bin/charts" \
                           --save=false --version "$(VERSION_MAJOR_MINOR_PATCH)"

HELM_CMD = $(HELM) template "$(ROOTDIR)/chart/$(CHART_NAME)" \
         	       --name "$(NAME)" \
         	       --set "operator.image=$(OPERATORIMAGE)" \
         	       --set "operator.imagePullPolicy=Always" \
         	       --set "operator.resources=null" \
         	       --namespace "$(DEPLOYMENTNAMESPACE)"

ifndef LOCALONLY
	PUSHIMAGES := 1
endif

ifdef IMAGETAG
	IMAGESUFFIX := :$(IMAGETAG)
else
	IMAGESUFFIX := :dev
endif

ifeq ($(MANIFESTSUFFIX),-)
	# Release setting
	MANIFESTSUFFIX :=
else
ifndef MANIFESTSUFFIX
	MANIFESTSUFFIX := -dev
endif
endif
MANIFESTPATHCRD := manifests/arango-crd$(MANIFESTSUFFIX).yaml
MANIFESTPATHDEPLOYMENT := manifests/arango-deployment$(MANIFESTSUFFIX).yaml
MANIFESTPATHDEPLOYMENTREPLICATION := manifests/arango-deployment-replication$(MANIFESTSUFFIX).yaml
MANIFESTPATHBACKUP := manifests/arango-backup$(MANIFESTSUFFIX).yaml
MANIFESTPATHSTORAGE := manifests/arango-storage$(MANIFESTSUFFIX).yaml
MANIFESTPATHALL := manifests/arango-all$(MANIFESTSUFFIX).yaml
MANIFESTPATHTEST := manifests/arango-test$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHCRD := manifests/kustomize/crd/arango-crd$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHDEPLOYMENT := manifests/kustomize/deployment/arango-deployment$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHDEPLOYMENTREPLICATION := manifests/kustomize/deployment-replication/arango-deployment-replication$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHBACKUP := manifests/kustomize/backup/arango-backup$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHSTORAGE := manifests/kustomize/storage/arango-storage$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHALL := manifests/kustomize/all/arango-all$(MANIFESTSUFFIX).yaml
KUSTOMIZEPATHTEST := manifests/kustomize/test/arango-test$(MANIFESTSUFFIX).yaml
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
VBIN := $(BINDIR)/$(RELEASE_MODE)/$(BINNAME)

ifdef VERBOSE
	TESTVERBOSEOPTIONS := -v
endif

EXCLUDE_DIRS := tests vendor .gobuild deps tools
SOURCES_QUERY := find ./ -type f -name '*.go' $(foreach EXCLUDE_DIR,$(EXCLUDE_DIRS), ! -path "./$(EXCLUDE_DIR)/*")
SOURCES := $(shell $(SOURCES_QUERY))
DASHBOARDSOURCES := $(shell find $(DASHBOARDDIR)/src -name '*.js') $(DASHBOARDDIR)/package.json
LINT_EXCLUDES:=
ifeq ($(RELEASE_MODE),enterprise)
LINT_EXCLUDES+=.*\.community\.go$$
else
LINT_EXCLUDES+=.*\.enterprise\.go$$
endif


.DEFAULT_GOAL := all
.PHONY: all
all: check-vars verify-generated build

.PHONY: compile
compile:	check-vars build

# allall  is now obsolete
.PHONY: allall
allall: all

#
# Tip: Run `eval $(minikube docker-env)` before calling make if you're developing on minikube.
#

GOLANGCI_ENABLED=deadcode gosimple govet ineffassign staticcheck structcheck typecheck unconvert unparam unused varcheck
#GOLANGCI_ENABLED=gocyclo goconst golint maligned errcheck interfacer megacheck
#GOLANGCI_ENABLED+=dupl - disable dupl check

.PHONY: license-verify
license-verify:
	@echo ">> Verify license of files"
	@$(GOPATH)/bin/addlicense -f "./tools/codegen/boilerplate.go.txt" -check $(SOURCES)

.PHONY: fmt
fmt:
	@echo ">> Ensuring style of files"
	@$(GOPATH)/bin/goimports -w $(SOURCES)

.PHONY: license
license:
	@echo ">> Ensuring license of files"
	@$(GOPATH)/bin/addlicense -f "./tools/codegen/boilerplate.go.txt" $(SOURCES)

.PHONY: fmt-verify
fmt-verify: license-verify
	@echo ">> Verify files style"
	@if [ X"$$($(GOPATH)/bin/goimports -l $(SOURCES) | wc -l)" != X"0" ]; then echo ">> Style errors"; $(GOPATH)/bin/goimports -l $(SOURCES); exit 1; fi

.PHONY: linter
linter:
	$(GOPATH)/bin/golangci-lint run --build-tags "$(RELEASE_MODE)" --no-config --issues-exit-code=1 --deadline=30m --exclude-use-default=false \
	--disable-all $(foreach EXCLUDE_DIR,$(EXCLUDE_DIRS),--skip-dirs $(EXCLUDE_DIR)) \
	$(foreach MODE,$(GOLANGCI_ENABLED),--enable $(MODE)) \
	$(foreach LINT_EXCLUDE,$(LINT_EXCLUDES),--exclude '$(LINT_EXCLUDE)') \
	./...

.PHONY: build
build: docker manifests

ifndef IGNORE_UBI
build: docker-ubi
endif

.PHONY: clean
clean:
	rm -Rf $(BIN) $(BINDIR) $(DASHBOARDDIR)/build $(DASHBOARDDIR)/node_modules

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
	@git clone --branch kubernetes-1.19.8 https://github.com/kubernetes/code-generator.git $(VENDORDIR)/k8s.io/code-generator
	@rm -Rf $(VENDORDIR)/k8s.io/code-generator/.git


.PHONY: update-generated
update-generated:
	@rm -fr $(ORGDIR)
	@mkdir -p $(ORGDIR)
	@ln -s -f $(SCRIPTDIR) $(ORGDIR)/kube-arangodb
	GOPATH=$(GOBUILDDIR) $(VENDORDIR)/k8s.io/code-generator/generate-groups.sh  \
			"all" \
			"github.com/arangodb/kube-arangodb/pkg/generated" \
			"github.com/arangodb/kube-arangodb/pkg/apis" \
			"deployment:v1 replication:v1 storage:v1alpha backup:v1 deployment:v2alpha1 replication:v2alpha1" \
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
	$(GOPATH)/bin/go-assets-builder -s /dashboard/build/ -o dashboard/assets.go -p dashboard dashboard/build

.PHONY: bin
bin: $(BIN)

$(VBIN): $(SOURCES) dashboard/assets.go VERSION
	@mkdir -p $(VBINDIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --tags "$(RELEASE_MODE)" -installsuffix netgo -ldflags "-X $(REPOPATH)/pkg/version.version=$(VERSION) -X $(REPOPATH)/pkg/version.buildDate=$(BUILDTIME) -X $(REPOPATH)/pkg/version.build=$(COMMIT)" -o $(VBIN) $(REPOPATH)

$(BIN): $(VBIN)
	@cp "$(VBIN)" "$(BIN)"

.PHONY: docker
docker: check-vars $(BIN)
	docker build --no-cache -f $(DOCKERFILE) --build-arg "VERSION=${VERSION_MAJOR_MINOR_PATCH}" -t $(OPERATORIMAGE) .
ifdef PUSHIMAGES
	docker push $(OPERATORIMAGE)
endif

.PHONY: docker-ubi
docker-ubi: check-vars $(BIN)
	docker build --no-cache -f "$(DOCKERFILE).ubi" --build-arg "VERSION=${VERSION_MAJOR_MINOR_PATCH}" --build-arg "IMAGE=$(BASEUBIIMAGE)" -t $(OPERATORUBIIMAGE)-local-only-build .
	docker build --no-cache -f $(DOCKERFILE) --build-arg "VERSION=${VERSION_MAJOR_MINOR_PATCH}" --build-arg "IMAGE=$(OPERATORUBIIMAGE)-local-only-build" -t $(OPERATORUBIIMAGE) .
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

$(eval $(call manifest-generator, crd, kube-arangodb-crd))

$(eval $(call manifest-generator, test, kube-arangodb-test))

$(eval $(call manifest-generator, deployment, kube-arangodb, \
       --set "operator.features.deployment=true" \
	   --set "operator.features.deploymentReplications=false" \
	   --set "operator.features.storage=false" \
	   --set "operator.features.backup=false"))

$(eval $(call manifest-generator, deployment-replication, kube-arangodb, \
       --set "operator.features.deployment=false" \
       --set "operator.features.deploymentReplications=true" \
       --set "operator.features.storage=false" \
       --set "operator.features.backup=false"))

$(eval $(call manifest-generator, storage, kube-arangodb, \
       --set "operator.features.deployment=false" \
       --set "operator.features.deploymentReplications=false" \
       --set "operator.features.storage=true" \
       --set "operator.features.backup=false"))

$(eval $(call manifest-generator, backup, kube-arangodb, \
       --set "operator.features.deployment=false" \
       --set "operator.features.deploymentReplications=false" \
       --set "operator.features.storage=false" \
       --set "operator.features.backup=true"))

$(eval $(call manifest-generator, all, kube-arangodb, \
       --set "operator.features.deployment=true" \
       --set "operator.features.deploymentReplications=true" \
       --set "operator.features.storage=true" \
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
		$(REPOPATH)/pkg/util/k8sutil \
		$(REPOPATH)/pkg/util/k8sutil/test \
		$(REPOPATH)/pkg/util/probe \
		$(REPOPATH)/pkg/util/validation \
		$(REPOPATH)/pkg/backup/...

# Release building

.PHONY: patch-readme
patch-readme:
	$(ROOTDIR)/scripts/patch_readme.sh $(VERSION_MAJOR_MINOR_PATCH)

.PHONY: patch-examples
patch-examples:
	$(ROOTDIR)/scripts/patch_examples.sh $(VERSION_MAJOR_MINOR_PATCH)

.PHONY: patch-release
patch-release: patch-readme patch-examples

.PHONY: patch-chart
patch-chart:
	$(ROOTDIR)/scripts/patch_chart.sh "$(VERSION_MAJOR_MINOR_PATCH)" "$(OPERATORIMAGE)"

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: deps-reload
deps-reload: tidy init

.PHONY: init
init: tools update-generated $(BIN) vendor

.PHONY: tools
tools: update-vendor
	@echo ">> Fetching golangci-lint linter"
	@GOBIN=$(GOPATH)/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.40.0
	@echo ">> Fetching goimports"
	@GOBIN=$(GOPATH)/bin go get golang.org/x/tools/cmd/goimports@0bb7e5c47b1a31f85d4f173edc878a8e049764a5
	@echo ">> Fetching license check"
	@GOBIN=$(GOPATH)/bin go get github.com/google/addlicense@6d92264d717064f28b32464f0f9693a5b4ef0239
	@echo ">> Fetching GO Assets Builder"
	@GOBIN=$(GOPATH)/bin go get github.com/jessevdk/go-assets-builder@b8483521738fd2198ecfc378067a4e8a6079f8e5

.PHONY: vendor
vendor:
	@echo ">> Updating vendor"
	@go mod vendor

set-deployment-api-version-v2alpha1: export API_VERSION=2alpha1
set-deployment-api-version-v2alpha1: set-api-version/deployment set-api-version/replication

set-deployment-api-version-v1: export API_VERSION=1
set-deployment-api-version-v1: set-api-version/deployment set-api-version/replication

set-api-version/%:
	@grep -rHn "github.com/arangodb/kube-arangodb/pkg/apis/$*/v[A-Za-z0-9]\+" \
	      "$(ROOT)/pkg/deployment/" \
	      "$(ROOT)/pkg/replication/" \
	      "$(ROOT)/pkg/operator/" \
	      "$(ROOT)/pkg/server/" \
	      "$(ROOT)/pkg/util/" \
	      "$(ROOT)/pkg/backup/" \
	      "$(ROOT)/pkg/apis/backup/" \
	  | cut -d ':' -f 1 | sort | uniq \
	  | xargs -n 1 sed -i "s#github.com/arangodb/kube-arangodb/pkg/apis/$*/v[A-Za-z0-9]\+#github.com/arangodb/kube-arangodb/pkg/apis/$*/v$(API_VERSION)#g"
	@grep -rHn "DatabaseV[A-Za-z0-9]\+()" \
		  "$(ROOT)/pkg/deployment/" \
	      "$(ROOT)/pkg/replication/" \
	      "$(ROOT)/pkg/operator/" \
	      "$(ROOT)/pkg/server/" \
		  "$(ROOT)/pkg/util/" \
	      "$(ROOT)/pkg/backup/" \
	      "$(ROOT)/pkg/apis/backup/" \
	  | cut -d ':' -f 1 | sort | uniq \
	  | xargs -n 1 sed -i "s#DatabaseV[A-Za-z0-9]\+()\.#DatabaseV$(API_VERSION)().#g"
	@grep -rHn "ReplicationV[A-Za-z0-9]\+()" \
		  "$(ROOT)/pkg/deployment/" \
		  "$(ROOT)/pkg/replication/" \
		  "$(ROOT)/pkg/operator/" \
		  "$(ROOT)/pkg/server/" \
		  "$(ROOT)/pkg/util/" \
		  "$(ROOT)/pkg/backup/" \
		  "$(ROOT)/pkg/apis/backup/" \
	  | cut -d ':' -f 1 | sort | uniq \
	  | xargs -n 1 sed -i "s#ReplicationV[A-Za-z0-9]\+()\.#ReplicationV$(API_VERSION)().#g"

synchronize-v2alpha1-with-v1:
	@rm -f pkg/apis/deployment/v1/zz_generated.deepcopy.go pkg/apis/deployment/v2alpha1/zz_generated.deepcopy.go
	@for file in $$(find "$(ROOT)/pkg/apis/deployment/v1/" -type f -exec realpath --relative-to "$(ROOT)/pkg/apis/deployment/v1/" {} \;); do if [ ! -d "$(ROOT)/pkg/apis/deployment/v2alpha1/$$(dirname $${file})" ]; then mkdir -p "$(ROOT)/pkg/apis/deployment/v2alpha1/$$(dirname $${file})"; fi; done
	@for file in $$(find "$(ROOT)/pkg/apis/deployment/v1/" -type f -exec realpath --relative-to "$(ROOT)/pkg/apis/deployment/v1/" {} \;); do cat "$(ROOT)/pkg/apis/deployment/v1/$${file}" | sed "s#package v1#package v2alpha1#g" | sed 's#ArangoDeploymentVersion = "v1"#ArangoDeploymentVersion = "v2alpha1"#g' > "$(ROOT)/pkg/apis/deployment/v2alpha1/$${file}"; done
	@make update-generated
	@make set-deployment-api-version-v2alpha1 bin
	@make set-deployment-api-version-v1 bin
