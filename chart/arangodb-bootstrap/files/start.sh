#!/bin/bash -e

ARANGODB_BIN=${ARANGODB_BIN:-/usr/bin/arangosh}
ARANGODB_URL=${ARANGODB_URL:-127.0.0.1}
ARANGODB_PORT=${ARANGODB_PORT:=8529}

EXEC="${ARANGODB_BIN}"

if [[ ! -z "${ARANGODB_JWT}" ]]; then
  EXEC="${EXEC} --server.authentication true --server.jwt-secret-keyfile ${ARANGODB_JWT}/token"
else
  EXEC="${EXEC} --server.authentication false"
fi

if [[ ! -z "${ARANGODB_TLS}" ]]; then
  EXEC="${EXEC} --server.endpoint http+ssl://${ARANGODB_URL}:${ARANGODB_PORT}"
else
  EXEC="${EXEC} --server.endpoint http+tcp://${ARANGODB_URL}:${ARANGODB_PORT}"
fi

EXEC="${EXEC} ${@}"

echo "Executing \`${EXEC}\`"

exec $EXEC
