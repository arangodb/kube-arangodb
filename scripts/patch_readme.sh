#!/bin/bash

# Updates the installation instructions in README.md to reflect the current 
# version.

VERSION=$1

if [ -z $VERSION ]; then
    echo "Specify a version argument"
    exit 1
fi

f=README.md
sed -e "s@^kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/.*/manifests/crd.yaml\$@kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/${VERSION}/manifests/crd.yaml@g" -i "" $f
sed -e "s@^kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/.*/manifests/arango-deployment.yaml\$@kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/${VERSION}/manifests/arango-deployment.yaml@g" -i "" $f
sed -e "s@^kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/.*/manifests/arango-deployment-replication.yaml\$@kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/${VERSION}/manifests/arango-deployment-replication.yaml@g" -i "" $f
sed -e "s@^kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/.*/manifests/arango-storage.yaml\$@kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/${VERSION}/manifests/arango-storage.yaml@g" -i "" $f

sed -e "s@^helm install https://github.com/arangodb/kube-arangodb/releases/download/.*/kube-arangodb.tgz\$@helm install https://github.com/arangodb/kube-arangodb/releases/download/${VERSION}/kube-arangodb.tgz@g" -i "" $f
sed -e "s@^helm install https://github.com/arangodb/kube-arangodb/releases/download/.*/kube-arangodb-storage.tgz\$@helm install https://github.com/arangodb/kube-arangodb/releases/download/${VERSION}/kube-arangodb-storage.tgz@g" -i "" $f
