#!/bin/bash

ROOT=$1

SHA_CODE=$(
find "${ROOT}/" \
  '(' -type f -name '*.go' -not -path "${ROOT}/vendor/*" -not -path "${ROOT}/.gobuild/*" -not -path "${ROOT}/deps/*" -exec sha256sum {} \; ')' -o \
  '(' -type f -name 'go.sum' -not -path "${ROOT}/vendor/*" -not -path "${ROOT}/.gobuild/*" -not -path "${ROOT}/deps/*" -exec sha256sum {} \; ')' -o \
  '(' -type f -name 'go.mod' -not -path "${ROOT}/vendor/*" -not -path "${ROOT}/.gobuild/*" -not -path "${ROOT}/deps/*" -exec sha256sum {} \; ')' \
  | cut -d ' ' -f1 | sha256sum | cut -d ' ' -f1
)

SHA_MOD=$(
find "${ROOT}/" \
  '(' -type f -name 'go.sum' -not -path "${ROOT}/vendor/*" -not -path "${ROOT}/.gobuild/*" -not -path "${ROOT}/deps/*" -exec sha256sum {} \; ')' -o \
  '(' -type f -name 'go.mod' -not -path "${ROOT}/vendor/*" -not -path "${ROOT}/.gobuild/*" -not -path "${ROOT}/deps/*" -exec sha256sum {} \; ')' \
  | cut -d ' ' -f1 | sha256sum | cut -d ' ' -f1
)

echo "Checksum Code: ${SHA_CODE}"
echo "Checksum Mod: ${SHA_MOD}"

echo -n "${SHA_CODE}" > ${ROOT}/.checksum.code
echo -n "${SHA_MOD}" > ${ROOT}/.checksum.mod