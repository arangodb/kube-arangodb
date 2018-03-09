#!/bin/bash

# Sets the configuration of the cluster in a ConfigMap in kube-system.

cluster=$(kubectl config current-context)
echo "Configuring cluster $cluster"

read -p "Does the cluster require local storage (Y/n)? " answer
case ${answer:0:1} in
    n|N )
        REQUIRE_LOCAL_STORAGE=
    ;;
    * )
        REQUIRE_LOCAL_STORAGE=1
    ;;
esac

mapname="arango-operator-test"
configfile=$(mktemp)
cat <<EOF > $configfile
REQUIRE_LOCAL_STORAGE=${REQUIRE_LOCAL_STORAGE}
EOF
kubectl delete configmap $mapname -n kube-system --ignore-not-found 
kubectl create configmap $mapname -n kube-system --from-env-file=$configfile || exit 1

echo Stored configuration:
cat $configfile
rm $configfile
