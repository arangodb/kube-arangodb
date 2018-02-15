GOPATH=$(shell pwd)/.gobuild
ORGDIR=$(GOPATH)/src/github.com/bugagazavr
REPODIR=$(ORGDIR)/go-gitlab-client

all:deps test

clean:
	rm -Rf $(GOPATH)

deps:
	@mkdir -p $(ORGDIR)
	@rm -f $(REPODIR) && ln -s ../../../.. $(REPODIR)
	GOPATH=$(GOPATH) go get github.com/stretchr/testify
	GOPATH=$(GOPATH) go get $(REPODIR)

test:
	GOPATH=$(GOPATH) cd $(REPODIR) && go test -short ./...
