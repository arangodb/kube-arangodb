# arango-deployment chart — TODO

## Intentionally not exposed

- `storageEngine` — not passed; leave the operator default (RocksDB).
- `sync` + `syncmasters` + `syncworkers` — dead (ArangoSync removed); deprecated in the API.
- `chaos` — chaos-monkey testing (kills pods on an interval); not a production setting.
- `id` — already exposed as the `deployment.id` block.

## Deployment-level fields

All deployment-level spec fields are now exposed (except the ones under "Intentionally not exposed").

## Chart infrastructure

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
