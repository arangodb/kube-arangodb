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
    case $(uname) in
        Darwin)
            sed -e "${EXPR}" -i "" ${FILE}
            ;;
        *)
            sed -i --expression "${EXPR}" ${FILE}
            ;;
    esac
}


f=README.md
replaceInFile "s@^kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/.*/manifests/arango-crd.yaml\$@kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/${VERSION}/manifests/arango-crd.yaml@g" ${f}
replaceInFile "s@^kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/.*/manifests/arango-deployment.yaml\$@kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/${VERSION}/manifests/arango-deployment.yaml@g" ${f}
replaceInFile "s@^kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/.*/manifests/arango-deployment-replication.yaml\$@kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/${VERSION}/manifests/arango-deployment-replication.yaml@g" ${f}
replaceInFile "s@^kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/.*/manifests/arango-storage.yaml\$@kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/${VERSION}/manifests/arango-storage.yaml@g" ${f}

replaceInFile "s@^helm install https://github.com/arangodb/kube-arangodb/releases/download/.*/kube-arangodb-crd.tgz\$@helm install https://github.com/arangodb/kube-arangodb/releases/download/${VERSION}/kube-arangodb-crd.tgz@g" ${f}
replaceInFile "s@^helm install https://github.com/arangodb/kube-arangodb/releases/download/.*/kube-arangodb.tgz\$@helm install https://github.com/arangodb/kube-arangodb/releases/download/${VERSION}/kube-arangodb.tgz@g" ${f}
replaceInFile "s@^helm install https://github.com/arangodb/kube-arangodb/releases/download/.*/kube-arangodb-storage.tgz\$@helm install https://github.com/arangodb/kube-arangodb/releases/download/${VERSION}/kube-arangodb-storage.tgz@g" ${f}
