PROJECT := go-velocypack
SCRIPTDIR := $(shell pwd)
ROOTDIR := $(shell cd $(SCRIPTDIR) && pwd)

GOBUILDDIR := $(SCRIPTDIR)/.gobuild
GOVERSION := 1.7.5-alpine

TESTOPTIONS := 
ifdef VERBOSE
	TESTOPTIONS := -v
endif

ORGPATH := github.com/arangodb
ORGDIR := $(GOBUILDDIR)/src/$(ORGPATH)
REPONAME := $(PROJECT)
REPODIR := $(ORGDIR)/$(REPONAME)
REPOPATH := $(ORGPATH)/$(REPONAME)

SOURCES := $(shell find . -name '*.go')

.PHONY: all build clean run-tests show-coverage

all: build

build: $(GOBUILDDIR) $(SOURCES)
	GOPATH=$(GOBUILDDIR) go build -v github.com/arangodb/go-velocypack 

clean:
	rm -Rf $(GOBUILDDIR)

$(GOBUILDDIR):
	@mkdir -p $(ORGDIR)
	@rm -f $(REPODIR) && ln -s ../../../.. $(REPODIR)

# All unit tests
run-tests: $(GOBUILDDIR)
	@GOPATH=$(GOBUILDDIR) go get github.com/stretchr/testify/assert
	@docker run \
		--rm \
		-v $(ROOTDIR):/usr/code \
		-e GOPATH=/usr/code/.gobuild \
		-w /usr/code/ \
		golang:$(GOVERSION) \
		sh -c "go test -v $(REPOPATH)/test/runtime && go test $(TESTOPTIONS) $(REPOPATH) && go test -cover -coverpkg $(REPOPATH) -coverprofile=coverage.out $(TESTOPTIONS) $(REPOPATH)/test"

# All benchmarks
run-benchmarks: $(GOBUILDDIR)
	@GOPATH=$(GOBUILDDIR) go get github.com/stretchr/testify/assert
	@docker run \
		--rm \
		-v $(ROOTDIR):/usr/code \
		-e GOPATH=/usr/code/.gobuild \
		-w /usr/code/ \
		golang:$(GOVERSION) \
		go test $(TESTOPTIONS) -bench=. -run=notests -cpu=1,2,4 $(REPOPATH)/test

# All benchmarks using local profiling
run-benchmarks-prof: $(GOBUILDDIR)
	@GOPATH=$(GOBUILDDIR) go get github.com/stretchr/testify/assert
	@GOPATH=$(GOBUILDDIR) go test $(TESTOPTIONS) -bench=. -run=notests -cpu=1,2,4 -cpuprofile=cpu.out $(REPOPATH)/test
	@echo Now profile using: go tool pprof test.test cpu.out

# All unit tests using local Go tools
run-tests-local: $(GOBUILDDIR)
	@GOPATH=$(GOBUILDDIR) go get github.com/stretchr/testify/assert
	@GOPATH=$(GOBUILDDIR) go test -v $(REPOPATH)/test/runtime
	@GOPATH=$(GOBUILDDIR) go test $(TESTOPTIONS) $(REPOPATH) 
	@GOPATH=$(GOBUILDDIR) go test -cover -coverpkg $(REPOPATH) -coverprofile=coverage.out $(TESTOPTIONS) $(REPOPATH)/test

# All (except large memory) unit tests using local Go tools
run-tests-local-nolarge: $(GOBUILDDIR)
	@GOPATH=$(GOBUILDDIR) go get github.com/stretchr/testify/assert
	@GOPATH=$(GOBUILDDIR) go test -tags nolarge -v $(REPOPATH)/test/runtime
	@GOPATH=$(GOBUILDDIR) go test -tags nolarge $(TESTOPTIONS) $(REPOPATH) 
	@GOPATH=$(GOBUILDDIR) go test -tags nolarge -cover -coverpkg $(REPOPATH) -coverprofile=coverage.out $(TESTOPTIONS) $(REPOPATH)/test

show-coverage: run-tests
	go tool cover -html coverage.out 
