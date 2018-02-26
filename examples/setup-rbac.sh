#!/bin/bash

ROLE_NAME="${ROLE_NAME:-arangodb-operator}"
ROLE_BINDING_NAME="${ROLE_BINDING_NAME:-arangodb-operator}"
NAMESPACE="${NAMESPACE:-default}"

function usage {
echo "$(basename "$0") - Create Kubernetes RBAC role and bindings for ArangoDB operator
Usage: $(basename "$0") [options...]
Options:
  --role-name=STRING         Name of ClusterRole to create
                               (default=\"arangodb-operator\", environment variable: ROLE_NAME)
  --role-binding-name=STRING Name of ClusterRoleBinding to create
                               (default=\"arangodb-operator\", environment variable: ROLE_BINDING_NAME)
  --namespace=STRING         namespace to create role and role binding in. Must already exist.
                               (default=\"default\", environment vairable: NAMESPACE)
" >&2
}

function setupRole {
kubectl apply -f - << EOYAML
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: ${ROLE_NAME}
rules:
- apiGroups:
  - database.arangodb.com
  resources:
  - arangodeployments
  verbs:
  - "*"
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - "*"
- apiGroups:
  - ""
  resources:
  - pods
  - services
  - endpoints
  - persistentvolumeclaims
  - events
  - secrets
  verbs:
  - "*"
- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - "*"
EOYAML

local code=$?
if (code != 0); then
    exit $code
fi
}

function setupRoleBinding {
kubectl apply -f - << EOYAML
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: ${ROLE_BINDING_NAME}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ${ROLE_NAME}
subjects:
- kind: ServiceAccount
  name: default
  namespace: ${NAMESPACE}
EOYAML

local code=$?
if (code != 0); then
    exit $code
fi
}

for i in "$@"; do
    case $i in
        --role-name=*)
        ROLE_NAME="${i#*=}"
        ;;
        --role-binding-name=*)
        ROLE_BINDING_NAME="${i#*=}"
        ;;
        --namespace=*)
        NAMESPACE="${i#*=}"
        ;;
        -h|--help)
          usage
          exit 0
        ;;
        *)
          usage
          exit 1
        ;;
    esac
done

setupRole
setupRoleBinding
