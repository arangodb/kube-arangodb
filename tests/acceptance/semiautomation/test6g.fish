#!/usr/bin/fish

source helper.fish
checkImages

set -g TESTNAME test6g
set -g TESTDESC "Node resilience in mode cluster (development, enterprise, local storage)"
set -g YAMLFILE cluster-local-storage.yaml
set -g YAMLFILESTORAGE local-storage.yaml
set -g DEPLOYMENT acceptance-cluster
printheader

patchYamlFile $YAMLFILE $ARANGODB_ENTERPRISE Development work.yaml

# Ensure enterprise license key
ensureLicenseKey

# Deploy local storage:
kubectl apply -f $YAMLFILESTORAGE
and waitForKubectl "get storageclass" "acceptance.*arangodb.*localstorage" "" 1 1
or fail "Local storage could not be deployed."

# Deploy and check
kubectl apply -f work.yaml
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 3 2
and waitForKubectl "get service" "$DEPLOYMENT *ClusterIP" 8529 1 2
and waitForKubectl "get service" "$DEPLOYMENT-ea *LoadBalancer" "-v;pending" 1 3
and waitForKubectl "get pvc" "$DEPLOYMENT" "RWO *acceptance" 6 2
or fail "Deployment did not get ready."

# Automatic check
set ip (getLoadBalancerIP "$DEPLOYMENT-ea")
testArangoDB $ip 2
or fail "ArangoDB was not reachable."

# Manual check
output "Work" "Now please check external access on this URL with your browser:" "  https://$ip:8529/" "then put some data in there in different collections, some with" "replicationFactor set to 1 and some set to 2." "Then cordon off a node running a dbserver pod and delete the pod." "Service (including writes) must continue, except for the collection without" "replication. It should be possible to drop that collection and eventually" "remove the dbserver. A new dbserver should come up on a different node" "after some time."
inputAndLogResult

# Cleanup
kubectl delete -f work.yaml
waitForKubectl "get pod" $DEPLOYMENT "" 0 2
or fail "Could not delete deployment."

kubectl delete -f $YAMLFILESTORAGE
kubectl delete storageclass acceptance
waitForKubectl "get storageclass" "acceptance.*arangodb.*localstorage" "" 0 2
or fail "Could not delete deployed storageclass."

output "Ready" ""
