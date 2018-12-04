#!/bin/sh

LICENSE=$2
NS=$1

if [ -z $LICENSE ]; then
    echo "No enterprise license set"
    exit 0
fi

if [ -z $NS ]; then
    echo "Specify a namespace argument"
    exit 1
fi

kubectl apply -f - <<EOF
apiVersion: v1
data:
  token: ${LICENSE}
kind: Secret
metadata:
  name: arangodb-jenkins-license-key
  namespace: ${NS}
type: Opaque
EOF
