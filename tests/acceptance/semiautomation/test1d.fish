#!/usr/bin/fish

source helper.fish
checkImages

set -g TESTNAME test1d
set -g TESTDESC "Deployment of mode cluster with sync (development, enterprise)"
set -g YAMLFILE cluster-sync.yaml
set -g DEPLOYMENT acceptance-cluster
printheader

patchYamlFile $YAMLFILE $ARANGODB_ENTERPRISE Development work.yaml

# Deploy and check
kubectl apply -f work.yaml
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 15 2
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-syma" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-sywo" "1/1 *Running" 3 2
and waitForKubectl "get service" "$DEPLOYMENT *ClusterIP" 8529 1 2
and waitForKubectl "get service" "$DEPLOYMENT-ea *LoadBalancer" "-v;pending" 1 3
and waitForKubectl "get service" "$DEPLOYMENT-sync *LoadBalancer" "-v;pending" 1 3
or fail "Deployment did not get ready."

# Automatic check
set ip (getLoadBalancerIP "$DEPLOYMENT-ea")
testArangoDB $ip 2
or fail "ArangoDB was not reachable."

# Manual check
output "Work" "Now please check external access on this URL with your browser:" "  https://$ip:8529/" "then type the outcome followed by ENTER."
inputAndLogResult

# Cleanup
kubectl delete -f work.yaml
waitForKubectl "get pod" $DEPLOYMENT "" 0 2
or fail "Could not delete deployment."

output "Ready" ""
