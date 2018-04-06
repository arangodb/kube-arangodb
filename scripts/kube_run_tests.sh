#!/bin/bash

# Run kubectl run to run the integration tests.

DEPLOYMENTNAMESPACE=$1
TESTIMAGE=$2
ENTERPRISEIMAGE=$3
TESTTIMEOUT=$4
TESTLENGTHOPTIONS=$5

IMAGEID=$(docker inspect ${TESTIMAGE} '--format={{index .RepoDigests 0}}')

kubectl --namespace ${DEPLOYMENTNAMESPACE} \
    run arangodb-operator-test -i --rm --quiet --restart=Never \
    --image=${IMAGEID} \
    --env="ENTERPRISEIMAGE=${ENTERPRISEIMAGE}" \
    --env="TEST_NAMESPACE=${DEPLOYMENTNAMESPACE}" \
    --env="CLEANDEPLOYMENTS=${CLEANDEPLOYMENTS}" \
    -- \
    -test.v -test.timeout $TESTTIMEOUT $TESTLENGTHOPTIONS $TESTOPTIONS
