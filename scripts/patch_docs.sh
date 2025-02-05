#!/bin/bash

# Updates the installation instructions in README.md to reflect the current 
# version.

VERSION=$1

if [ -z $VERSION ]; then
    echo "Specify a version argument"
    exit 1
fi

function replaceInFile {
    local EXPR=$1
    local FILE=$2
    sed -E -i --expression "${EXPR}" ${FILE}
}


shift 1

for f in "$@"; do
  replaceInFile "s@https://github.com/arangodb/kube-arangodb/blob/[0-9]+\.[0-9]+\.[0-9]+/@https://github.com/arangodb/kube-arangodb/blob/${VERSION}/@g" ${f}
done
