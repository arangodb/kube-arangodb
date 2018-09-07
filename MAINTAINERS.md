# Maintainer Instructions

## Running tests

To run the entire test set, run:

```bash
export DOCKERNAMESPACE=<your docker hub account>
make clean
make build
make run-tests
```

## Preparing a release

To prepare for a release, do the following:

- Make sure all tests are OK.

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
