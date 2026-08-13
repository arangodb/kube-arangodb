# ArangoDeployment

Renders a single `database.arangodb.com/v1` **ArangoDeployment** (Single or Cluster) managed by the
[ArangoDB Kubernetes Operator](https://github.com/arangodb/kube-arangodb).

All configuration lives under a single `deployment` block in `values.yaml` and is rendered into the
ArangoDeployment spec. The values are validated by `values.schema.json`, so typos, unknown keys and
wrong enum/type values are rejected by `helm` before anything is applied.

## Prerequisites

The operator and the ArangoDeployment CRD must already be installed (charts `kube-arangodb` and
`kube-arangodb-crd`). This chart only creates the `ArangoDeployment` resource (and, optionally, a few
secrets — see [Generated resources](#generated-resources)).

## Installing

```bash
# Cluster (default)
helm install my-db ./chart/arango-deployment --namespace my-namespace

# Single server
helm install my-db ./chart/arango-deployment --set deployment.mode=Single

# From a values file
helm install my-db ./chart/arango-deployment -f my-values.yaml
```

## Deployment modes

`deployment.mode` selects the topology, and only the groups relevant to it are rendered:

| `deployment.mode` | Rendered server groups |
|---|---|
| `Cluster` (default) | `agents`, `dbservers`, `coordinators` |
| `Single` | `single` |

The `gateways` group is rendered additionally when `deployment.platform.enabled` is `true`.

## Configuration

Every key is documented inline in [`values.yaml`](values.yaml); the sections below cover the parts
that have modes or that make Helm create additional resources.

### Image

`deployment.image` — `name`, `pullPolicy`, `pullSecrets`, and `discoveryMode` (`kubelet` uses the
resolved image sha256 in pods, `direct` uses `image.name` as-is; `""` keeps the operator default).

### Authentication — `deployment.auth.mode`

| Mode | Behaviour |
|---|---|
| `None` | Authentication disabled. |
| `Auto` (default) | The operator generates the JWT secret. |
| `Generated` | Helm creates the JWT secret `<release>-jwt` with a random token. |
| `Existing` | Use the pre-existing secret named in `auth.secretName`. |

### TLS — `deployment.tls.mode`

`None` (disabled), `Auto` (operator generates the CA, default), `Generated` (Helm creates the CA
secret `<release>-ca` with `ca.crt`/`ca.key`), or `Existing` (use `tls.secretName`).

### Encryption at rest — `deployment.encryption.mode` (Enterprise)

`None` (default), `Generated` (Helm creates a 32-byte key secret `<release>-encryption`, never
rotated), or `Existing` (use `encryption.secretName`, data key `key`, exactly 32 bytes).

> The encryption key cannot be changed after the deployment is created.

### License — `deployment.license`

Provide an existing secret via `license.secretName` (leave the inputs empty), or supply the license
as input and let Helm create the secret: set `value` for `version: v1`/`v2`, or `clientID` +
`clientSecret` for `version: manager` (License Manager credentials).

### Bootstrap passwords — `deployment.bootstrap`

- `passwordSecretNames`: map of user → secret holding the initial password (each value is `Auto`,
  `None`, or an existing secret name), e.g. `{ root: Auto }`.
- `generatedPasswords`: users for whom Helm creates a random password secret
  `<release>-<user>-password` (reused on upgrade, never rotated), e.g. `[ root ]`.

### Server groups

`agents`, `dbservers`, `coordinators`, `single`, `gateways` each map to a `ServerGroupSpec`. Notable
conventions:

- `resources.cpu` / `resources.memory` are rendered as **both** requests and limits.
- `memoryReservation` defaults to `15` (%) on the data-bearing `dbservers` and `single` groups, `0`
  elsewhere.
- `volumeClaimTemplate` (`size`, `storageClass`, `accessMode`) applies to the storage groups
  (`agents`, `dbservers`, `single`).

### Advanced deployment-level fields

Exposed under `deployment` and rendered only when set (see `values.yaml` for the full list and
defaults): `topology`, `externalAccess`, `communicationMethod`, `clusterDomain`, `rebalancer`,
`architecture`, `timezone`, `restore`, `features`, `memberPropagationMode`, `allowUnsafeUpgrade`,
`downtimeAllowed`, `disableIPv6`, `networkAttachedVolumes`, `labels`/`annotations`
(+ `*Mode` / `*IgnoreList`), `upgrade`, `rotate`, `recovery`, `database`, `timeouts`, `lifecycle`,
`integration`, `id`, and `platform`.

Tri-state flags (e.g. `allowUnsafeUpgrade`, `features.foxxQueues`) accept `true`, `false`, or `null`
(omit the field and use the operator default).

## Generated resources

Depending on the selected modes, Helm may create these secrets alongside the ArangoDeployment. They
use predefined names and are reused on upgrade (never rotated):

| Resource | Name | Created when |
|---|---|---|
| JWT secret | `<release>-jwt` | `auth.mode: Generated` |
| TLS CA secret | `<release>-ca` | `tls.mode: Generated` |
| Encryption key secret | `<release>-encryption` | `encryption.mode: Generated` |
| User password secret | `<release>-<user>-password` | user listed in `bootstrap.generatedPasswords` |
| License secret | `license.secretName` | `license.secretName` set **and** `license.value` (or `clientID`+`clientSecret`) provided |
