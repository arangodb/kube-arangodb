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
| Dependency graph (`go.mod`, `go.sum`) | yes | `dependency-cve-scan`, `govulncheck`, `nightly-dependency-report` |
| Published operator image, Debian base | yes | `image-cve-scan`, `image-cve-scan-nightly` |
| Published operator image, UBI9 base | yes | `image-cve-scan-ubi`, `image-cve-scan-ubi-nightly` |
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
| `dependency-cve-scan` | `trivy fs`, dependency CVEs | gate band fixable CRITICAL,HIGH; report band all severities | warn-only by team decision, see below |
| `image-cve-scan` | `trivy image` over the published Debian image, comprehensive | same bands | warn-only by team decision |
| `image-cve-scan-ubi` | `trivy image` over the published UBI9 image, comprehensive | same bands | warn-only by team decision |
| `govulncheck` | reachability over the real toolchain, symbol level | any reachable symbol | warn-only by team decision |
| `secret-scan` | `trivy fs`, secret scanner, `testdata/` back in scope | CRITICAL,HIGH,MEDIUM,LOW | report-only, see below |
| `misconfig-scan` | `trivy fs`, misconfiguration scanner over charts, manifests, Dockerfiles | CRITICAL,HIGH gate band | report-only; the Dockerfile credential guard in the same job is blocking |
| `sast-scan` | Semgrep CE, full repo, `p/default` plus the org rules | reports every level | report-only, see below |
| `sast-scan-diff` | Semgrep CE, diff against `master` | fails at WARNING on new findings only | blocking |
| `disposition-check` | expiry and integrity of every suppression | any undated, expired or out-of-window disposition, any SAST scope reduction | blocking |
| `required-checks-drift` | committed required-check list versus live protection | drift in either direction | warn-only until protection lists the gate contexts |
| `nightly-self-test` | positive control against a digest-pinned known-vulnerable image | fails if the scanner comes back clean | blocking in the nightly |
| `nightly-dependency-report` | all severities, unfixed included, licences, full SBOM, KEV correlation | none | report-only |
| `misconfig-report-nightly` | the same misconfiguration scan widened to MEDIUM and LOW | none | report-only |
| `sast-report-nightly` | Semgrep at INFO | none | report-only |

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

## Why the scanning tiers are not enforcing yet

Each one has a written flip criterion. None of them is "when someone gets round
to it".

- **`dependency-cve-scan`** is `fail-on-findings: false` because that is the
  posture `#2118` was reviewed and merged with. The fixable CRITICAL,HIGH band
  holds exactly one finding at this commit, `CVE-2026-56852` in
  `golang.org/x/text` v0.37.0 (fixed in 0.39.0). Once that bump lands the band is
  empty, so the flip is one line here plus one line in
  `.github/required-checks.txt`, and it does not create a gate that is red on
  arrival.
- **`image-cve-scan` and `image-cve-scan-ubi`** are warn-only for a reason a
  pull request cannot fix: the released 1.4.4 Debian image carries five fixable
  HIGH findings in the compiled `arangodb_operator` binary, four of which are
  already fixed on `master` (`otel/sdk`, `grpc`, `oras-go/v2`, `x/text`) and one
  of which is the Go stdlib the binary was built with (`CVE-2026-39822`, built
  with 1.26.4, fixed in 1.26.5). The UBI9 variant adds seven more fixable HIGH
  findings from its own package set. These tiers can only go blocking after a
  release ships the fixes, which is also exactly what makes them worth having:
  they measure the artifact customers pull, not the branch.
- **`govulncheck`** reports five symbol-reachable findings. One is fixable in
  range (`GO-2026-5970`, `x/text`). Three are containerd CRI checkpoint issues
  whose fix exists only on the `github.com/containerd/containerd/v2` module path,
  a major-version move on an indirect dependency. The fifth, `GO-2026-5932`, is
  the unmaintained `golang.org/x/crypto/openpgp` package, which has no fix at
  all. It flips when those four are resolved upstream or recorded as dated
  owner-approved acceptances.
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
three ignorefile paths and audits every source extension for inline
suppressions, so reintroducing one unreviewed fails the build.

Rules for any new disposition:

- Inline SAST suppressions are
  `# nosemgrep: <rule-id> -- reason -- review by: YYYY-MM-DD`. The date is
  mandatory and enforced.
- A CVE or GHSA disposition additionally needs `severity:` and `detected:`, and
  its review date must fall inside that severity's remediation window.
- The first accepted risk that needs gate-time subtraction lands
  `.circleci/security-waivers.yaml` together with a `validate-waivers` step and
  the `emit-vex` step that publishes the waiver set as OpenVEX. The file is
  deliberately absent until then: a waivers file with zero entries is a hard
  failure on trivy-scan 1.1.2, which is correct, because deleting the entries
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
| On the CISA KEV catalogue | the CISA due date, and no more than 90 days regardless of severity |

A disposition may not be renewed indefinitely: 366 days from the day it is
written is the ceiling on any review-by date, enforced by `disposition-check`.

## Branch protection and required status checks

`master` is protected by classic branch protection: one approving review, code
owner review, signed commits required, strict status checks, and one required
status check (`ci/circleci: check-code`). The two rules that reach the branch
through the ruleset API are the enterprise-sourced deletion and
non-fast-forward rules; neither can carry status checks.

The authoritative list of contexts this repository expects lives in
`.github/required-checks.txt`, and `required-checks-drift` compares it against
the live protection on every push.

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

