#!/bin/sh

NS=$1

if [ -z "$NS" ]; then
    echo "Specify a namespace argument"
    exit 1
fi

if [ -z "$2" ]; then
    echo "No secret set"
    exit 0
fi

SECRET=$(echo -n $2 | base64 -w 0)

kubectl apply -f - <<EOF
apiVersion: v1
data:
  token: ${SECRET}
kind: Secret
metadata:
  name: arangodb-backup-remote-secret
  namespace: ${NS}
type: Opaque
EOF
