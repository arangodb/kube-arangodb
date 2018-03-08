#!/bin/bash

# Create the local storage in the cluster if the cluster needs it.

set -e

NS=$1

if [ -z $NS ]; then
    echo "Specify a namespace argument"
    exit 1
fi

# Fetch cluster config
mapname="arango-operator-test"
eval $(kubectl get configmap $mapname -n kube-system --ignore-not-found --template='{{ range $key, $value := .data }}export {{$key}}={{$value}}
{{ end }}')

if [ "${REQUIRE_LOCAL_STORAGE}" = "1" ]; then
  echo "Preparing local storage"
  kubectl apply -n $NS -f examples/arango-local-storage.yaml
else
  echo "No local storage needed for this cluster"
fi