Nothing published from this repository is signed today, and no SLSA level is
claimed.

The chain is wired: both image jobs read the digest of the image they scanned out
of the trivy report and, behind the `sign-artifacts` pipeline parameter, sign
that digest and attach the CycloneDX SBOM and an in-toto provenance predicate.
Signing is bound to the digest and never to the tag, because a tag can be
re-pushed and a signature over a tag proves nothing about the artifact that runs.

What blocks the flip is not the config. This CircleCI project neither builds nor
pushes the operator image, the Jenkins release job does, and the project holds no
Docker Hub push credential, so cosign has nowhere to write the signature and
attestation layers. Resolving it means either giving this project a scoped push
credential or moving the signing step into the Jenkins release job. The
Rekor-publicity decision (`tlog-upload`) is a separate org-level call; for a
public image and public digests the recommendation is keyless with a transparency
log entry.

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

   A legacy in-config `triggers: schedule` is deliberately not used. It is read
   only from the top-level config, so a block in the continuation config
   validates and never runs, and it sets no pipeline parameters, so it cannot
   select between the two continuation paths and a pipeline that continues twice
   fails. Elsewhere in this fleet a legacy schedule was also observed to stop
   firing for eight months without anyone noticing.
2. **Evidence is CircleCI artifacts, not an archive of record.** The
   `publish-evidence` and `dtrack-upload` steps are wired and inert: they log why
   they did nothing and exit 0 while `evidence-bucket` and `dtrack-url` are
   empty. An S3 bucket with a COMPLIANCE-mode retention period and a
   Dependency-Track 5 host are org-owner actions. Until they exist, scan evidence
   expires with the CircleCI artifact retention.
3. **`helm lint` is not in the pipeline, because two charts do not lint.**
   `chart/kube-arangodb-crd/Chart.yaml` is a Helm 2 era chart: it carries a
   `tillerVersion` key and no `apiVersion`, which `helm lint` rejects outright.
   `chart/platform-storage/templates/passwords.yaml` fails the linter's YAML
   parse. Both are product-chart fixes rather than CI changes, so
   `trivy-scan/helm-lint` is deliberately absent instead of being pointed at a
   subset of charts. The coverage that step exists to protect is separately
   proven: trivy loaded all six charts and reported 114 Helm targets, so the
   misconfiguration scan is not silently reading zero Helm files. Adding
   `helm-lint` is the same change that fixes those two charts.
4. **SAST runs with `strict: false`.** Semgrep records 288 analysis errors on
   this checkout: 214 because `chart/*/templates/*.yaml` are Go templates rather
   than YAML documents, 61 partial parses in the generated clientset under
   `pkg/generated`, and 13 elsewhere. With `strict: true` the orb correctly fails
   on an incomplete analysis, so every SAST job would be red for a scanner
   limitation on Helm sources rather than for a finding. The blind spot is not
   hidden: the orb prints the error count and type breakdown on every run, and
   the chart templates are covered by the misconfiguration tier, which renders
   them properly.
5. **The release pipeline is not observed by these gates.** The published image
   is built by Jenkins (`Jenkinsfile.groovy`). Everything here inspects either the
   source tree or the image after it is published, which is why the image tiers
   can see fixable HIGH findings that `master` has already fixed. Closing this
   means gating inside the release job, which is work in that pipeline.
6. **The image tag is derived from `VERSION`.** Between a `VERSION` bump and the
   corresponding Docker Hub publish, the tag does not exist and the image jobs
   fail on a missing image rather than on a finding. This is pre-existing
   behaviour from `#2132`; the alternative considered there, `latest`, was found
   several releases stale and `latest-ubi` does not exist at all.
7. **`govulncheck` runs twice.** `check-code` runs `make vulncheck-optional` on
   pull requests only, and the `govulncheck` job runs the same analysis with a
   stable status-check context, an artifact, and coverage on nightly and tag
   pipelines. Consolidating them edits the body of the one job branch protection
   currently requires, so it is an owner item rather than part of this change.
8. **No lockfile-currency check.** `go mod verify` proves the module graph
   matches `go.sum`. `make ci-check` runs `tidy` and fails on a dirty tree, which
   covers the `go.mod` side on pull requests but not on tag pipelines.
9. **No OpenVEX document is published.** It is generated from the waiver set, and
   there is no owner-approved waiver to make a statement about. Publishing an
   empty document would look like coverage while asserting nothing. FedRAMP Rev-5
   makes risk-based VDR/VER mandatory on 2026-12-07, so this lands with the first
   accepted risk at the latest.

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
| `arangodb/trivy-scan` | 1.1.2 | carries the Trivy pin (v0.72.0), `disposition-check`, `govulncheck`, `kev-epss-check` |
| `arangodb/semgrep-scan` | 1.0.0 | carries the Semgrep pin (1.168.0) and the org rule pack |
| `arangodb/supply-chain` | 1.0.0 | SBOM validation, cosign signing, evidence export, Dependency-Track ingest |
| `circleci/slack` | 4.1.4 | the version the previous `@4.1` float already resolved to |
| `circleci/path-filtering` | 1.2.0 | unchanged, previously already exact |
| `circleci/continuation` | 2.0.1 | drives the nightly continuation |
| Positive-control fixture | `python:3.9-slim@sha256:2d97f6910b16bd338d3060f261f53f144965f755599aab1acda1e13cf1731b1b` | digest-pinned so the control cannot drift |

Last reviewed: 2026-07-30.
