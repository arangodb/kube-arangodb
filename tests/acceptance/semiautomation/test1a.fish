#!/usr/bin/fish

source helper.fish
checkImages

set -g TESTNAME test1a
set -g TESTDESC "Deployment of mode single (development)"
set -g YAMLFILE single.yaml
set -g DEPLOYMENT acceptance-single
printheader

patchYamlFile $YAMLFILE $ARANGODB_COMMUNITY Development work.yaml

# Deploy and check
kubectl apply -f work.yaml
and waitForKubectl "get pod" "$DEPLOYMENT-sngl" "1/1 *Running" 1 120
and waitForKubectl "get service" "$DEPLOYMENT *ClusterIP" 8529 1 120
and waitForKubectl "get service" "$DEPLOYMENT-ea *LoadBalancer" "-v;pending" 1 180
or fail "Deployment did not get ready."

# Automatic check
set ip (getLoadBalancerIP "$DEPLOYMENT-ea")
testArangoDB $ip 120
or fail "ArangoDB was not reachable."

# Manual check
output "Work" "Now please check external access on this URL with your browser:" "  https://$ip:8529/" "then type the outcome followed by ENTER."
inputAndLogResult

# Cleanup
kubectl delete -f work.yaml
waitForKubectl "get pod" $DEPLOYMENT-sngl "" 0 120
or fail "Could not delete deployment."

output "Ready" ""
