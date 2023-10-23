# Operator API

A running operator exposes HTTP and gRPC API listeners to allow retrieving and setting some configuration values programmatically.
Both listeners require a secured connection to be established. It is possible to provide TLS certificate via k8s secret
using command line option `--api.tls-secret-name`. If secret name is not provided, operator will use self-signed certificate.

Some HTTP endpoints require the authorization to work with. All gRPC endpoints require the authorization.
The authorization can be accomplished by providing JWT token in 'Authorization' header, e.g. `Authorization: Bearer <token>`
The JWT token can be fetched from k8s secret (by default `arangodb-operator-api-jwt`). The token is generated automatically
on operator startup using the signing key specified in `arangodb-operator-api-jwt-key` secret. If it is empty or not exists,
the signing key will be auto-generated and saved into secret. You can specify other signing key using `--api.jwt-key-secret-name` CLI option.

## HTTP

The HTTP API is running at endpoint specified by operator command line options `--api.http-port` (8628 by default).

The HTTP API exposes endpoints used to get operator health and readiness status, operator version, and prometheus-compatible metrics.

For now only `/metrics` and `/log/level` endpoints require authorization.


## gRPC

The gRPC API is running at endpoint specified by operator command line options `--api.grpc-port` (8728 by default).

The gRPC API is exposed to allow programmatic access to some operator features and status.

gRPC protobuf definitions and go-client can be found at `github.com/kube-arangodb/pkg/api/server` package.

All gRPC requests require per-RPC metadata set to contain a valid Authorization header.
