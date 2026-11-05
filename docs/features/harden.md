---
layout: page
title: Harden ArangoDB Server Containers
parent: List of all features
---

# Harden ArangoDB Server Containers

## Overview

When enabled, the operator appends a set of hardening arguments to the `arangod` command line of the
server containers of every server group (Agents, DBServers, Coordinators, Single). The arguments lock
down the JavaScript/Foxx runtime, restrict which process environment variables scripts can read, and
require a JWT for the log and backup APIs.

The feature is disabled by default and depends on the [Secured containers](secured_containers.md)
feature - it only takes effect when `secured-containers` is enabled.

## What is changed

The following arguments are appended to the `arangod` containers of all server groups:

| Argument | Effect |
|:--- |:--- |
| `--server.harden` | Restricts privileged server APIs to superusers |
| `--javascript.harden` | Disables JavaScript functions that expose host/process information |
| `--javascript.startup-options-denylist=.*` | Hides every startup option from JavaScript actions |
| `--javascript.environment-variables-allowlist=^HOSTNAME$` | Allows JavaScript to read only the `HOSTNAME` environment variable... |
| `--javascript.environment-variables-allowlist=^PATH$` | ...and the `PATH` environment variable (all others are hidden) |
| `--log.api-enabled=jwt` | Requires a valid JWT (superuser) to access the log API |
| `--backup.api-enabled=jwt` | Requires a valid JWT (superuser) to access the backup API |

### Cluster replication constraints

In cluster mode, once there are at least **2** DBServers, the following replication constraints are
additionally appended (they are not applied in Single mode, nor while there is a single DBServer):

| Argument | Value | Notes |
|:--- |:--- |:--- |
| `--cluster.default-replication-factor` | `min(DBServers, 3)` | Default replication factor for new collections |
| `--cluster.min-replication-factor` | `2` | Collections must be replicated at least twice |
| `--cluster.write-concern` | `2` | Only set when the default replication factor is `3` (i.e. 3 or more DBServers) |

For example, with 2 DBServers the operator appends `--cluster.default-replication-factor=2` and
`--cluster.min-replication-factor=2` (no write concern); with 3 or more DBServers it appends
`--cluster.default-replication-factor=3`, `--cluster.min-replication-factor=2` and
`--cluster.write-concern=2`.

## Dependencies

- [Secured containers](secured_containers.md) must be Enabled.

## How to use

To enable this feature use the `--deployment.feature.harden` arg, which needs to be passed to the
operator (together with its dependency, `--deployment.feature.secured-containers`):

```shell
helm upgrade --install kube-arangodb \
https://github.com/arangodb/kube-arangodb/releases/download/$VER/kube-arangodb-$VER.tgz \
  --set "operator.args={--deployment.feature.secured-containers,--deployment.feature.harden}"
```
