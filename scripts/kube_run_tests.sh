#!/bin/bash

# Run kubectl run to run the integration tests.

DEPLOYMENTNAMESPACE=$1
TESTIMAGE=$2
ARANGODIMAGE=$3
ENTERPRISEIMAGE=$4
TESTTIMEOUT=$5
TESTLENGTHOPTIONS=$6

IMAGEID=$(docker inspect ${TESTIMAGE} '--format={{index .RepoDigests 0}}')

kubectl --namespace ${DEPLOYMENTNAMESPACE} \
    run arangodb-operator-test -i --rm --quiet --restart=Never \
    --image=${IMAGEID} \
    --env="ENTERPRISEIMAGE=${ENTERPRISEIMAGE}" \
    --env="ARANGODIMAGE=${ARANGODIMAGE}" \
    --env="TEST_NAMESPACE=${DEPLOYMENTNAMESPACE}" \
    --env="CLEANDEPLOYMENTS=${CLEANDEPLOYMENTS}" \
    -- \
    -test.v -test.timeout $TESTTIMEOUT $TESTLENGTHOPTIONS $TESTOPTIONS
