#!/bin/bash

# Delete a namespace and wait until it is gone

NS=$1

if [ -z $NS ]; then
    echo "Specify a namespace argument"
    exit 1
fi

kubectl delete namespace $NS --now --ignore-not-found
response=$(kubectl get namespace $NS --template="non-empty" --ignore-not-found)
while [ ! -z $response ]; do
    sleep 1
    response=$(kubectl get namespace $NS --template="non-empty" --ignore-not-found)
done
