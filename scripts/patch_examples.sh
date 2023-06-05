#!/bin/bash

# Updates examples to match current version.

VERSION=$1

if [ -z $VERSION ]; then
    echo "Specify a version argument"
    exit 1
fi

ARANGODB_VERSION=3.10.6

function replaceInFile {
    local EXPR=$1
    local FILE=$2
    sed -E -i --expression "${EXPR}" ${FILE}
}

FILES=$(find ./examples -type f -name '*.yaml')

for FILE in ${FILES}; do
  replaceInFile "s@arangodb/arangodb:[0-9]+\\.[0-9]+\\.[0-9]+@arangodb/arangodb:${ARANGODB_VERSION}@g" ${FILE}
  replaceInFile "s@arangodb/enterprise:[0-9]+\\.[0-9]+\\.[0-9]+@arangodb/enterprise:${ARANGODB_VERSION}@g" ${FILE}
  replaceInFile "s@arangodb/kube-arangodb:[0-9]+\\.[0-9]+\\.[0-9]+@arangodb/kube-arangodb:${VERSION}@g" ${FILE}
  replaceInFile "s@arangodb/kube-arangodb-enterprise:[0-9]+\\.[0-9]+\\.[0-9]+@arangodb/kube-arangodb-enterprise:${VERSION}@g" ${FILE}
done