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
GOX := $(GOBUILDDIR)/bin/gox$(shell go env GOEXE)
MANIFESTTOOL := $(GOBUILDDIR)/bin/manifest-tool$(shell go env GOEXE)

DOCKERFILE := Dockerfile 
DOCKERTESTFILE := Dockerfile.test
DOCKERDURATIONTESTFILE := tests/duration/Dockerfile

ifndef LOCALONLY 
	PUSHIMAGES := 1
	IMAGESHA256 := true
else
	IMAGESHA256 := false
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
MANIFESTPATHSTORAGE := manifests/arango-storage$(MANIFESTSUFFIX).yaml
MANIFESTPATHTEST := manifests/arango-test$(MANIFESTSUFFIX).yaml
ifndef DEPLOYMENTNAMESPACE
	DEPLOYMENTNAMESPACE := default
endif

ifndef OPERATORIMAGE
	OPERATORIMAGE := $(DOCKERNAMESPACE)/kube-arangodb$(IMAGESUFFIX)
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

# Magical rubbish to teach make what commas and spaces are.
EMPTY :=
SPACE := $(EMPTY) $(EMPTY)
COMMA := $(EMPTY),$(EMPTY)

ARCHS:=amd64 arm arm64
PLATFORMS:=$(subst $(SPACE),$(COMMA),$(foreach arch,$(ARCHS),linux/$(arch)))

BINNAME := $(PROJECT)
BIN := $(foreach arch,$(ARCHS),$(BINDIR)/linux/$(arch)/$(BINNAME))
TESTBINNAME := $(PROJECT)_test
TESTBIN := $(BINDIR)/$(TESTBINNAME)
DURATIONTESTBINNAME := $(PROJECT)_duration_test
DURATIONTESTBIN := $(BINDIR)/$(DURATIONTESTBINNAME)
RELEASE := $(GOBUILDDIR)/bin/release 
GHRELEASE := $(GOBUILDDIR)/bin/github-release 

TESTLENGTHOPTIONS := -test.short
TESTTIMEOUT := 20m
ifeq ($(LONG), 1)
	TESTLENGTHOPTIONS :=
	TESTTIMEOUT := 180m
endif
ifdef VERBOSE
	TESTVERBOSEOPTIONS := -v 
endif

SOURCES := $(shell find $(SRCDIR) -name '*.go' -not -path './test/*')
DASHBOARDSOURCES := $(shell find $(DASHBOARDDIR)/src -name '*.js' -not -path './test/*') $(DASHBOARDDIR)/package.json

.PHONY: all
all: verify-generated build

#
# Tip: Run `eval $(minikube docker-env)` before calling make if you're developing on minikube.
#

.PHONY: build
build: check-vars build-bin docker manifests

.PHONY: clean
clean:
	rm -Rf $(BIN) $(BINDIR) $(GOBUILDDIR) $(DASHBOARDDIR)/build $(DASHBOARDDIR)/node_modules

.PHONY: check-vars
check-vars:
ifndef DOCKERNAMESPACE
	@echo "DOCKERNAMESPACE must be set"
	@exit 1
endif
	@echo "Using docker namespace: $(DOCKERNAMESPACE)"

.PHONY: deps
deps:
	@${MAKE} -B -s $(GOBUILDDIR)

$(GOBUILDDIR):
	# Build pulsar & go-assets-builder from vendor
	@mkdir -p $(GOBUILDDIR)
	@ln -sf $(VENDORDIR) $(GOBUILDDIR)/src
	@GOPATH=$(GOBUILDDIR) go install github.com/pulcy/pulsar
	@GOPATH=$(GOBUILDDIR) go install github.com/jessevdk/go-assets-builder
	@rm -Rf $(GOBUILDDIR)/src
	# Prepare .gobuild directory
	@mkdir -p $(ORGDIR)
	@rm -f $(REPODIR) && ln -sf ../../../.. $(REPODIR)
	GOPATH=$(GOBUILDDIR) $(PULSAR) go flatten -V $(VENDORDIR)
	# Note: Next library is not vendored, since we always want the latest version
	GOPATH=$(GOBUILDDIR) go get github.com/arangodb/go-upgrade-rules

$(CACHEVOL):
	@docker volume create $(CACHEVOL)

