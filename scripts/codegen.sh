#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$1

THIS_PKG=github.com/arangodb/kube-arangodb

sed -i --posix "s@\^v\[0-9\]\+((alpha|beta)\[0-9\]\+)\?@^v[0-9]+((alpha|beta)[0-9]*)?@g" "${SCRIPT_ROOT}/deps/k8s.io/code-generator/kube_codegen.sh"

source "${SCRIPT_ROOT}/deps/k8s.io/code-generator/kube_codegen.sh"

kube::codegen::gen_helpers \
    --boilerplate "${SCRIPT_ROOT}/tools/codegen/boilerplate.go.txt" \
    "${SCRIPT_ROOT}/pkg/apis"

kube::codegen::gen_client \
    --boilerplate "${SCRIPT_ROOT}/tools/codegen/boilerplate.go.txt" \
    --with-watch \
    --output-dir "${SCRIPT_ROOT}/pkg/generated" \
    --output-pkg "${THIS_PKG}/pkg/generated" \
    "${SCRIPT_ROOT}/pkg/apis"