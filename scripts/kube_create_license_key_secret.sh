#!/bin/sh

if [ -z $ENTERPRISELICENSE ]; then
    exit 0
fi

LICENSE=$(echo "${ENTERPRISELICENSE}" | base64 )

kubectl apply -f - <<EOF
apiVersion: v1
data:
  token: ${LICENSE}
kind: Secret
metadata:
  name: arangodb-jenkins-license-key
  namespace: default
type: Opaque
EOF
