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
BINDIR := $(ROOTDIR)/bin
VENDORDIR := $(ROOTDIR)/deps

ORGPATH := github.com/arangodb
ORGDIR := $(GOBUILDDIR)/src/$(ORGPATH)
REPONAME := kube-arangodb
REPODIR := $(ORGDIR)/$(REPONAME)
REPOPATH := $(ORGPATH)/$(REPONAME)

GOPATH := $(GOBUILDDIR)
GOVERSION := 1.10.0-alpine

PULSAR := $(GOBUILDDIR)/bin/pulsar$(shell go env GOEXE)

DOCKERFILE := Dockerfile 
DOCKERTESTFILE := Dockerfile.test

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

ifndef MANIFESTPATH 
	MANIFESTPATH := manifests/arango-operator-dev.yaml
endif
ifndef DEPLOYMENTNAMESPACE
	DEPLOYMENTNAMESPACE := default
endif

ifndef OPERATORIMAGE
	OPERATORIMAGE := $(DOCKERNAMESPACE)/arangodb-operator$(IMAGESUFFIX)
endif
ifndef TESTIMAGE
	TESTIMAGE := $(DOCKERNAMESPACE)/arangodb-operator-test$(IMAGESUFFIX)
endif
ifndef ENTERPRISEIMAGE
	ENTERPRISEIMAGE := $(DEFAULTENTERPRISEIMAGE)
endif

BINNAME := $(PROJECT)
BIN := $(BINDIR)/$(BINNAME)
TESTBINNAME := $(PROJECT)_test
TESTBIN := $(BINDIR)/$(TESTBINNAME)
RELEASE := $(GOBUILDDIR)/bin/release 
GHRELEASE := $(GOBUILDDIR)/bin/github-release 

TESTLENGTHOPTIONS := -test.short
TESTTIMEOUT := 20m
ifeq ($(LONG), 1)
	TESTLENGTHOPTIONS :=
	TESTTIMEOUT := 40m
endif
ifdef VERBOSE
	TESTVERBOSEOPTIONS := -v 
endif

SOURCES := $(shell find $(SRCDIR) -name '*.go' -not -path './test/*')

.PHONY: all
all: verify-generated build

#
# Tip: Run `eval $(minikube docker-env)` before calling make if you're developing on minikube.
#

.PHONY: build
build: check-vars docker manifests

.PHONY: clean
clean:
	rm -Rf $(BIN) $(BINDIR) $(GOBUILDDIR)

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
	# Build pulsar from vendor
	@mkdir -p $(GOBUILDDIR)
	@ln -sf $(VENDORDIR) $(GOBUILDDIR)/src
	@GOPATH=$(GOBUILDDIR) go install github.com/pulcy/pulsar
	@rm -Rf $(GOBUILDDIR)/src
	# Prepare .gobuild directory
	@mkdir -p $(ORGDIR)
	@rm -f $(REPODIR) && ln -sf ../../../.. $(REPODIR)
	GOPATH=$(GOBUILDDIR) $(PULSAR) go flatten -V $(VENDORDIR)

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
		github.com/julienschmidt/httprouter \
		github.com/pkg/errors \
		github.com/prometheus/client_golang/prometheus \
		github.com/pulcy/pulsar \
		github.com/rs/zerolog \
		github.com/spf13/cobra \
		github.com/stretchr/testify
	@$(PULSAR) go flatten -V $(VENDORDIR) $(VENDORDIR)
	@${MAKE} -B -s clean

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
		"deployment:v1alpha storage:v1alpha" \
		--go-header-file "./tools/codegen/boilerplate.go.txt" \
		$(VERIFYARGS)

.PHONY: verify-generated
verify-generated:
	@${MAKE} -B -s VERIFYARGS=--verify-only update-generated

$(BIN): $(GOBUILDDIR) $(SOURCES)
	@mkdir -p $(BINDIR)
	docker run \
		--rm \
		-v $(SRCDIR):/usr/code \
		-e GOPATH=/usr/code/.gobuild \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=0 \
		-w /usr/code/ \
		golang:$(GOVERSION) \
		go build -installsuffix cgo -ldflags "-X main.projectVersion=$(VERSION) -X main.projectBuild=$(COMMIT)" -o /usr/code/bin/$(BINNAME) $(REPOPATH)

.PHONY: docker
docker: check-vars $(BIN)
	docker build -f $(DOCKERFILE) -t $(OPERATORIMAGE) .
ifdef PUSHIMAGES
	docker push $(OPERATORIMAGE)
endif

# Manifests 

.PHONY: manifests
manifests: $(GOBUILDDIR)
	GOPATH=$(GOBUILDDIR) go run $(ROOTDIR)/tools/manifests/manifest_builder.go \
		--output=$(MANIFESTPATH) \
		--image=$(OPERATORIMAGE) \
		--image-sha256=$(IMAGESHA256) \
		--namespace=$(DEPLOYMENTNAMESPACE)

# Testing

.PHONY: run-unit-tests
run-unit-tests: $(GOBUILDDIR) $(SOURCES)
	docker run \
		--rm \
		-v $(SRCDIR):/usr/code \
		-e GOPATH=/usr/code/.gobuild \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=0 \
		-w /usr/code/ \
		golang:$(GOVERSION) \
		go test $(TESTVERBOSEOPTIONS) \
			$(REPOPATH)/pkg/apis/deployment/v1alpha \
			$(REPOPATH)/pkg/deployment \
			$(REPOPATH)/pkg/util/k8sutil \
			$(REPOPATH)/pkg/util/k8sutil/test

$(TESTBIN): $(GOBUILDDIR) $(SOURCES)
	@mkdir -p $(BINDIR)
	docker run \
		--rm \
		-v $(SRCDIR):/usr/code \
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
	kubectl apply -f $(MANIFESTPATH)
	$(ROOTDIR)/scripts/kube_create_storage.sh $(DEPLOYMENTNAMESPACE)
	kubectl --namespace $(DEPLOYMENTNAMESPACE) \
		run arangodb-operator-test -i --rm --quiet --restart=Never \
		--image=$(TESTIMAGE) \
		--env="ENTERPRISEIMAGE=$(ENTERPRISEIMAGE)" \
		--env="TEST_NAMESPACE=$(DEPLOYMENTNAMESPACE)" \
		-- \
		-test.v -test.timeout $(TESTTIMEOUT) $(TESTLENGTHOPTIONS)
ifneq ($(DEPLOYMENTNAMESPACE), default)
	kubectl delete namespace $(DEPLOYMENTNAMESPACE) --ignore-not-found --now
endif

.PHONY: cleanup-tests
cleanup-tests:
ifneq ($(DEPLOYMENTNAMESPACE), default)
	$(ROOTDIR)/scripts/kube_delete_namespace.sh $(DEPLOYMENTNAMESPACE)
endif

# Release building

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
	kubectl delete -f $(MANIFESTPATH) --ignore-not-found

.PHONY: redeploy-operator
redeploy-operator: delete-operator manifests
	kubectl apply -f $(MANIFESTPATH)
	kubectl get pods 
