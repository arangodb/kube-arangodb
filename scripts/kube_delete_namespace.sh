#!/bin/bash

# Delete a namespace and wait until it is gone

NS=$1

kubectl delete namespace $NS --now --ignore-not-found
response=$(kubectl get namespace $NS --template="non-empty" --ignore-not-found)
while [ ! -z $response ]; do
    sleep 1
    response=$(kubectl get namespace $NS --template="non-empty" --ignore-not-found)
done