.PHONY: update-vendor
update-vendor:
	@mkdir -p $(GOBUILDDIR)
	@GOPATH=$(GOBUILDDIR) go get github.com/pulcy/pulsar
	@rm -Rf $(VENDORDIR)
	@mkdir -p $(VENDORDIR)
	@git clone https://github.com/kubernetes/code-generator.git $(VENDORDIR)/k8s.io/code-generator
	@rm -Rf $(VENDORDIR)/k8s.io/code-generator/.git
	@$(PULSAR) go vendor -V $(VENDORDIR) \
		k8s.io/client-go/... \
		k8s.io/gengo/args \
		k8s.io/apiextensions-apiserver \
		github.com/aktau/github-release \
		github.com/arangodb-helper/go-certificates \
		github.com/arangodb/go-driver \
		github.com/cenkalti/backoff \
		github.com/coreos/go-semver/semver \
		github.com/dchest/uniuri \
		github.com/dgrijalva/jwt-go \
		github.com/gin-gonic/gin \
		github.com/jessevdk/go-assets-builder \
		github.com/julienschmidt/httprouter \
		github.com/pkg/errors \
		github.com/prometheus/client_golang/prometheus \
		github.com/pulcy/pulsar \
		github.com/rs/zerolog \
		github.com/spf13/cobra \
		github.com/mitchellh/gox \
		github.com/estesp/manifest-tool \
		github.com/arangodb-helper/glog \
		github.com/stretchr/testify
	@$(PULSAR) go flatten -V $(VENDORDIR) $(VENDORDIR)
	@${MAKE} -B -s clean
	# Manually restore arangosync vendor with: git checkout deps/github.com/arangodb/arangosync

.PHONY: update-generated
update-generated: $(GOBUILDDIR) 
	@docker build $(SRCDIR)/tools/codegen --build-arg GOVERSION=$(GOVERSION) -t k8s-codegen
	docker run \
		--rm \
		-v $(SRCDIR):/usr/code \
		-e GOPATH=/usr/code/.gobuild \
		-e GOBIN=/usr/code/.gobuild/bin \
		-w /usr/code/ \
		k8s-codegen \
		"./deps/k8s.io/code-generator/generate-groups.sh"  \
		"all" \
		"github.com/arangodb/kube-arangodb/pkg/generated" \
		"github.com/arangodb/kube-arangodb/pkg/apis" \
		"deployment:v1alpha replication:v1alpha storage:v1alpha" \
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
	$(GOASSETSBUILDER) -s /dashboard/build/ -o dashboard/assets.go -p dashboard dashboard/build

.PHONY: build-bin
build-bin: $(GOX) $(GOBUILDDIR) $(CACHEVOL) $(SOURCES) dashboard/assets.go
	@mkdir -p $(BINDIR)
	CGO_ENABLED=0 GOPATH=$(GOBUILDDIR) $(GOX) \
		-os="linux" \
		-arch="$(ARCHS)" \
		-osarch="!darwin/arm !darwin/arm64" \
		-ldflags="-X main.projectVersion=${VERSION} -X main.projectBuild=${COMMIT}" \
		-output="bin/{{.OS}}/{{.Arch}}/$(BINNAME)" \
		-tags="netgo" \
		github.com/arangodb/kube-arangodb
	@ln -sf $(BINDIR)/$(shell go env GOOS)/$(shell go env GOARCH)/kube-arangodb$(shell go env GOEXE)



.PHONY: docker
docker: check-vars build-bin $(MANIFESTTOOL)
	for arch in $(ARCHS); do \
		docker build --build-arg=GOARCH=$$arch -t $(OPERATORIMAGE)-$$arch . ;\
		docker push $(OPERATORIMAGE)-$$arch ;\
	done
	$(MANIFESTTOOL) $(MANIFESTAUTH) push from-args \
    	--platforms $(PLATFORMS) \
    	--template $(OPERATORIMAGE)-ARCH \
    	--target $(OPERATORIMAGE)


$(GOX):  $(GOBUILDDIR)
	GOPATH=$(GOBUILDDIR) go build -o $(GOX) github.com/mitchellh/gox

$(MANIFESTTOOL): 
	GOPATH=$(GOBUILDDIR) go build -o $(MANIFESTTOOL) github.com/estesp/manifest-tool

# Manifests 

