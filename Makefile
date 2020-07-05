PROJECT := arangodb_operator
SCRIPTDIR := $(shell pwd)
ROOTDIR := $(shell cd $(SCRIPTDIR) && pwd)
VERSION := $(shell cat $(ROOTDIR)/VERSION)
VERSION_MAJOR_MINOR_PATCH := $(shell echo $(VERSION) | cut -f 1 -d '+')
VERSION_MAJOR_MINOR := $(shell echo $(VERSION_MAJOR_MINOR_PATCH) | cut -f 1,2 -d '.')
VERSION_MAJOR := $(shell echo $(VERSION_MAJOR_MINOR) | cut -f 1 -d '.')
COMMIT := $(shell git rev-parse --short HEAD)
DOCKERCLI := $(shell which docker)

GOBUILDDIR := $(SCRIPTDIR)/.gobuild
SRCDIR := $(SCRIPTDIR)
CACHEVOL := $(PROJECT)-gocache
BINDIR := $(ROOTDIR)/bin
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

DOCKERFILE := Dockerfile
DOCKERTESTFILE := Dockerfile.test
DOCKERDURATIONTESTFILE := tests/duration/Dockerfile

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
                           --save=false

HELM_CMD = $(HELM) template "$(ROOTDIR)/chart/$(CHART_NAME)" \
         	       --kube-version 1.14 \
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

BASEUBIIMAGE ?= registry.access.redhat.com/ubi8/ubi-minimal:8.0

ifndef OPERATORIMAGE
	OPERATORIMAGE := $(DOCKERNAMESPACE)/kube-arangodb$(IMAGESUFFIX)
endif
ifndef OPERATORUBIIMAGE
	OPERATORUBIIMAGE := $(DOCKERNAMESPACE)/kube-arangodb$(IMAGESUFFIX)-ubi
endif
ifndef TESTIMAGE
	TESTIMAGE := $(DOCKERNAMESPACE)/kube-arangodb-test$(IMAGESUFFIX)
endif
ifndef DURATIONTESTIMAGE
	DURATIONTESTIMAGE := $(DOCKERNAMESPACE)/kube-arangodb-durationtest$(IMAGESUFFIX)
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
TESTBINNAME := $(PROJECT)_test
TESTBIN := $(BINDIR)/$(TESTBINNAME)
DURATIONTESTBINNAME := $(PROJECT)_duration_test
DURATIONTESTBIN := $(BINDIR)/$(DURATIONTESTBINNAME)
RELEASE := $(GOBUILDDIR)/bin/release
GHRELEASE := $(GOBUILDDIR)/bin/github-release

TESTLENGTHOPTIONS := -test.short
TESTTIMEOUT := 30m
ifeq ($(LONG), 1)
	TESTLENGTHOPTIONS :=
	TESTTIMEOUT := 300m
endif
ifdef VERBOSE
	TESTVERBOSEOPTIONS := -v
endif

SOURCES_QUERY := find $(SRCDIR) -name '*.go' -type f -not -path '$(SRCDIR)/tests/*' -not -path '$(SRCDIR)/vendor/*' -not -path '$(SRCDIR)/.gobuild/*' -not -path '$(SRCDIR)/deps/*' -not -path '$(SRCDIR)/tools/*'
SOURCES := $(shell $(SOURCES_QUERY))
SOURCES_PACKAGES := $(shell $(SOURCES_QUERY) -exec dirname {} \; | sort | uniq)
DASHBOARDSOURCES := $(shell find $(DASHBOARDDIR)/src -name '*.js' -not -path './test/*') $(DASHBOARDDIR)/package.json

ifndef ARANGOSYNCSRCDIR
	ARANGOSYNCSRCDIR := $(SCRIPTDIR)/arangosync
endif
DOCKERARANGOSYNCCTRLFILE=tests/sync/Dockerfile
ifndef ARANGOSYNCTESTCTRLIMAGE
	ARANGOSYNCTESTCTRLIMAGE := $(DOCKERNAMESPACE)/kube-arangodb-sync-test-ctrl$(IMAGESUFFIX)
endif
ifndef ARANGOSYNCTESTIMAGE
	ARANGOSYNCTESTIMAGE := $(DOCKERNAMESPACE)/kube-arangodb-sync-test$(IMAGESUFFIX)
endif
ifndef ARANGOSYNCIMAGE
	ARANGOSYNCIMAGE := $(DOCKERNAMESPACE)/kube-arangodb-sync$(IMAGESUFFIX)
endif
ARANGOSYNCTESTCTRLBINNAME := $(PROJECT)_sync_test_ctrl
ARANGOSYNCTESTCTRLBIN := $(BINDIR)/$(ARANGOSYNCTESTCTRLBINNAME)

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

