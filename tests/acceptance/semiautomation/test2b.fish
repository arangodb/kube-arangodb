#!/usr/bin/fish

source helper.fish
checkImages

set -g TESTNAME test2b
set -g TESTDESC "Scale a cluster deployment (development, enterprise)"
set -g YAMLFILE cluster.yaml
set -g DEPLOYMENT acceptance-cluster
printheader

patchYamlFile $YAMLFILE $ARANGODB_ENTERPRISE Development work.yaml

# Ensure enterprise license key
ensureLicenseKey

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
output "Scaling db servers up" "Patching Spec for Scaling up DBservers"
kubectl patch arango $DEPLOYMENT --type='json' -p='[{"op": "replace", "path": "/spec/dbservers/count", "value":5}]'
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 11 2
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 5 2
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 3 2
or fail "Deployment did not get ready."

# Patching
output "Scaling coordinators up" "Patching Spec for Scaling up coordinators"
kubectl patch arango $DEPLOYMENT --type='json' -p='[{"op": "replace", "path": "/spec/coordinators/count", "value":4}]'
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 12 2
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 5 2
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 4 2
or fail "Deployment did not get ready."

# Patching
output "Scaling dbservers down" "Patching Spec for Scaling down dbservers"
kubectl patch arango $DEPLOYMENT --type='json' -p='[{"op": "replace", "path": "/spec/dbservers/count", "value":2}]'
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 9 2
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 2 2
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 4 2
or fail "Deployment did not get ready."

# Patching
output "Scaling coordinators down" "Patching Spec for Scaling down coordinators"
kubectl patch arango $DEPLOYMENT --type='json' -p='[{"op": "replace", "path": "/spec/coordinators/count", "value":1}]'
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 6 2
and waitForKubectl "get pod" "$DEPLOYMENT-prmr" "1/1 *Running" 2 2
and waitForKubectl "get pod" "$DEPLOYMENT-agnt" "1/1 *Running" 3 2
and waitForKubectl "get pod" "$DEPLOYMENT-crdn" "1/1 *Running" 1 2
or fail "Deployment did not get ready."

# Manual check
output "Work" "Now please check external access on this URL with your browser:" "  https://$ip:8529/" "then type the outcome followed by ENTER."
inputAndLogResult

# Cleanup
kubectl delete -f work.yaml
and waitForKubectl "get pod" "$DEPLOYMENT" "1/1 *Running" 0 2
or fail "Could not delete deployment."

output "Ready" ""