.PHONY: manifests
manifests: $(GOBUILDDIR)
	@echo Building manifests
	GOPATH=$(GOBUILDDIR) go run $(ROOTDIR)/tools/manifests/manifest_builder.go \
		--output-suffix=$(MANIFESTSUFFIX) \
		--image=$(OPERATORIMAGE) \
		--image-sha256=$(IMAGESHA256) \
		--namespace=$(DEPLOYMENTNAMESPACE) \
		--allow-chaos=$(ALLOWCHAOS)

# Testing

.PHONY: run-unit-tests
run-unit-tests: $(GOBUILDDIR) $(SOURCES)
	docker run \
		--rm \
		-v $(SRCDIR):/usr/code \
		-v $(CACHEVOL):/usr/gocache \
		-e GOCACHE=/usr/gocache \
		-e GOPATH=/usr/code/.gobuild \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=0 \
		-w /usr/code/ \
		golang:$(GOVERSION) \
		go test $(TESTVERBOSEOPTIONS) \
			$(REPOPATH)/pkg/apis/deployment/v1alpha \
			$(REPOPATH)/pkg/apis/replication/v1alpha \
			$(REPOPATH)/pkg/apis/storage/v1alpha \
			$(REPOPATH)/pkg/deployment/reconcile \
			$(REPOPATH)/pkg/deployment/resources \
			$(REPOPATH)/pkg/storage \
			$(REPOPATH)/pkg/util/k8sutil \
			$(REPOPATH)/pkg/util/k8sutil/test \
			$(REPOPATH)/pkg/util/probe \
			$(REPOPATH)/pkg/util/validation 

$(TESTBIN): $(GOBUILDDIR) $(SOURCES)
	@mkdir -p $(BINDIR)
	docker run \
		--rm \
		-v $(SRCDIR):/usr/code \
		-v $(CACHEVOL):/usr/gocache \
		-e GOCACHE=/usr/gocache \
		-e GOPATH=/usr/code/.gobuild \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=0 \
		-w /usr/code/ \
		golang:$(GOVERSION) \
		go test -c -installsuffix cgo -ldflags "-X main.projectVersion=$(VERSION) -X main.projectBuild=$(COMMIT)" -o /usr/code/bin/$(TESTBINNAME) $(REPOPATH)/tests

.PHONY: docker-test
docker-test: $(TESTBIN)
	docker build --quiet -f $(DOCKERTESTFILE) -t $(TESTIMAGE) .

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
	kubectl apply -f $(MANIFESTPATHTEST)
	$(ROOTDIR)/scripts/kube_create_storage.sh $(DEPLOYMENTNAMESPACE)
	$(ROOTDIR)/scripts/kube_create_license_key_secret.sh "$(DEPLOYMENTNAMESPACE)" '$(ENTERPRISELICENSE)'
	$(ROOTDIR)/scripts/kube_run_tests.sh $(DEPLOYMENTNAMESPACE) $(TESTIMAGE) "$(ARANGODIMAGE)" '$(ENTERPRISEIMAGE)' $(TESTTIMEOUT) $(TESTLENGTHOPTIONS)

$(DURATIONTESTBIN): $(GOBUILDDIR) $(SOURCES)
	@mkdir -p $(BINDIR)
	docker run \
		--rm \
		-v $(SRCDIR):/usr/code \
		-v $(CACHEVOL):/usr/gocache \
		-e GOCACHE=/usr/gocache \
		-e GOPATH=/usr/code/.gobuild \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=0 \
		-w /usr/code/ \
		golang:$(GOVERSION) \
		go build -installsuffix cgo -ldflags "-X main.projectVersion=$(VERSION) -X main.projectBuild=$(COMMIT)" -o /usr/code/bin/$(DURATIONTESTBINNAME) $(REPOPATH)/tests/duration

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
	kubectl delete -f $(MANIFESTPATHSTORAGE) --ignore-not-found
	kubectl delete -f $(MANIFESTPATHCRD) --ignore-not-found

.PHONY: redeploy-operator
redeploy-operator: delete-operator manifests
	kubectl apply -f $(MANIFESTPATHCRD)
	kubectl apply -f $(MANIFESTPATHSTORAGE)
	kubectl apply -f $(MANIFESTPATHDEPLOYMENT)
	kubectl apply -f $(MANIFESTPATHDEPLOYMENTREPLICATION)
	kubectl apply -f $(MANIFESTPATHTEST)
	kubectl get pods 
