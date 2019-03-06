#!/bin/bash

# Run kubectl run to run the integration tests.

DEPLOYMENTNAMESPACE=$1
ARANGODIMAGE=$2
ARANGOSYNCIMAGE=$3
ARANOSYNCTESTIMAGE=$4
ARANOSYNCTESTCTRLIMAGE=$5

ARANGOSYNCIMAGEID=$(docker inspect ${ARANGOSYNCIMAGE} '--format={{index .RepoDigests 0}}')
ARANOSYNCTESTIMAGEID=$(docker inspect ${ARANOSYNCTESTIMAGE} '--format={{index .RepoDigests 0}}')
ARANOSYNCTESTCTRLIMAGEID=$(docker inspect ${ARANOSYNCTESTCTRLIMAGE} '--format={{index .RepoDigests 0}}')

kubectl --namespace ${DEPLOYMENTNAMESPACE} \
    run arangodb-sync-test-controller -i --rm --quiet --restart=Never \
    --image=${ARANOSYNCTESTCTRLIMAGEID} \
    -- \
    --arango-image=${ARANGODIMAGE} \
    --arango-sync-image=${ARANGOSYNCIMAGEID} \
    --arango-sync-test-image=${ARANOSYNCTESTIMAGEID} \
    --license-key-secret-name=arangodb-jenkins-license-key \
    --namespace=${DEPLOYMENTNAMESPACE}