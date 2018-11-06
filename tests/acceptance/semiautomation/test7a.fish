#!/usr/bin/fish

source helper.fish
checkImages

set -g TESTNAME test7a
set -g TESTDESC "Deployment of 2 clusters with sync with DC2DC (production, enterprise)"
set -g YAMLFILE cluster-sync1.yaml
set -g YAMLFILE2 cluster-sync2.yaml
set -g DEPLOYMENT acceptance-cluster1
set -g DEPLOYMENT2 acceptance-cluster2
printheader

patchYamlFile $YAMLFILE $ARANGODB_ENTERPRISE Production work.yaml
patchYamlFile $YAMLFILE2 $ARANGODB_ENTERPRISE Production work2.yaml
cp replication.yaml work3.yaml

# Deploy and check
kubectl apply -f work.yaml
kubectl apply -f work2.yaml
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

# Deploy secrets separately for sync to pick them up:
kubectl get secret src-accesspackage --template='{{index .data "accessPackage.yaml"}}' | base64 -d > accessPackage.yaml
and kubectl apply -f accessPackage.yaml
or fail "Could not redeploy secrets for replication auth."

# Automatic check
set ip (getLoadBalancerIP "$DEPLOYMENT-ea")
testArangoDB $ip 120
or fail "ArangoDB (1) was not reachable."

set ip2 (getLoadBalancerIP "$DEPLOYMENT2-ea")
testArangoDB $ip2 120
or fail "ArangoDB (2) was not reachable."

set ip3 (getLoadBalancerIP "$DEPLOYMENT-sync")
sed -i "s|@ADDRESS@|$ip3|" work3.yaml

# Set up replication, rest is manual:
# run sed here on replication.yaml, find sync-ea first
kubectl apply -f work3.yaml

# Manual check
output "Work" "Now please check external access on this URL with your browser:" "  https://$ip:8529/" "then type the outcome followed by ENTER."
inputAndLogResult

# Cleanup
kubectl delete -f work3.yaml
sleep 15
kubectl delete -f work.yaml
kubectl delete -f work2.yaml
waitForKubectl "get pod" $DEPLOYMENT "" 0 120
waitForKubectl "get pod" $DEPLOYMENT2 "" 0 120
or fail "Could not delete deployment."

output "Ready" ""
