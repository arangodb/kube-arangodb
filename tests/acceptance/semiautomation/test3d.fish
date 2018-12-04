#!/usr/bin/fish

source helper.fish
checkImages

set -g TESTNAME test3d
set -g TESTDESC "Scale a cluster deployment (production, enterprise)"
set -g YAMLFILE cluster.yaml
set -g DEPLOYMENT acceptance-cluster
printheader

patchYamlFile $YAMLFILE $ARANGODB_ENTERPRISE Production work.yaml

# Deploy and check
kubectl apply -f work.yaml
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 9 2
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 3 2
and waitForKubectl "get service" "$DEPLOYMENT *ClusterIP" 8529 1 2
and waitForKubectl "get service" "$DEPLOYMENT-ea *LoadBalancer" "-v;pending" 1 3
or fail "Deployment did not get ready."

# Automatic check
set ip (getLoadBalancerIP "$DEPLOYMENT-ea")
testArangoDB $ip 2
or fail "ArangoDB was not reachable."

# Patching
output "Scaling dbservers down" "Patching Spec for Scaling down dbservers"
kubectl patch arango $DEPLOYMENT --type='json' -p='[{"op": "replace", "path": "/spec/dbservers/count", "value":2}]'
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 8 2
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 2 2
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 3 2
or fail "Deployment did not get ready."

# Patching
output "Scaling coordinators down" "Patching Spec for Scaling down coordinators"
kubectl patch arango $DEPLOYMENT --type='json' -p='[{"op": "replace", "path": "/spec/coordinators/count", "value":2}]'
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 7 2
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 2 2
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 2 2
or fail "Deployment did not get ready."

# Patching
output "Scaling db servers up" "Patching Spec for Scaling up DBservers"
kubectl patch arango $DEPLOYMENT --type='json' -p='[{"op": "replace", "path": "/spec/dbservers/count", "value":3}]'
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 8 2
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 2 2
or fail "Deployment did not get ready."

# Patching
output "Scaling coordinators up" "Patching Spec for Scaling up coordinators"
kubectl patch arango $DEPLOYMENT --type='json' -p='[{"op": "replace", "path": "/spec/coordinators/count", "value":3}]'
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 9 2
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 3 2
or fail "Deployment did not get ready."

# Manual check
output "Work" "Now please check external access on this URL with your browser:" "  https://$ip:8529/" "then type the outcome followed by ENTER."
inputAndLogResult

# Cleanup
kubectl delete -f work.yaml
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 0 2
or fail "Could not delete deployment."

output "Ready" ""
