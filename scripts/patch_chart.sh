#!/bin/bash

# Updates the versions in helm charts to reflect the current 
# version.

VERSION=$1

if [ -z $VERSION ]; then
    echo "Specify a version argument"
    exit 1
fi

function replaceInFile {
    local EXPR=$1
    local FILE=$2
    sed -i --expression "${EXPR}" ${FILE}
}

for f in kube-arangodb kube-arangodb-enterprise kube-arangodb-arm64 kube-arangodb-enterprise-arm64 kube-arangodb-crd; do
    replaceInFile "s@^version: .*\$@version: ${VERSION}@g" "chart/${f}/Chart.yaml"
    if [[ -f "chart/${f}/values.yaml" ]]; then
        replaceInFile "s@^  image: arangodb/kube-arangodb:[[:digit:]].*\$@  image: arangodb/kube-arangodb:${VERSION}@g" "chart/${f}/values.yaml"
        replaceInFile "s@^  image: arangodb/kube-arangodb-enterprise:[[:digit:]].*\$@  image: arangodb/kube-arangodb-enterprise:${VERSION}@g" "chart/${f}/values.yaml"
    fi
done