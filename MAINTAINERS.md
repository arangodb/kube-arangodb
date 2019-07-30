# Maintainer Instructions

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