GOLANGCI_ENABLED=deadcode gocyclo golint varcheck structcheck maligned errcheck \
                 ineffassign interfacer unconvert goconst \
                 megacheck

#GOLANGCI_ENABLED+=dupl - disable dupl check

.PHONY: license-verify
license-verify:
	@echo ">> Verify license of files"
	@go run github.com/google/addlicense -f "./tools/codegen/boilerplate.go.txt" -check $(SOURCES)

.PHONY: fmt
fmt:
	@echo ">> Ensuring style of files"
	@go run golang.org/x/tools/cmd/goimports -w $(SOURCES)

.PHONY: fmt-verify
fmt-verify: license-verify
	@echo ">> Verify files style"
	@if [ X"$$(go run golang.org/x/tools/cmd/goimports -l $(SOURCES) | wc -l)" != X"0" ]; then echo ">> Style errors"; go run golang.org/x/tools/cmd/goimports -l $(SOURCES); exit 1; fi

.PHONY: linter
linter: fmt
	@golangci-lint run --no-config --issues-exit-code=1 --deadline=30m --disable-all \
	                  $(foreach MODE,$(GOLANGCI_ENABLED),--enable $(MODE) ) \
	                  --exclude-use-default=false \
	                  $(SOURCES_PACKAGES)

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
	@git clone --branch kubernetes-1.14.1 https://github.com/kubernetes/code-generator.git $(VENDORDIR)/k8s.io/code-generator
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
			"deployment:v1 replication:v1 storage:v1alpha backup:v1" \
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

dashboard/assets.go: $(DASHBOARDSOURCES) $(DASHBOARDDIR)/Dockerfile.build
	cd $(DASHBOARDDIR) && docker build -t $(DASHBOARDBUILDIMAGE) -f Dockerfile.build $(DASHBOARDDIR)
	@mkdir -p $(DASHBOARDDIR)/build
	docker run --rm \
		-u $(shell id -u):$(shell id -g) \
		-v $(DASHBOARDDIR)/build:/usr/code/build \
		-v $(DASHBOARDDIR)/public:/usr/code/public:ro \
		-v $(DASHBOARDDIR)/src:/usr/code/src:ro \
		$(DASHBOARDBUILDIMAGE)
	go run github.com/jessevdk/go-assets-builder -s /dashboard/build/ -o dashboard/assets.go -p dashboard dashboard/build

.PHONY: bin
bin: $(BIN)

