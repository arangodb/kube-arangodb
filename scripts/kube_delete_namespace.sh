#!/bin/bash

# Delete a namespace and wait until it is gone

NS=$1

if [ -z $NS ]; then
    echo "Specify a namespace argument"
    exit 1
fi

kubectl delete namespace $NS --now --ignore-not-found

# Wait until its really gone
while :; do
  response=$(kubectl get namespace $NS --template="non-empty" --ignore-not-found)
  if [ -z $response ]; then
    break
  fi
  sleep 0.5
done
