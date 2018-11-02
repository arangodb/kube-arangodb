#!/usr/bin/fish

source helper.fish

set -g TESTNAME test6a
set -g TESTDESC "Node resilience in mode single (production)"
set -g YAMLFILE single.yaml
set -g DEPLOYMENT acceptance-single
printheader

patchYamlFile $YAMLFILE $ARANGODB_COMMUNITY Production work.yaml

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
output "Work" "Now please check external access on this URL with your browser:" "  https://$ip:8529/" "then type the outcome followed by ENTER." "Furthermore, put some data in and reboot node the single pod is running on." "Wait until it comes back and then see if the data is still there and the server is responsive."
inputAndLogResult

# Cleanup
kubectl delete -f work.yaml
waitForKubectl "get pod" $DEPLOYMENT-sngl "" 0 120
or fail "Could not delete deployment."

output "Ready" ""