$(BIN): $(SOURCES) dashboard/assets.go VERSION
	@mkdir -p $(BINDIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -installsuffix netgo -ldflags "-X main.projectVersion=$(VERSION) -X main.projectBuild=$(COMMIT)" -o $(BIN) $(REPOPATH)

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
	go test --count=1 $(TESTVERBOSEOPTIONS) \
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

$(TESTBIN): $(GOBUILDDIR) $(SOURCES)
	@mkdir -p $(BINDIR)
	CGO_ENABLED=0 go test -c -installsuffix netgo -ldflags "-X main.projectVersion=$(VERSION) -X main.projectBuild=$(COMMIT)" -o $(TESTBIN) $(REPOPATH)/tests


.PHONY: docker-test
docker-test: $(TESTBIN)
	docker build --quiet -f $(DOCKERTESTFILE) -t $(TESTIMAGE) .

.PHONY: run-upgrade-tests
run-upgrade-tests:
	TESTOPTIONS="-test.run=TestUpgrade" make run-tests

.PHONY: prepare-run-tests
prepare-run-tests:
ifdef PUSHIMAGES
	docker push $(OPERATORIMAGE)
endif
ifneq ($(DEPLOYMENTNAMESPACE), default)
	$(ROOTDIR)/scripts/kube_delete_namespace.sh $(DEPLOYMENTNAMESPACE)
	kubectl create namespace $(DEPLOYMENTNAMESPACE)
endif
	kubectl apply -f $(MANIFESTPATHCRD)
	kubectl apply -f $(MANIFESTPATHSTORAGE)
	kubectl apply -f $(MANIFESTPATHDEPLOYMENT)
	kubectl apply -f $(MANIFESTPATHDEPLOYMENTREPLICATION)
	kubectl apply -f $(MANIFESTPATHBACKUP)
	kubectl apply -f $(MANIFESTPATHTEST)
	$(ROOTDIR)/scripts/kube_create_storage.sh $(DEPLOYMENTNAMESPACE)
	$(ROOTDIR)/scripts/kube_create_license_key_secret.sh "$(DEPLOYMENTNAMESPACE)" '$(ENTERPRISELICENSE)'
	$(ROOTDIR)/scripts/kube_create_backup_remote_secret.sh "$(DEPLOYMENTNAMESPACE)" '$(TEST_REMOTE_SECRET)'

.PHONY: run-tests
run-tests: docker-test
ifdef PUSHIMAGES
	docker push $(OPERATORIMAGE)
	docker push $(TESTIMAGE)
endif
ifneq ($(DEPLOYMENTNAMESPACE), default)
	$(ROOTDIR)/scripts/kube_delete_namespace.sh $(DEPLOYMENTNAMESPACE)
	kubectl create namespace $(DEPLOYMENTNAMESPACE)
endif
	kubectl apply -f $(MANIFESTPATHCRD)
	kubectl apply -f $(MANIFESTPATHSTORAGE)
	kubectl apply -f $(MANIFESTPATHDEPLOYMENT)
	kubectl apply -f $(MANIFESTPATHDEPLOYMENTREPLICATION)
	kubectl apply -f $(MANIFESTPATHBACKUP)
	kubectl apply -f $(MANIFESTPATHTEST)
	$(ROOTDIR)/scripts/kube_create_storage.sh $(DEPLOYMENTNAMESPACE)
	$(ROOTDIR)/scripts/kube_create_license_key_secret.sh "$(DEPLOYMENTNAMESPACE)" '$(ENTERPRISELICENSE)'
	$(ROOTDIR)/scripts/kube_create_backup_remote_secret.sh "$(DEPLOYMENTNAMESPACE)" '$(TEST_REMOTE_SECRET)'
	$(ROOTDIR)/scripts/kube_run_tests.sh $(DEPLOYMENTNAMESPACE) $(TESTIMAGE) "$(ARANGODIMAGE)" '$(ENTERPRISEIMAGE)' '$(TESTTIMEOUT)' '$(TESTLENGTHOPTIONS)' '$(TESTOPTIONS)' '$(TEST_REMOTE_REPOSITORY)'

$(DURATIONTESTBIN): $(SOURCES)
	CGO_ENABLED=0 go build -installsuffix cgo -ldflags "-X main.projectVersion=$(VERSION) -X main.projectBuild=$(COMMIT)" -o $(DURATIONTESTBINNAME) $(REPOPATH)/tests/duration


.PHONY: docker-duration-test
docker-duration-test: $(DURATIONTESTBIN)
	docker build --quiet -f $(DOCKERDURATIONTESTFILE) -t $(DURATIONTESTIMAGE) .
ifdef PUSHIMAGES
	docker push $(DURATIONTESTIMAGE)
endif

.PHONY: cleanup-tests
cleanup-tests:
	kubectl delete ArangoDeployment -n $(DEPLOYMENTNAMESPACE) --all
	sleep 10
ifneq ($(DEPLOYMENTNAMESPACE), default)
	$(ROOTDIR)/scripts/kube_delete_namespace.sh $(DEPLOYMENTNAMESPACE)
endif

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

.PHONY: changelog
changelog:
	docker run --rm \
		-e CHANGELOG_GITHUB_TOKEN=$(shell cat ~/.arangodb/github-token) \
		-v "$(ROOTDIR)":/usr/local/src/your-app \
		ferrarimarco/github-changelog-generator:1.14.3 \
		--user arangodb \
		--project kube-arangodb \
		--no-author

.PHONY: docker-push
docker-push: docker
ifneq ($(DOCKERNAMESPACE), arangodb)
	docker tag $(OPERATORIMAGE) $(DOCKERNAMESPACE)/arangodb-operator
endif
	docker push $(DOCKERNAMESPACE)/arangodb-operator

.PHONY: docker-push-version
docker-push-version: docker
	docker tag arangodb/arangodb-operator arangodb/arangodb-operator:$(VERSION)
	docker tag arangodb/arangodb-operator arangodb/arangodb-operator:$(VERSION_MAJOR_MINOR)
	docker tag arangodb/arangodb-operator arangodb/arangodb-operator:$(VERSION_MAJOR)
	docker tag arangodb/arangodb-operator arangodb/arangodb-operator:latest
	docker push arangodb/arangodb-operator:$(VERSION)
	docker push arangodb/arangodb-operator:$(VERSION_MAJOR_MINOR)
	docker push arangodb/arangodb-operator:$(VERSION_MAJOR)
	docker push arangodb/arangodb-operator:latest

$(RELEASE): $(GOBUILDDIR) $(SOURCES) $(GHRELEASE)
	GOPATH=$(GOBUILDDIR) go build -o $(RELEASE) $(REPOPATH)/tools/release

.PHONY: build-ghrelease
build-ghrelease: $(GHRELEASE)

$(GHRELEASE): $(GOBUILDDIR)
	GOPATH=$(GOBUILDDIR) go build -o $(GHRELEASE) github.com/aktau/github-release

.PHONY: release-patch
release-patch: $(RELEASE)
	GOPATH=$(GOBUILDDIR) $(RELEASE) -type=patch

.PHONY: release-minor
release-minor: $(RELEASE)
	GOPATH=$(GOBUILDDIR) $(RELEASE) -type=minor

.PHONY: release-major
release-major: $(RELEASE)
	GOPATH=$(GOBUILDDIR) $(RELEASE) -type=major

## Kubernetes utilities

.PHONY: minikube-start
minikube-start:
	minikube start --cpus=4 --memory=6144

.PHONY: delete-operator
delete-operator:
	kubectl delete -f $(MANIFESTPATHTEST) --ignore-not-found
	kubectl delete -f $(MANIFESTPATHDEPLOYMENT) --ignore-not-found
	kubectl delete -f $(MANIFESTPATHDEPLOYMENTREPLICATION) --ignore-not-found
	kubectl delete -f $(MANIFESTPATHBACKUP) --ignore-not-found
	kubectl delete -f $(MANIFESTPATHSTORAGE) --ignore-not-found
	kubectl delete -f $(MANIFESTPATHCRD) --ignore-not-found

.PHONY: redeploy-operator
redeploy-operator: delete-operator manifests
	kubectl apply -f $(MANIFESTPATHCRD)
	kubectl apply -f $(MANIFESTPATHSTORAGE)
	kubectl apply -f $(MANIFESTPATHDEPLOYMENT)
	kubectl apply -f $(MANIFESTPATHDEPLOYMENTREPLICATION)
	kubectl apply -f $(MANIFESTPATHBACKUP)
	kubectl apply -f $(MANIFESTPATHTEST)
	kubectl get pods

## ArangoSync Tests

$(ARANGOSYNCTESTCTRLBIN): $(GOBUILDDIR) $(SOURCES)
	@mkdir -p $(BINDIR)
	CGO_ENABLED=0 go build -installsuffix cgo -ldflags "-X main.projectVersion=$(VERSION) -X main.projectBuild=$(COMMIT)" -o $(ARANGOSYNCTESTCTRLBIN) $(REPOPATH)/tests/sync

.PHONY: check-sync-vars
check-sync-vars:
ifndef ARANGOSYNCSRCDIR
	@echo ARANGOSYNCSRCDIR must point to the arangosync source directory
	@exit 1
endif
ifndef ARANGODIMAGE
	@echo ARANGODIMAGE must point to the usable arangodb enterprise image
	@exit 1
endif
ifndef ENTERPRISELICENSE
	@echo For tests using ArangoSync you most likely need the license key. Please set ENTERPRISELICENSE.
	@exit 1
endif
	@echo Using ArangoSync source at $(ARANGOSYNCSRCDIR)
	@echo Using ArangoDB image $(ARANGODIMAGE)

.PHONY: docker-sync
docker-sync: check-sync-vars
	SYNCIMAGE=$(ARANGOSYNCIMAGE) TESTIMAGE=$(ARANGOSYNCTESTIMAGE) $(MAKE) -C $(ARANGOSYNCSRCDIR) docker docker-test

.PHONY:
docker-sync-test-ctrl: $(ARANGOSYNCTESTCTRLBIN)
	docker build --quiet -f $(DOCKERARANGOSYNCCTRLFILE) -t $(ARANGOSYNCTESTCTRLIMAGE) .

.PHONY:
run-sync-tests: check-vars docker-sync docker-sync-test-ctrl prepare-run-tests
ifdef PUSHIMAGES
	docker push $(ARANGOSYNCTESTCTRLIMAGE)
	docker push $(ARANGOSYNCTESTIMAGE)
	docker push $(ARANGOSYNCIMAGE)
endif
	$(ROOTDIR)/scripts/kube_run_sync_tests.sh $(DEPLOYMENTNAMESPACE) '$(ARANGODIMAGE)' '$(ARANGOSYNCIMAGE)' '$(ARANGOSYNCTESTIMAGE)' '$(ARANGOSYNCTESTCTRLIMAGE)' '$(TESTOPTIONS)'

.PHONY: init
init: tools vendor

.PHONY: tools
tools:
	@echo ">> Fetching goimports"
	@go get -u golang.org/x/tools/cmd/goimports
	@echo ">> Fetching license check"
	@go get -u github.com/google/addlicense

.PHONY: vendor
vendor:
	@echo ">> Updating vendor"
	@go mod vendor