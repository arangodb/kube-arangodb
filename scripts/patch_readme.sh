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


f=README.md

replaceInFile "s@arangodb/kube-arangodb:[0-9]+\.[0-9]+\.[0-9]+@arangodb/kube-arangodb:${VERSION}@g" ${f}
replaceInFile "s@arangodb/kube-arangodb-enterprise:[0-9]+\.[0-9]+\.[0-9]+@arangodb/kube-arangodb-enterprise:${VERSION}@g" ${f}

replaceInFile "s@^kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/.*/manifests/arango-crd.yaml\$@kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/${VERSION}/manifests/arango-crd.yaml@g" ${f}
replaceInFile "s@^kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/.*/manifests/arango-deployment.yaml\$@kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/${VERSION}/manifests/arango-deployment.yaml@g" ${f}
replaceInFile "s@^kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/.*/manifests/arango-deployment-replication.yaml\$@kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/${VERSION}/manifests/arango-deployment-replication.yaml@g" ${f}
replaceInFile "s@^kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/.*/manifests/arango-storage.yaml\$@kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/${VERSION}/manifests/arango-storage.yaml@g" ${f}

replaceInFile "s@^kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/.*/manifests/enterprise-crd.yaml\$@kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/${VERSION}/manifests/enterprise-crd.yaml@g" ${f}
replaceInFile "s@^kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/.*/manifests/enterprise-deployment.yaml\$@kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/${VERSION}/manifests/enterprise-deployment.yaml@g" ${f}
replaceInFile "s@^kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/.*/manifests/enterprise-deployment-replication.yaml\$@kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/${VERSION}/manifests/enterprise-deployment-replication.yaml@g" ${f}
replaceInFile "s@^kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/.*/manifests/enterprise-storage.yaml\$@kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/${VERSION}/manifests/enterprise-storage.yaml@g" ${f}

replaceInFile "s@https://github\.com/arangodb/kube-arangodb/releases/download/.*/kube-arangodb-crd-[[:digit:]].*\.tgz@https://github.com/arangodb/kube-arangodb/releases/download/${VERSION}/kube-arangodb-crd-${VERSION}.tgz@g" ${f}
replaceInFile "s@https://github\.com/arangodb/kube-arangodb/releases/download/.*/kube-arangodb-[[:digit:]].*\.tgz@https://github.com/arangodb/kube-arangodb/releases/download/${VERSION}/kube-arangodb-${VERSION}.tgz@g" ${f}
