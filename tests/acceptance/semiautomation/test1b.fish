#!/usr/bin/fish

source helper.fish

set -g TESTNAME test1b
set -g TESTDESC "Deployment of mode active/failover"
set -g YAMLFILE generated/activefailover-community-pro.yaml
set -g DEPLOYMENT acceptance-activefailover
printheader

# Deploy and check
kubectl apply -f $YAMLFILE
and waitForKubectl "get pod" $DEPLOYMENT "1 *Running" 5 120
and waitForKubectl "get pod" "$DEPLOYMENT-sngl.*1/1 *Running" "" 1 120
and waitForKubectl "get pod" "$DEPLOYMENT-sngl.*0/1 *Running" "" 1 120
and waitForKubectl "get service" "$DEPLOYMENT *ClusterIP" 8529 1 120
and waitForKubectl "get service" "$DEPLOYMENT-ea *LoadBalancer" "-v;pending" 1 120
or fail "Deployment did not get ready."

# Automatic check
set ip (getLoadBalancerIP "$DEPLOYMENT-ea")
testArangoDB $ip 60
or fail "ArangoDB was not reachable."

# Manual check
output "Work" "Now please check external access on this URL with your browser:" "  https://$ip:8529/" "then type the outcome followed by ENTER."
inputAndLogResult

# Cleanup
kubectl delete -f $YAMLFILE
waitForKubectl "get pod" $DEPLOYMENT "" 0 120
or fail "Could not delete deployment."

output "Ready" ""
