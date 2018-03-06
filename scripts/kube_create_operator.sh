#!/bin/bash

# Create the operator deployment with custom image option

NS=$1
IMAGE=$2
PULLPOLICY="${PULLPOLICY:-IfNotPresent}"

if [ -z $NS ]; then
    echo "Specify a namespace argument"
    exit 1
fi
if [ -z $IMAGE ]; then
    echo "Specify an image argument"
    exit 1
fi

if [ ! -z $USESHA256 ]; then
  IMAGE=$(docker inspect --format='{{index .RepoDigests 0}}' ${IMAGE})
fi

kubectl --namespace=$NS create -f - << EOYAML
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: arangodb-operator
spec:
  replicas: 1
  template:
    metadata:
      labels:
        name: arangodb-operator
    spec:
      containers:
      - name: arangodb-operator
        imagePullPolicy: ${PULLPOLICY}
        image: ${IMAGE}
        env:
        - name: MY_POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: MY_POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name

EOYAML

exit $?
