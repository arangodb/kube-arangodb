#!/usr/bin/env bash
set -e

# This script builds various binary from a checkout of the manifest-tool
# source code.
#
# Requirements:
# - The current directory should be a checkout of the manifest-tool source code
#   (https://github.com/estesp/manifest-tool). Whatever version is checked out
#   will be built.
# - The script is intended to be run inside the docker container specified
#   in the Dockerfile at the root of the source. In other words:
#   DO NOT CALL THIS SCRIPT DIRECTLY.
# - The right way to call this script is to invoke "make" from
#   your checkout of the manifest-tool repository.
#   the Makefile will do a "docker build -t manifest-tool ." and then
#   "docker run hack/make.sh" in the resulting image.
#

set -o pipefail

export MANIFEST_PKG='github.com/estesp/manifest-tool'
export SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export MAKEDIR="$SCRIPTDIR/make"

# We're a nice, sexy, little shell script, and people might try to run us;
# but really, they shouldn't. We want to be in a container!
inContainer="AssumeSoInitially"
if [ "$PWD" != "/go/src/$MANIFEST_PKG" ]; then
	unset inContainer
fi

if [ -z "$inContainer" ]; then
	{
		echo "# WARNING! I don't seem to be running in a Docker container."
		echo "# The result of this command might be an incorrect build, and will not be"
		echo "# officially supported."
		echo "#"
		echo "# Try this instead: make all"
		echo "#"
	} >&2
fi

echo

DEFAULT_BUNDLES=(
	validate-gofmt
	validate-lint
	validate-vet
	validate-git-marks
)

TESTFLAGS+=" -test.timeout=10m"

# If $TESTFLAGS is set in the environment, it is passed as extra arguments to 'go test'.
# You can use this to select certain tests to run, eg.
#
#     TESTFLAGS='-test.run ^TestBuild$' ./hack/make.sh test-unit
#
# For integration-cli test, we use [gocheck](https://labix.org/gocheck), if you want
# to run certain tests on your local host, you should run with command:
#
#     TESTFLAGS='-check.f DockerSuite.TestBuild*' ./hack/make.sh binary test-integration-cli
#
go_test_dir() {
	dir=$1
	(
		echo '+ go test' $TESTFLAGS "${MANIFEST_PKG}${dir#.}"
		cd "$dir"
		export DEST="$ABS_DEST" # we're in a subshell, so this is safe -- our integration-cli tests need DEST, and "cd" screws it up
		go test $TESTFLAGS
	)
}

bundle() {
	local bundle="$1"; shift
	echo "---> Making bundle: $(basename "$bundle")"
	source "$SCRIPTDIR/make/$bundle" "$@"
}

main() {
	if [ $# -lt 1 ]; then
		bundles=(${DEFAULT_BUNDLES[@]})
	else
		bundles=($@)
	fi
	for bundle in ${bundles[@]}; do
		bundle "$bundle"
		echo
	done
}

main "$@"
