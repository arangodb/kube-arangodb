#!/bin/bash

# Collect logs from kube-arangodb operators

NS=$1
POSTFIX=$2

if [ -z $NS ]; then
    echo "Specify a namespace argument"
    exit 1
fi
if [ -z $POSTFIX ]; then
    echo "Specify a postfix argument"
    exit 1
fi

mkdir -p ./logs
kubectl logs -n ${NS} --selector=name=arango-deployment-operator &> ./logs/deployment-${POSTFIX}.log
kubectl logs -n kube-system --selector=name=arango-storage-operator &> ./logs/storage-${POSTFIX}.log
