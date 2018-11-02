#!/usr/bin/fish

source helper.fish

set -g TESTNAME test7a
set -g TESTDESC "Deployment of 2 clusters with sync with DC2DC (production, enterprise)"
set -g YAMLFILE generated/cluster-sync-enterprise-pro.yaml
set -g YAMLFILE2 generated/cluster-sync2-enterprise-pro.yaml
set -g DEPLOYMENT acceptance-cluster
set -g DEPLOYMENT2 acceptance-cluster2
printheader

# Deploy and check
kubectl apply -f $YAMLFILE
kubectl apply -f $YAMLFILE2
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 15 120
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 3 120
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 120
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 3 120
and waitForKubectl "get pod" "$DEPLOYMENT-syma" "1/1 *Running" 3 120
and waitForKubectl "get pod" "$DEPLOYMENT-sywo" "1/1 *Running" 3 120
and waitForKubectl "get service" "$DEPLOYMENT *ClusterIP" 8529 1 120
and waitForKubectl "get service" "$DEPLOYMENT-ea *LoadBalancer" "-v;pending" 1 180
and waitForKubectl "get service" "$DEPLOYMENT-sync *LoadBalancer" "-v;pending" 1 180
and waitForKubectl "get pod" "$DEPLOYMENT2" "1/1 *Running" 15 120
and waitForKubectl "get pod" "$DEPLOYMENT2-prmr" "1/1 *Running" 3 120
and waitForKubectl "get pod" "$DEPLOYMENT2-agnt" "1/1 *Running" 3 120
and waitForKubectl "get pod" "$DEPLOYMENT2-crdn" "1/1 *Running" 3 120
and waitForKubectl "get pod" "$DEPLOYMENT2-syma" "1/1 *Running" 3 120
and waitForKubectl "get pod" "$DEPLOYMENT2-sywo" "1/1 *Running" 3 120
and waitForKubectl "get service" "$DEPLOYMENT2 *ClusterIP" 8529 1 120
and waitForKubectl "get service" "$DEPLOYMENT2-ea *LoadBalancer" "-v;pending" 1 180
and waitForKubectl "get service" "$DEPLOYMENT2-sync *LoadBalancer" "-v;pending" 1 180
or fail "Deployment did not get ready."

# Automatic check
set ip (getLoadBalancerIP "$DEPLOYMENT-ea")
testArangoDB $ip 120
or fail "ArangoDB (1) was not reachable."

set ip2 (getLoadBalancerIP "$DEPLOYMENT2-ea")
testArangoDB $ip2 120
or fail "ArangoDB (2) was not reachable."

# Set up replication, rest is manual:
# run sed here on replication.yaml, find sync-ea first
kubectl apply -f replication.yaml

# Manual check
output "Work" "Now please check external access on this URL with your browser:" "  https://$ip:8529/" "then type the outcome followed by ENTER."
inputAndLogResult

# Cleanup
kubectl delete -f replication.yaml
sleep 15
kubectl delete -f $YAMLFILE
kubectl delete -f $YAMLFILE2
waitForKubectl "get pod" $DEPLOYMENT "" 0 120
waitForKubectl "get pod" $DEPLOYMENT2 "" 0 120
or fail "Could not delete deployment."

output "Ready" ""
