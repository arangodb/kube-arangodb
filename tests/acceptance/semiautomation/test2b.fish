#!/usr/bin/fish

source helper.fish

set -g TESTNAME test2b
set -g TESTDESC "Scale an cluster deployment (enterprise)"
set -g YAMLFILE generated/cluster-enterprise-dev.yaml
set -g DEPLOYMENT acceptance-cluster
printheader

# Deploy and check
kubectl apply -f $YAMLFILE
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 9 120
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 3 120
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 120
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 3 120
and waitForKubectl "get service" "$DEPLOYMENT *ClusterIP" 8529 1 120
and waitForKubectl "get service" "$DEPLOYMENT-ea *LoadBalancer" "-v;pending" 1 120
or fail "Deployment did not get ready."

# Automatic check
set ip (getLoadBalancerIP "$DEPLOYMENT-ea")
testArangoDB $ip 60
or fail "ArangoDB was not reachable."

# Patching
output "Patching" "Patching Spec for Scaling up"
kubectl patch arango $DEPLOYMENT --type='json' -p='[{"op": "replace", "path": "/spec/dbservers/count", "value":5}]'
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 11 120
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 5 120
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 120
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 3 120
or fail "Deployment did not get ready."

# Patching
output "Patching" "Patching Spec for Scaling up"
kubectl patch arango $DEPLOYMENT --type='json' -p='[{"op": "replace", "path": "/spec/coordinators/count", "value":4}]'
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 12 120
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 5 120
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 120
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 4 120
or fail "Deployment did not get ready."

# Patching
output "Patching" "Patching Spec for Scaling up"
kubectl patch arango $DEPLOYMENT --type='json' -p='[{"op": "replace", "path": "/spec/dbservers/count", "value":2}]'
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 9 120
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 2 120
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 120
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 4 120
or fail "Deployment did not get ready."

# Patching
output "Patching" "Patching Spec for Scaling up"
kubectl patch arango $DEPLOYMENT --type='json' -p='[{"op": "replace", "path": "/spec/coordinators/count", "value":1}]'
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 6 120
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 2 120
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 120
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 1 120
or fail "Deployment did not get ready."

# Manual check
output "Work" "Now please check external access on this URL with your browser:" "  https://$ip:8529/" "then type the outcome followed by ENTER."
inputAndLogResult

# Cleanup
kubectl delete -f $YAMLFILE
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 0 120
or fail "Could not delete deployment."

output "Ready" ""
