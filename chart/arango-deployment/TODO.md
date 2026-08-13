# arango-deployment chart — TODO

## Intentionally not exposed

- `storageEngine` — not passed; leave the operator default (RocksDB).
- `sync` + `syncmasters` + `syncworkers` — dead (ArangoSync removed); deprecated in the API.

## Deployment-level fields still missing

- top-level `annotations` / `labels` (+ `annotationsMode` / `labelsMode` / `*IgnoreList`)
- `features` (per-deployment feature flags)
- `memberPropagationMode`
- `allowUnsafeUpgrade`, `downtimeAllowed`, `disableIPv6`, `networkAttachedVolumes`
- `upgrade`, `rotate`, `recovery`, `chaos`, `timeouts`, `lifecycle`, `id`, `database`, `integration`

## Chart infrastructure

- **`values.schema.json`**: rewrite to match the current `deployment` structure. It is stale
  (old top-level keys, `additionalProperties: false`) and currently **rejects** the values, so it
  was left untouched — `helm template`/`helm lint` only pass when it is moved aside.
- **`README.md`**: document values, modes and usage examples.
- **`NOTES.txt`**: post-install hints (connection info, generated secret names).
- **`Chart.yaml`**: set `appVersion`, review `version`.
- Add a `CHANGELOG.md` entry when this is merged.
- Regenerate the CRD schema after marking `syncmasters`/`syncworkers` deprecated in the API
  (doc comment flows into the generated schema descriptions).

## Polish / open questions

- Group `int`/`bool` fields render via `with`, so an explicit `0`/`false` is omitted (treated as
  unset). Switch specific fields to a presence check if explicit `0`/`false` is ever needed.
- `gateways` resources use the uniform A4 sizing; the true A4 gateway ratio (`request_to_limit: 2.0`,
  `memory_to_cpu: 1.0`) is not special-cased.
- Decide whether to plug this chart into the platform release-chart packager
  (`images.yaml` + `values.schema.json` validation).
