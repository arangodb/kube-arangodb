# Security Policy

`kube-arangodb` is the ArangoDB Kubernetes Operator. It ships container images
(`arangodb/kube-arangodb`, `arangodb/kube-arangodb-enterprise` and their `-ubi`
and `-arm64` variants on Docker Hub) and Helm charts that customers install into
their own clusters. That shapes everything below, so read the scope statement
before the controls.

## Reporting a vulnerability

Report suspected security vulnerabilities to
**[security@arango.ai](mailto:security@arango.ai)**. You will receive a response
within 72 hours. If the issue is confirmed, a patch follows as soon as
practical, depending on complexity.

GitHub private vulnerability reporting is enabled on this repository, so a
report can also be opened from the Security tab. Please do not open a public
issue for a suspected vulnerability.

Patches are released for the versions listed in the
[ArangoDB product support and end-of-life announcements](https://arango.ai/arangodb-product-support-end-of-life-announcements/).

## Scope

| Surface | Exists here | Covered by |
| --- | --- | --- |
| Dependency graph (`go.mod`, `go.sum`) | yes | `dependency-cve-scan`, `vulncheck`, `nightly-dependency-report` |
| Enterprise operator image, Debian base | yes | `image-cve-scan`, `image-cve-scan-nightly` |
| Enterprise operator image, UBI9 base | yes | `image-cve-scan-ubi`, `image-cve-scan-ubi-nightly` |
| Community image and the `-arm64` variants | yes | nothing here, see Known gaps |
| First-party Go code (`pkg/`, `cmd/`, `integrations/`, `internal/`) | yes | `sast-scan`, `sast-scan-diff` |
| Helm charts and Kubernetes manifests | yes | `misconfig-scan` |
| Dockerfiles | yes | `misconfig-scan`, including the deterministic credential guard |
| Committed secrets | yes | `secret-scan`, plus GitHub secret scanning and push protection |
| The build that produces the released image | Jenkins, not this config | nothing here, see Known gaps |

## What is scanned automatically

Everything runs on the shared in-house orbs, pinned to exact versions. A float
(`@1`, `@4.1`) resolves at config-compile time, so a future orb publish would
change what runs here with no diff and no review.

| Job | Tier | Band | Verdict |
| --- | --- | --- | --- |
| `dependency-cve-scan` | `trivy fs`, dependency CVEs | gate band fixable CRITICAL,HIGH; report band all severities | blocking, see below |
| `image-cve-scan` | `trivy image` over the published Debian image, comprehensive | same bands | warn-only by team decision |
| `image-cve-scan-ubi` | `trivy image` over the published UBI9 image, comprehensive | same bands | warn-only by team decision |
| `vulncheck` | reachability, the repository's own `make vulncheck-optional` | any reachable symbol | warn-only, pre-existing posture, see below |
| `secret-scan` | `trivy fs`, secret scanner, `testdata/` back in scope | CRITICAL,HIGH,MEDIUM,LOW | report-only, see below |
| `misconfig-scan` | `trivy fs`, misconfiguration scanner over charts, manifests, Dockerfiles | CRITICAL,HIGH gate band | report-only; the Dockerfile credential guard in the same job is blocking |
| `sast-scan` | Semgrep CE, full repo, `p/default` plus the org rules | reports every level | report-only, see below |
| `sast-scan-diff` | Semgrep CE, diff against `master` | fails at WARNING on new findings only | blocking |
| `disposition-check` | expiry and integrity of every suppression | any undated, expired or out-of-window disposition, any SAST scope reduction | blocking |
| `required-checks-drift` | committed required-check list versus live protection | drift in either direction | warn-only, and currently UNVERIFIED, see Known gaps 2 |
| `nightly-self-test` | positive control against a digest-pinned known-vulnerable image | fails if the scanner comes back clean | blocking in the nightly |
| `nightly-dependency-report` | all severities, unfixed included, licences, full SBOM, KEV correlation | none | report-only |
| `misconfig-report-nightly` | the same misconfiguration scan widened to MEDIUM and LOW | none | report-only |
| `sast-report-nightly` | Semgrep at INFO | none | report-only |
| `vulncheck-nightly` | the same reachability analysis between pull requests | none | report-only |

The gate band and the report band are separate on purpose. `--severity` is
applied at scan time, so without `report-severity` the MEDIUM, LOW and UNKNOWN
findings would be physically absent from the JSON, the table, the JUnit report
and the SBOM, and there would be no evidence of what the gate did not count.

## The two-tier gate policy

1. Blocking dependency and image tiers gate on fixable CRITICAL and HIGH only.
   UNKNOWN is report-tier: in the Go advisory data it means the source assigned
   no CVSS at all, which is a property of the data and not a risk statement.
2. Dev-only and test-only dependencies are reported, never gated. A Go module
   graph has no dev-only group, so on this repository the distinction is moot and
   `--include-dev-deps` was measured to change nothing.
3. MEDIUM, LOW, licence and misconfiguration findings are visibility tiers. They
   are patched on the SLA below; they do not block a merge.
4. Any aging or SLA check warns and annotates. It never fails a build.
5. `disposition-check` is blocking. That is not gate strictness, it is whether
   the exception path can be trusted at all.

## Enforcement status of the scanning tiers

Every non-blocking tier has a written flip criterion. None of them is "when
someone gets round to it".

- **`dependency-cve-scan`** now blocks, per the criterion it was introduced
  with: the `x/text` bump (`CVE-2026-56852`) landed, and the fixable
  CRITICAL,HIGH band is measurably empty at this commit, so the gate is not
  red on arrival. Making it a required context on `master` is the remaining
  admin step.
- **`image-cve-scan` and `image-cve-scan-ubi`** are warn-only for a reason a
  pull request cannot fix: the released 1.4.4 Debian image carries five fixable
  HIGH findings in the compiled `arangodb_operator` binary, four of which are
  already fixed on `master` (`otel/sdk`, `grpc`, `oras-go/v2`, `x/text`) and one
  of which is the Go stdlib the binary was built with (`CVE-2026-39822`, built
  with 1.26.4, fixed in 1.26.5). The UBI9 variant adds seven more fixable HIGH
  findings from its own package set. These tiers can only go blocking after a
  release ships the fixes, which is also exactly what makes them worth having:
  they measure the artifact customers pull, not the branch.
- **`vulncheck`** is the repository's own reachability analysis, and it is not new
  in this change. `kube-arangodb` has run `govulncheck` through the Makefile's
  `vulncheck` and `vulncheck-optional` targets, with the repository's build tags
  and a pinned tool, since before this pull request. What this change adds is
  enforcement and freshness, not capability: the single invocation moves out of
  `check-code` into its own job, so it has a status context, a log of its own, an
  artifact and a nightly run, and the pinned tool moves from v1.1.4 to v1.6.0.

  It stays on the `-optional` target, whose leading `-` makes `make` ignore the
  exit status, because `make vulncheck` exits non-zero on four reachable
  findings from two modules, none fixable in range (the `x/text` finding
  `GO-2026-5970` was cleared by the 0.39.0 bump):

  | Finding | Module | Fix |
  | --- | --- | --- |
  | `GO-2026-5932` | `golang.org/x/crypto` v0.53.0 | none, `openpgp` is unmaintained |
  | `GO-2026-5622` | `containerd` v1.7.33 | only on the `containerd/v2` module path |
  | `GO-2026-5338` | `containerd` v1.7.33 | only on the `containerd/v2` module path |
  | `GO-2026-5064` | `containerd` v1.7.33 | only on the `containerd/v2` module path |

  In CI the count is higher again: the executor image is `cicd/golang:1.25.9`
  while `go.mod` declares `toolchain go1.25.12`, so reachable Go standard-library
  findings fixed in 1.25.10 are counted as well (for example `GO-2026-4918` in
  `net/http`). Those disappear with a rebuild on a current image and need no
  dependency change.

  The flip is one word in `.circleci/continue_config.yml`, `vulncheck` instead of
  `vulncheck-optional`, plus the context in `.github/required-checks.txt`. Its
  criterion is a decision rather than a date: bump the executor image to at least
  1.25.12, and either move off `containerd` v1 and `x/crypto/openpgp` or record
  them as dated acceptances.
- **`secret-scan`** is report-only because disabling trivy's built-in `tests`
  allow-rule surfaced nine HIGH `private-key` findings in two committed
  agency-state fixtures,
  `pkg/deployment/agency/state/testdata/sync.source.json` and `sync.target.json`.
  They are captured dumps of an arangosync deployment's agency, so the same EC
  keys appear three times per file (agency view, raft log entry, compacted
  readDB). What happens to them, redaction or rotation plus a dated acceptance,
  is the owning team's call. It flips when they are dispositioned.
- **`sast-scan`** is `fail-on-severity: NONE` because the full-repo band holds
  nine pre-existing ERROR findings: three `jwt-go-none-algorithm` in
  `pkg/util/token`, three `missing-unlock-before-return` in
  `integrations/storage/v1/shared/s3`, one `dangerous-exec-command` in
  `pkg/debug_package`, one generic-secret match in `docs/platform.sso.openid.md`
  and one missing-user in `Dockerfile`. It flips when their owners clear them or
  give them dated inline dispositions. New code is not waiting on that:
  `sast-scan-diff` gates it at WARNING today.

A gate that is permanently red on findings its owner cannot fix gets switched
off, which is worse than a warn-only gate that is read.

## Accepted-risk dispositions

There is no `.trivyignore`, no `.trivyignore.yaml`, no waivers file and no inline
`nosemgrep` comment in this repository, and that is the honest state: the open
findings above are reported, not waived. `disposition-check` still lists all
three ignorefile paths and audits the tracked source and doc extensions for
inline suppressions, so reintroducing one unreviewed fails the build.

Rules for any new disposition:

- Inline SAST suppressions take the form
  `nosemgrep: <rule-id> -- reason -- review by: YYYY-MM-DD` after the
  comment marker. The date is mandatory and enforced.
- A CVE or GHSA disposition additionally needs `severity:` and `detected:`, and
  its review date must fall inside that severity's remediation window.
- The first accepted risk that needs gate-time subtraction lands
  `.circleci/security-waivers.yaml` together with a `validate-waivers` step and
  the `emit-vex` step that publishes the waiver set as OpenVEX. The file is
  deliberately absent until then: a waivers file with zero entries is a hard
  failure on the pinned orb, which is correct, because deleting the entries
  must not silently widen every gate.
- `trivy-secret.yaml` allow-rules carry a review date as well, checked as part of
  the same job.

## Remediation SLAs

Windows run from detection, and they are the same numbers the `poam-windows`
parameter in `.circleci/continue_config.yml` enforces, so this table and the
build agree by construction.

| Severity | Window |
| --- | --- |
| CRITICAL | 30 days |
| HIGH | 30 days |
| MEDIUM | 90 days |
| LOW | 180 days |
| On the CISA KEV catalogue | the CISA due date, and no more than 14 days from detection |

A disposition may not be renewed indefinitely: 366 days from the day it is
written is the ceiling on any review-by date, enforced by `disposition-check`.

## Branch protection and required status checks

`master` is protected by classic branch protection: one approving review, code
owner review, signed commits required, strict status checks, and one required
status check (`ci/circleci: check-code`). The two rules that reach the branch
through the ruleset API are the enterprise-sourced deletion and
non-fast-forward rules; neither can carry status checks.

The authoritative list of contexts this repository expects lives in
`.github/required-checks.txt`. `required-checks-drift` reads it on every push and
compares it against the effective branch rules and the classic protection object,
but only when a `GITHUB_TOKEN` is in the job environment. No context supplies one
today, so the job prints the committed contexts and says the comparison is
UNVERIFIED rather than reporting a false match. See Known gaps 2.

READ THE SUFFIX WARNING in that file before editing it. CircleCI appends an
instance suffix (`-1`, `-2`) to the status context of a job it considers
duplicated across the config, and the numbering is not predictable from the
config text. Every workflow entry in `continue_config.yml` therefore carries an
explicit unique `name:`, and `check-code` deliberately appears in one workflow
only so that the one string protection already requires cannot move.

Adding the gate contexts to protection is a repository-admin action, and
`required-checks-drift` is warn-only until it is done. Making that job itself
required is the last step of the sequence.

## Artifact signing, attestation and SLSA level

Nothing published from this repository is signed today, no SLSA level is
claimed, and nothing here pretends otherwise: no signing is wired in this
pipeline. This CircleCI project neither builds nor pushes the operator image
(the Jenkins release job does) and holds no Docker Hub push credential, so
cosign would have nowhere to write signature or attestation layers. Signing
belongs where the image is pushed; landing it in the Jenkins release job,
bound to the pushed digest (never the tag, which can be re-pushed), is the
plan of record, together with the org-level Rekor-publicity decision.

Provenance for consumers today is therefore: the release tag, the published
digest, and the scan evidence attached to the pipeline that inspected it.

## Known gaps

Each of these is a deliberate, named gap rather than an oversight.

1. **The nightly tier does not run until a schedule is created.** It is selected
   by the `nightly` pipeline parameter, and this project has no Scheduled
   Pipeline: its v2 schedule collection is empty. The config is complete on both
   sides, so this is one owner action and no further config change:

   ```
   curl -X POST https://circleci.com/api/v2/project/gh/arangodb/kube-arangodb/schedule \
     -H "Circle-Token: $TOKEN" -H "Content-Type: application/json" \
     -d '{"name":"nightly-security-rescan","description":"Nightly CVE, SAST and disposition rescan","attribution-actor":"system","parameters":{"branch":"master","nightly":true},"timetable":{"per-hour":1,"hours-of-day":[3],"days-of-week":["MON","TUE","WED","THU","FRI","SAT","SUN"]}}'
   ```

   `nightly` is declared in both `.circleci/config.yml` and
   `.circleci/continue_config.yml`, set by the schedule, and passed on by nobody.
   CircleCI forwards a trigger-time pipeline parameter into the continued config,
   so the continued config must declare it (otherwise the continuation fails to
   compile with `Unexpected argument(s): nightly`) and the setup job must not pass
   it again (otherwise the continuation API answers `Conflicting pipeline
   parameters`). Both configs validate under either mistake, which is why the
   nightly path was exercised on the branch by triggering a pipeline with
   `nightly: true` rather than reasoned about.

   A legacy in-config `triggers: schedule` is deliberately not used. It is read
   only from the top-level config, so a block in the continuation config
   validates and never runs, and it sets no pipeline parameters, so it cannot
   select between the two continuation paths and a pipeline that continues twice
   fails. Elsewhere in this fleet a legacy schedule was also observed to stop
   firing for eight months without anyone noticing.
2. **Drift detection is wired but unverified.** `required-checks-drift` needs a
   read-only `GITHUB_TOKEN` with `administration:read` to read the live branch
   rules, and no CircleCI context supplies one to this project. Observed on this
   branch (build 6874): the job prints the committed contexts and reports
   `drift is UNVERIFIED on this run`. It fails open by design, because a drift job
   that cannot read the live state must not claim a match. Creating the context is
   an owner action; attaching it to this job is then one line.
3. **Evidence is CircleCI artifacts, not an archive of record.** Durable
   evidence export and Dependency-Track ingestion are not wired: the S3 bucket
   (COMPLIANCE-mode retention) and the Dependency-Track 5 host are org-owner
   actions that do not exist yet, and wiring dry against them was reviewed out
   (the inert steps hard-failed once armed, on non-sha256 subjects and a
   missing AWS CLI). The wiring lands together with the resources. Until then,
   scan evidence expires with the CircleCI artifact retention.
4. **`helm lint` is not in the pipeline, because two charts do not lint.**
   `chart/kube-arangodb-crd/Chart.yaml` is a Helm 2 era chart: it carries a
   `tillerVersion` key and no `apiVersion`, which `helm lint` rejects outright.
   `chart/platform-storage/templates/passwords.yaml` fails the linter's YAML
   parse. Both are product-chart fixes rather than CI changes, so
   `trivy-scan/helm-lint` is deliberately absent instead of being pointed at a
   subset of charts. The coverage that step exists to protect is separately
   proven: trivy loaded all six charts and reported 114 Helm targets, so the
   misconfiguration scan is not silently reading zero Helm files. Adding
   `helm-lint` is the same change that fixes those two charts.
5. **SAST runs with `strict: false`.** Semgrep records 288 analysis errors on
   this checkout: 214 because `chart/*/templates/*.yaml` are Go templates rather
   than YAML documents, 61 partial parses in the generated clientset under
   `pkg/generated`, and 13 elsewhere. With `strict: true` the orb correctly fails
   on an incomplete analysis, so every SAST job would be red for a scanner
   limitation on Helm sources rather than for a finding. The blind spot is not
   hidden: the orb prints the error count and type breakdown on every run, and
   the chart templates are covered by the misconfiguration tier, which renders
   them properly.
6. **The release pipeline is not observed by these gates.** The published image
   is built by Jenkins (`Jenkinsfile.groovy`). Everything here inspects either the
   source tree or the image after it is published, which is why the image tiers
   can see fixable HIGH findings that `master` has already fixed. Closing this
   means gating inside the release job, which is work in that pipeline.
7. **The image tag is derived from `VERSION`.** Between a `VERSION` bump and the
   corresponding Docker Hub publish, the tag does not exist and the image jobs
   fail on a missing image rather than on a finding. This is pre-existing
   behaviour from `#2132`; the alternative considered there, `latest`, was found
   several releases stale and `latest-ubi` does not exist at all.
8. **The CircleCI Go executor image is behind the module's toolchain.** It is
   `cicd/golang:1.25.9`; `go.mod` declares `toolchain go1.25.12`. That is what
   makes `govulncheck` report reachable Go standard-library findings in CI that do
   not exist under a current toolchain, and it means the binaries `make bin`
   produces in CI carry stdlib issues fixed in 1.25.10. Bumping the image is an
   owner action outside this repository.
9. **No lockfile-currency check.** `go mod verify` proves the module graph
   matches `go.sum`. `make ci-check` runs `tidy` and fails on a dirty tree, which
   covers the `go.mod` side on pull requests but not on tag pipelines.
10. **No OpenVEX document is published.** It is generated from the waiver set, and
   there is no owner-approved waiver to make a statement about. Publishing an
   empty document would look like coverage while asserting nothing. FedRAMP Rev-5
   makes risk-based VDR/VER mandatory on 2026-12-07, so this lands with the first
   accepted risk at the latest.
11. **The community image and the `-arm64` variants are not scanned.** Only
   `arangodb/kube-arangodb-enterprise` (Debian and UBI9) has image tiers. The
   community `arangodb/kube-arangodb` image and both `-arm64` variants ship the
   same operator binaries from the same tree, so the dependency and SAST tiers
   cover their contents indirectly, but their base layers are unobserved.
   Adding instances is one workflow entry per image once an owner sizes the
   scan cost.

### Tooling choices and cost

Everything gating this repository is free for internal use: Trivy (Apache-2.0)
and Semgrep Community Edition (LGPL-2.1) both cost $0 at any scale, and the orbs
wrapping them are in-house. The paid alternatives are deliberately not purchased
and none of them beats the free baseline for this repository's surface: Semgrep
AppSec Platform (per-contributor subscription, adds the hosted findings backend
and the pro engine), Aqua Platform (the commercial tier above Trivy), and GitHub
Code Security (per-committer subscription; it is what SARIF upload to code
scanning would need, which is why the SARIF file stays a job artifact here).
Revisit only with a named requirement the free tier cannot meet.

## Pinned components

Bumping any of these is a reviewed one-line change.

| Component | Pin | Notes |
| --- | --- | --- |
| `arangodb/trivy-scan` | 1.2.0 | carries the Trivy pin (v0.72.0), `disposition-check`, `govulncheck`, `kev-epss-check`; 1.2.0 also sets the FedRAMP waiver-window defaults |
| `arangodb/semgrep-scan` | 1.0.0 | carries the Semgrep pin (1.168.0) and the org rule pack |
| `arangodb/supply-chain` | 1.0.0 | SBOM validation (`validate-sbom`); its signing and evidence commands are not used here, see Artifact signing |
| `circleci/slack` | 4.1.4 | the version the previous `@4.1` float already resolved to |
| `golang.org/x/vuln` (`govulncheck`) | v1.6.0, in the Makefile as `GOVULNCHECK_VERSION` | pinned by TAG. golang/vuln stopped publishing GitHub Releases at v1.1.4 while it kept tagging, so anything resolving "latest" from the releases API reports v1.1.4 as current. Do not correct it back. |
| `circleci/path-filtering` | 1.2.0 | unchanged, previously already exact |
| `circleci/continuation` | 2.0.1 | drives the nightly continuation |
| Positive-control fixture | `python:3.9-slim@sha256:2d97f6910b16bd338d3060f261f53f144965f755599aab1acda1e13cf1731b1b` | digest-pinned so the control cannot drift |

Last reviewed: 2026-07-30.
