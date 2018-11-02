#!/usr/bin/fish

source helper.fish

set -g TESTNAME test6c
set -g TESTDESC "Node resilience in mode cluster (production, enterprise)"
set -g YAMLFILE generated/cluster-enterprise-pro.yaml
set -g DEPLOYMENT acceptance-cluster
printheader

# Deploy and check
kubectl apply -f $YAMLFILE
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 3 120
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 120
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 3 120
and waitForKubectl "get service" "$DEPLOYMENT *ClusterIP" 8529 1 120
and waitForKubectl "get service" "$DEPLOYMENT-ea *LoadBalancer" "-v;pending" 1 180
or fail "Deployment did not get ready."

# Automatic check
set ip (getLoadBalancerIP "$DEPLOYMENT-ea")
testArangoDB $ip 120
or fail "ArangoDB was not reachable."

# Manual check
output "Work" "Now please check external access on this URL with your browser:" "  https://$ip:8529/" "then type the outcome followed by ENTER." "Furthermore, put some data in with replication factor 2." "Then, remove a node." "Pods should come back, service should not be interrupted." "Even writes should be possible during the redeployment." "All data must still be there."
inputAndLogResult

# Cleanup
kubectl delete -f $YAMLFILE
waitForKubectl "get pod" $DEPLOYMENT "" 0 120
or fail "Could not delete deployment."

output "Ready" ""
