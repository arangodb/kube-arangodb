#!/bin/sh

if [ -z $ENTERPRISELICENSE ]; then
    exit 0
fi

LICENSE=$(echo "${ENTERPRISELICENSE}" | base64 )
NS=$1

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
