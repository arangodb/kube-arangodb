# arango-deployment chart — TODO

## Intentionally not exposed

- `storageEngine` — not passed; leave the operator default (RocksDB).
- `sync` + `syncmasters` + `syncworkers` — dead (ArangoSync removed); deprecated in the API.
- `chaos` — chaos-monkey testing (kills pods on an interval); not a production setting.
- `id` — already exposed as the `deployment.id` block.

## Deployment-level fields still missing

- `lifecycle`, `integration` (advanced; large nested specs)

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

- `features` currently exposes only `foxxQueues` (the sole `spec.features` field today); switch to a passthrough map once the API gains an arbitrary feature set.

- Group `int`/`bool` fields render via `with`, so an explicit `0`/`false` is omitted (treated as
  unset). Switch specific fields to a presence check if explicit `0`/`false` is ever needed.
- `gateways` resources use the uniform A4 sizing; the true A4 gateway ratio (`request_to_limit: 2.0`,
  `memory_to_cpu: 1.0`) is not special-cased.
- Decide whether to plug this chart into the platform release-chart packager
  (`images.yaml` + `values.schema.json` validation).
