#!/bin/bash

# Run kubectl run to run the integration tests.

DEPLOYMENTNAMESPACE=$1
TESTIMAGE=$2
ARANGODIMAGE=$3
ENTERPRISEIMAGE=$4
TESTTIMEOUT=$5
TESTLENGTHOPTIONS=$6
TESTOPTIONS=$7
TESTREPOPATH=$8

IMAGEID=$(docker inspect ${TESTIMAGE} '--format={{index .RepoDigests 0}}')

kubectl --namespace ${DEPLOYMENTNAMESPACE} \
    run arangodb-operator-test -i --restart=Never \
    --image=${IMAGEID} \
    --env="ENTERPRISEIMAGE=${ENTERPRISEIMAGE}" \
    --env="ARANGODIMAGE=${ARANGODIMAGE}" \
    --env="TEST_NAMESPACE=${DEPLOYMENTNAMESPACE}" \
    --env="CLEANDEPLOYMENTS=${CLEANDEPLOYMENTS}" \
    --env="TESTDISABLEIPV6=${TESTDISABLEIPV6}" \
    --serviceaccount=arangodb-test \
    --env="TEST_REMOTE_REPOSITORY=${TESTREPOPATH}" \
    -- \
    -test.v -test.timeout $TESTTIMEOUT $TESTLENGTHOPTIONS $TESTOPTIONS
