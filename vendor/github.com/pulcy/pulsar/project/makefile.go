// Copyright (c) 2016 Pulcy.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package project

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/op/go-logging"
)

const (
	makefilePath    = "Makefile"
	initialMakefile = `PROJECT := $(shell pulsar project name)
SCRIPTDIR := $(shell pwd)
ROOTDIR := $(shell cd $(SCRIPTDIR) && pwd)
VERSION:= $(shell pulsar project version)
COMMIT := $(shell pulsar project commit)

GOBUILDDIR := $(SCRIPTDIR)/.gobuild
SRCDIR := $(SCRIPTDIR)
BINDIR := $(ROOTDIR)/bin
VENDORDIR := $(SCRIPTDIR)/vendor

ORGPATH := $(shell pulsar project organization path)
ORGDIR := $(GOBUILDDIR)/src/$(ORGPATH)
REPONAME := $(PROJECT)
REPODIR := $(ORGDIR)/$(REPONAME)
REPOPATH := $(ORGPATH)/$(REPONAME)

GOPATH := $(GOBUILDDIR)
GOVERSION := {{.GoVersion}}

ifndef GOOS
	GOOS := $(shell go env GOHOSTOS)
endif
ifndef GOARCH
	GOARCH := $(shell go env GOHOSTARCH)
endif

BINNAME := $(PROJECT)
BIN := $(BINNAME)

SOURCES := $(shell find $(SRCDIR) -name '*.go')

.PHONY: all build clean deps update-vendor 

all: build    

build: $(BIN) 

clean:
	rm -Rf $(BIN) $(GOBUILDDIR)

deps:
	@${MAKE} -B -s $(GOBUILDDIR)

$(GOBUILDDIR):
	@mkdir -p $(ORGDIR)
	@mkdir -p $(VENDORDIR)
	@rm -f $(REPODIR) && ln -s ../../../.. $(REPODIR)
	@GOPATH=$(GOPATH) pulsar go flatten -V $(VENDORDIR)

update-vendor:
	@rm -Rf $(VENDORDIR)
	@pulsar go vendor -V $(VENDORDIR) \
		github.com/juju/errgo \
		github.com/op/go-logging
	@${MAKE} -B clean

$(BIN): $(GOBUILDDIR) $(SOURCES)
	@mkdir -p $(BINDIR)
	docker run \
		--rm \
		-v $(ROOTDIR):/usr/code \
		-e GOPATH=/usr/code/.gobuild \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		-e CGO_ENABLED=0 \
		-w /usr/code/ \
		golang:$(GOVERSION) \
		go build -a -installsuffix netgo -tags netgo -ldflags "-X main.projectVersion=$(VERSION) -X main.projectBuild=$(COMMIT)" -o /usr/code/$(BINNAME) $(REPOPATH)
`
)

func initMakefile(log *log.Logger, projectDir, projectType string) error {
	if projectType != ProjectTypeGo {
		return nil
	}
	path := filepath.Join(projectDir, makefilePath)
	if info, err := os.Stat(path); os.IsNotExist(err) {
		log.Infof("Creating %s", makefilePath)
		options := struct {
			GoVersion string
		}{
			GoVersion: "1.7.3-alpine",
		}
		t, err := template.New("makefile").Parse(initialMakefile)
		if err != nil {
			return maskAny(err)
		}
		buffer := &bytes.Buffer{}
		if err := t.Execute(buffer, options); err != nil {
			return maskAny(err)
		}

		if err := ioutil.WriteFile(path, buffer.Bytes(), 0644); err != nil {
			return maskAny(err)
		}
		return nil
	} else if err != nil {
		return maskAny(err)
	} else if info.IsDir() {
		return maskAny(fmt.Errorf("%s must be a file", path))
	} else {
		log.Debugf("%s already initialized in %s", gitIgnorePath, projectDir)
		return nil
	}
}
