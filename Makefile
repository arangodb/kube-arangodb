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
REPONAME := k8s-operator
REPODIR := $(ORGDIR)/$(REPONAME)
REPOPATH := $(ORGPATH)/$(REPONAME)

GOPATH := $(GOBUILDDIR)
GOVERSION := 1.10.0-alpine

PULSAR := $(GOBUILDDIR)/bin/pulsar$(shell go env GOEXE)

ifndef DOCKERNAMESPACE
	DOCKERNAMESPACE := arangodb
endif
ifndef DOCKERFILE
	DOCKERFILE := Dockerfile 
	#DOCKERFILE := Dockerfile.debug
endif

BINNAME := $(PROJECT)
BIN := $(BINDIR)/$(BINNAME)
RELEASE := $(GOBUILDDIR)/bin/release 
GHRELEASE := $(GOBUILDDIR)/bin/github-release 

SOURCES := $(shell find $(SRCDIR) -name '*.go' -not -path './test/*')

.PHONY: all clean deps docker update-vendor update-generated verify-generated

all: verify-generated docker

#
# Tip: Run `eval $(minikube docker-env)` before calling make if you're developing on minikube.
#

build: docker

clean:
	rm -Rf $(BIN) $(BINDIR) $(GOBUILDDIR)

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
		github.com/cenkalti/backoff \
		github.com/dchest/uniuri \
		github.com/pkg/errors \
		github.com/prometheus/client_golang/prometheus \
		github.com/pulcy/pulsar \
		github.com/rs/zerolog \
		github.com/spf13/cobra
	@$(PULSAR) go flatten -V $(VENDORDIR) $(VENDORDIR)
	@${MAKE} -B -s clean

update-generated: $(GOBUILDDIR)
	@sed 's/GOVERSION/$(GOVERSION)/' $(SRCDIR)/tools/codegen/Dockerfile.in > $(SRCDIR)/tools/codegen/Dockerfile
	@docker build $(SRCDIR)/tools/codegen -t k8s-codegen
	docker run \
		--rm \
		-v $(SRCDIR):/usr/code \
		-e GOPATH=/usr/code/.gobuild \
		-e GOBIN=/usr/code/.gobuild/bin \
		-w /usr/code/ \
		k8s-codegen \
		"./deps/k8s.io/code-generator/generate-groups.sh"  \
		"all" \
		"github.com/arangodb/k8s-operator/pkg/generated" \
		"github.com/arangodb/k8s-operator/pkg/apis" \
		"arangodb:v1alpha" \
		--go-header-file "./tools/codegen/boilerplate.go.txt" \
		$(VERIFYARGS)

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

docker: $(BIN)
	docker build -f $(DOCKERFILE) -t arangodb/arangodb-operator .

docker-push: docker
ifneq ($(DOCKERNAMESPACE), arangodb)
	docker tag arangodb/arangodb-operator $(DOCKERNAMESPACE)/arangodb-operator
endif
	docker push $(DOCKERNAMESPACE)/arangodb-operator

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

$(GHRELEASE): $(GOBUILDDIR) 
	GOPATH=$(GOBUILDDIR) go build -o $(GHRELEASE) github.com/aktau/github-release

release-patch: $(RELEASE)
	GOPATH=$(GOBUILDDIR) $(RELEASE) -type=patch 

release-minor: $(RELEASE)
	GOPATH=$(GOBUILDDIR) $(RELEASE) -type=minor

release-major: $(RELEASE)
	GOPATH=$(GOBUILDDIR) $(RELEASE) -type=major 

## Kubernetes utilities

minikube-start:
	minikube start --cpus=4 --memory=6144

delete-operator:
	kubectl delete -f examples/deployment.yaml || true

redeploy-operator: delete-operator
	kubectl create -f examples/deployment.yaml
	kubectl get pods 
