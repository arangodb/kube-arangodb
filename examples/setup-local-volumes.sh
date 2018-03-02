#!/bin/bash

VOLUME_NAME_PREFIX="${VOLUME_NAME_PREFIX:-vol}"
VOLUME_COUNT=${VOLUME_COUNT:-5}
LOCAL_PATH="${LOCAL_PATH:-/var/lib/kubernetes/volumes}"
STORAGECLASS="${STORAGECLASS:-local-ssd}"

function usage {
echo "$(basename "$0") - Create Kubernetes Persistent Local Volumes
Usage: $(basename "$0") [options...]
Options:
  --volume-name-prefix=STRING Prefix of volume names.
                              (default=\"vol\", environment variable: VOLUME_NAME_PREFIX)
  --volume-count=INT          Number of olumes to create.
                              (default=5, environment variable: VOLUME_COUNT)
  --node=STRING               Name of the node the olume is created on.
                              (default=empty, environment variable: NODE)
  --local-path=STRING         Local path on node where volumes will start.
                              (default=\"/var/lib/kubernetes/volumes\", environment variable: LOCAL_PATH)
  --storageclass=STRING       Name of the storageclass for the volumes.
                              (default=\"local-ssd\", environment variable: STORAGECLASS)
" >&2
}

function setupVolume {
  local NAME=$1
  local NODE=$2
  local LOCALPATH=$3
kubectl apply -f - << EOYAML
apiVersion: v1
kind: PersistentVolume
metadata:
  name: $NAME
  annotations:
        "volume.alpha.kubernetes.io/node-affinity": '{
            "requiredDuringSchedulingIgnoredDuringExecution": {
                "nodeSelectorTerms": [
                    { "matchExpressions": [
                        { "key": "kubernetes.io/hostname",
                          "operator": "In",
                          "values": ["$NODE"]
                        }
                    ]}
                 ]}
              }'
spec:
    capacity:
      storage: 100Gi
    accessModes:
    - ReadWriteOnce
    persistentVolumeReclaimPolicy: Retain
    storageClassName: $STORAGECLASS
    local:
      path: $LOCALPATH
EOYAML
}

function createVolumes {
  if [ -z ${NODE} ]; then
    echo "--node missing"
    exit 1
  fi
  for (( idx=1; idx<=$VOLUME_COUNT; idx++ )); do
    local NAME="${NODE}-${VOLUME_NAME_PREFIX}-${idx}"
    local LOCALPATH="${LOCAL_PATH}/${NAME}"
    setupVolume $NAME $NODE $LOCALPATH
  done
}

for i in "$@"; do
    case $i in
        --volume-name-prefix=*)
        VOLUME_NAME_PREFIX="${i#*=}"
        ;;
        --volume-count=*)
        VOLUME_COUNT="${i#*=}"
        ;;
        --node=*)
        NODE="${i#*=}"
        ;;
        --local-path=*)
        LOCAL_PATH="${i#*=}"
        ;;
        -h|--help)
          usage
          exit 0
        ;;
        *)
          usage
          exit 1
        ;;
    esac
done

createVolumes
