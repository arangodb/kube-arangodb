#!/usr/bin/fish

source helper.fish
checkImages

set -g TESTNAME test2a
set -g TESTDESC "Scale an active failover deployment (enterprise, development)"
set -g YAMLFILE activefailover.yaml
set -g DEPLOYMENT acceptance-activefailover
printheader

patchYamlFile $YAMLFILE $ARANGODB_ENTERPRISE Development work.yaml

# Ensure enterprise license key
ensureLicenseKey

# Deploy and check
kubectl apply -f work.yaml
and waitForKubectl "get pod" $DEPLOYMENT "1 *Running" 5 2
and waitForKubectl "get pod" "$DEPLOYMENT-sngl.*1/1 *Running" "" 1 2
and waitForKubectl "get pod" "$DEPLOYMENT-sngl.*0/1 *Running" "" 1 2
and waitForKubectl "get service" "$DEPLOYMENT *ClusterIP" 8529 1 2
and waitForKubectl "get service" "$DEPLOYMENT-ea *LoadBalancer" "-v;pending" 1 3
or fail "Deployment did not get ready."

# Automatic check
set ip (getLoadBalancerIP "$DEPLOYMENT-ea")
testArangoDB $ip 2
or fail "ArangoDB was not reachable."

# Scale up the deployment
output "Next" "Patching Spec for Scaling up"
kubectl patch arango $DEPLOYMENT --type='json' -p='[{"op": "replace", "path": "/spec/single/count", "value":3}]'
and waitForKubectl "get pod" $DEPLOYMENT "1 *Running" 6 2
and waitForKubectl "get pod" "$DEPLOYMENT-sngl.*1/1 *Running" "" 1 2
and waitForKubectl "get pod" "$DEPLOYMENT-sngl.*0/1 *Running" "" 2 2
or fail "Patched deployment did not get ready."

# Scale down the deployment
output "Next" "Patching Spec for Scaling down"
kubectl patch arango $DEPLOYMENT --type='json' -p='[{"op": "replace", "path": "/spec/single/count", "value":2}]'
and waitForKubectl "get pod" $DEPLOYMENT "1 *Running" 5 2
and waitForKubectl "get pod" "$DEPLOYMENT-sngl.*1/1 *Running" "" 1 2
and waitForKubectl "get pod" "$DEPLOYMENT-sngl.*0/1 *Running" "" 1 2
or fail "Patched deployment did not get ready."

# Manual check
output "Work" "Now please check external access on this URL with your browser:" "  https://$ip:8529/" "then type the outcome followed by ENTER."
inputAndLogResult

# Cleanup
kubectl delete -f work.yaml
waitForKubectl "get pod" $DEPLOYMENT-sngl "" 0 2
or fail "Could not delete deployment."

output "Ready" ""
