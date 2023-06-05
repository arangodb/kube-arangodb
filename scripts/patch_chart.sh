#!/bin/bash

# Updates the versions in helm charts to reflect the current 
# version.

VERSION=$1
IMAGE=$2

if [ -z $VERSION ]; then
    echo "Specify a version argument"
    exit 1
fi

function replaceInFile {
    local EXPR=$1
    local FILE=$2
    sed -i --expression "${EXPR}" ${FILE}
}

for f in kube-arangodb kube-arangodb-crd; do
    replaceInFile "s@^version: .*\$@version: ${VERSION}@g" "chart/${f}/Chart.yaml"
    if [[ -f "chart/${f}/values.yaml" ]]; then
        replaceInFile "s@^  image: .*\$@  image: ${IMAGE}@g" "chart/${f}/values.yaml"
    fi
done