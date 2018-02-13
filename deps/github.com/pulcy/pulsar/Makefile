PROJECT := pulsar
SCRIPTDIR := $(shell pwd)
VERSION:= $(shell cat VERSION)
COMMIT := $(shell git rev-parse --short HEAD)

GOBUILDDIR := $(SCRIPTDIR)/.gobuild
SRCDIR := $(SCRIPTDIR)
BINDIR := $(SCRIPTDIR)
VENDORDIR = $(SCRIPTDIR)/vendor

ORGPATH := github.com/pulcy
ORGDIR := $(GOBUILDDIR)/src/$(ORGPATH)
REPONAME := $(PROJECT)
REPODIR := $(ORGDIR)/$(REPONAME)
REPOPATH := $(ORGPATH)/$(REPONAME)
BIN := $(BINDIR)/$(PROJECT)

GOPATH := $(GOBUILDDIR)
GOVERSION := 1.9.0-alpine

SOURCES := $(shell find $(SRCDIR) -name '*.go')

ifndef GOOS
	GOOS := $(shell go env GOOS)
endif
ifndef GOARCH
	GOARCH := $(shell go env GOARCH)
endif


.PHONY: clean test release tgz

all: $(BIN)

clean:
	rm -Rf $(BIN) $(GOBUILDDIR) bin

$(GOBUILDDIR):
	mkdir -p $(ORGDIR)
	rm -f $(REPODIR) && ln -s ../../../../ $(REPODIR)

update-vendor:
	rm -Rf $(VENDORDIR)
	$(BIN) go vendor -V $(VENDORDIR) \
		github.com/coreos/go-semver/semver \
		github.com/cpuguy83/go-md2man \
		github.com/ewoutp/go-gitlab-client \
		github.com/inconshreveable/mousetrap \
		github.com/juju/errgo \
		github.com/mgutz/ansi \
		github.com/mitchellh/go-homedir \
		github.com/op/go-logging \
		github.com/russross/blackfriday \
		github.com/shurcooL/sanitized_anchor_name \
		github.com/sourcegraph/go-vcsurl \ 
		github.com/spf13/cobra \
		github.com/spf13/pflag

$(BIN): $(GOBUILDDIR) $(SOURCES)
	docker run \
	    --rm \
	    -v $(SRCDIR):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -e GOOS=$(GOOS) \
	    -e GOARCH=$(GOARCH) \
		-e CGO_ENABLED=0 \
	    -w /usr/code/ \
	    golang:$(GOVERSION) \
	    go build -a -installsuffix netgo -tags netgo -ldflags "-X main.projectVersion=$(VERSION) -X main.projectBuild=$(COMMIT)" -o /usr/code/$(PROJECT) $(REPOPATH)

release:
	@${MAKE} -B GOOS=linux GOARCH=amd64 tgz
	@${MAKE} -B GOOS=darwin GOARCH=amd64 tgz

tgz: $(BIN)
	mkdir -p bin
	tar zcf bin/$(PROJECT)-$(GOOS)-$(GOARCH).tgz -C $(SCRIPTDIR) $(PROJECT)
	rm $(BIN)

test:
	#GOPATH=$(GOPATH) go test -v $(REPOPATH)/scheduler
