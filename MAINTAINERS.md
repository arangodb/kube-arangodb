# Maintainer Instructions

## Before

To run templating models HELM needs to be installed. We are supporting HELM 2.14+

Helm URL: https://github.com/helm/helm/releases/tag/v2.14.3

You can install HELM in PATH or set HELM path before running make - `export HELM=<path to helm>`

## Running tests

To run the entire test set, first set the following environment variables:

  - `DOCKERNAMESPACE` to your docker hub account
  - `VERBOSE` to `1` (default is empty)
  - `LONG` to `1` (default is empty, which skips lots of tests)
  - `ARANGODB` to the name of a community image you want to test,
    default is `arangodb/arangodb:latest`
  - `ENTERPRISEIMAGE` to the name of an enterprise image, you want to
    test, if not set, some tests are skipped
  - `ENTERPRISELICENSE` to the enterpise license key
  - `KUBECONFIG` to the path to some k8s configuration with
    credentials, this indicates which cluster to use

```bash
make clean
make build
make run-tests
```

To run only a single test, set `TESTOPTIONS` to something like
`-test.run=TestRocksDBEncryptionSingle` where
`TestRocksDBEncryptionSingle` is the name of the test.

## Debugging with DLV

To attach DLV debugger, first prepare operator image with DLV server included:
```shell
IMAGETAG=1.2.4dlv DEBUG=true make docker
```

Then deploy it on your k8s and use following command to access DLV server on `localhost:2345` from your local machine:
```shell
kubectl port-forward deployment/arango-arango-deployment-operator 2345
```

## Preparing a release

To prepare for a release, do the following:

- Make sure all tests are OK.
- To run a complete set of tests, see above.
- Update the CHANGELOG manually, since the automatic CHANGELOG
  generation is switched off (did not work in many cases).

## Building a release

To make a release you must have:

- A github access token in `~/.arangodb/github-token` that has read/write access
  for this repository.
- Push permission for the current docker account (`docker login <your-docker-hub-account>`)
  for the `arangodb` docker hub namespace.
- The latest checked out `master` branch of this repository.

```bash
make release-patch
# or
make release-minor
# or
make release-major
```

If successful, a new version will be:

- Build docker images, yaml resources & helm charts.
- Tagged in github
- Uploaded as github release
- Pushed as docker image to docker hub
- `./VERSION` will be updated to a `+git` version (after the release process)

If the release process fails, it may leave:

- `./VERSION` uncommitted. To resolve, checkout `master` or edit it to
  the original value and commit to master.
- A git tag named `<major>.<minor>.<patch>` in your repository.
  To resolve remove it using `git tag -d ...`.
- A git tag named `<major>.<minor>.<patch>` in this repository in github.
  To resolve remove it manually.

## Development on MacOS

This repo requires GNU command line tools instead BSD one (which are by default available on Mac).
```shell
brew install coreutils ed findutils gawk gnu-sed gnu-tar grep make
```

Please add following to your `~/bashrc` or `~/.zshrc` file (it requires Homebrew to be installed):

```shell
HOMEBREW_PREFIX=$(brew --prefix)
for d in ${HOMEBREW_PREFIX}/opt/*/libexec/gnubin; do export PATH=$d:$PATH; done
```

## Change Go version
#### Change file Makefile
* GOVERSION := e.g. 1.17-alpine3.15
* DISTRIBUTION := e.g. alpine:3.15
#### Change file .travis.yml
#### Change file go.mod
