---
layout: page
title: Building a link
parent: Links
grand_parent: ArangoDBPlatform
nav_order: 4
---

# Building a link

A link is a program that lets AI tools execute operations on external
systems (databases, APIs, services) through a standardized job-based interface.

## What Is a link?

A link runs as a container inside a Kubernetes pod. The platform injects
an integration sidecar into the pod that provides:
- A **job queue** (MetaStore) — where submitted jobs wait to be processed
- A **file store** (StorageV2) — where the link uploads results
- An **internal HTTP API** at `http://127.0.0.1:9192` — how the link
  communicates with the platform

Your link binary is a simple loop: pick up a job, execute it, upload
results, report completion.

## How It Works

```
  AI Tool                    Platform Sidecar                Your Connector
  -------                    ----------------                --------------

  POST /connector/<name>/job
    ──────────────────────►  Stores job (Pending)
                                    │
                                    │  ◄──── POST /job/pickup ──────  polls every N seconds
                                    │        (Scheduled)
                                    │
                                    │  ◄──── POST /job/{id}/status ── Running
                                    │
                                    │        ... executes work ...
                                    │
                                    │  ◄──── POST /job/{id}/upload/result.json
                                    │
                                    │  ◄──── POST /job/{id}/status ── Completed
                                    │
  GET /connector/<name>/job/{id}
    ──────────────────────►  Returns job (Completed + result path)
```

## What Is a Job?

A job is a single unit of work — for example, "run this AQL query" or
"search this vector index". Each job has:
- A **query** — JSON input matching the link's schema
- A **status history** — tracks Pending → Scheduled → Running → Completed/Failed
- A **result path** — where output files are stored in FileStore

See [Jobs](jobs.md) for the full state machine and details.

## Authentication and RBAC

You do **not** need to implement authentication or RBAC in your connector.
The platform handles this:
- The external API (AI tool facing) goes through the gateway with authentication
- The internal API (your connector) runs on `127.0.0.1` — only accessible
  inside the pod, no auth needed
- The connector uses service credentials to access external systems

## Step 1: Create the CRD and Route

Create two Kubernetes resources in your Helm chart's `templates/` directory:

1. An `ArangoRoute` that gives your connector a user-friendly URL
2. An `ArangoPlatformLink` that registers it with the platform

```yaml
# templates/route.yaml
apiVersion: networking.arangodb.com/v1beta1
kind: ArangoRoute
metadata:
  name: {{ .Release.Name }}-route
spec:
  deployment: {{ .Values.arangodb_platform.deployment.name }}
  route:
    path: /connector/{{ .Release.Name }}/
  destination:
    path: /_integration/connector/v1/
---
# templates/connector.yaml
apiVersion: platform.arangodb.com/v1beta1
kind: ArangoPlatformLink
metadata:
  name: {{ .Release.Name }}
spec:
  type: Active
  deployment:
    name: {{ .Values.arangodb_platform.deployment.name }}
  route:
    name: {{ .Release.Name }}-route
  description: "What this link does"
  tags:
    - my-tag
  schema:
    type: object
    properties:
      myParam:
        type: string
    required:
      - myParam
  version: "1.0.0"
```

The `schema` field defines what input your connector accepts. The platform
validates submitted jobs against this schema — your link does not need
to validate the schema itself, but may optionally do so for defense in depth.

## Step 2: Implement the link Loop

Your link binary polls the internal API for jobs and processes them:

```
loop forever:
  1. POST /_internal/connector/v1/job/pickup
     → if empty: sleep 5 seconds, retry

  2. GET /_internal/connector/v1/job/{id}
     → parse job.query as JSON

  3. POST /_internal/connector/v1/job/{id}/status
     → { "state": "JOB_STATE_RUNNING", "description": "Executing..." }

  4. Execute the work (call external API, run query, etc.)

  5. POST /_internal/connector/v1/job/{id}/upload/result.json
     → upload output file(s)

  6. POST /_internal/connector/v1/job/{id}/status
     → { "state": "JOB_STATE_COMPLETED", "description": "Done" }
     or { "state": "JOB_STATE_FAILED", "description": "Error: ..." }
```

### Handling Cancellation

After picking up a job, your connector should periodically check whether
the job has been cancelled (by polling `GET /job/{id}` and checking the
state). If cancelled, stop processing and move on to the next job.
If your link does not check for cancellation, a cancelled job will
simply be ignored when it tries to update status — the update will fail
because the state transition from Cancelled is not allowed.

## Step 3: Create a Helm Chart

Your Helm chart is the link's own packaging — it deploys the link
binary alongside the CRD and route. It is **not** a chart for consumers of
the link.

The chart should include:

```
my-connector/
├── Chart.yaml              # apiVersion: v1, name must match directory
├── values.yaml             # accepts arangodb_platform.deployment.name
├── platform.yaml           # connector metadata (name, tags)
├── templates/
│   ├── connector.yaml      # ArangoPlatformLink CRD
│   ├── route.yaml          # ArangoRoute for user-friendly URL
│   └── deployment.yaml     # Deployment with sidecar labels
```

### Deployment with Sidecar Labels

```yaml
# templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}
  labels:
    profiles.arangodb.com/apply: "yes"
    profiles.arangodb.com/deployment: {{ .Values.arangodb_platform.deployment.name }}
spec:
  template:
    metadata:
      labels:
        profiles.arangodb.com/apply: "yes"
        profiles.arangodb.com/deployment: {{ .Values.arangodb_platform.deployment.name }}
    spec:
      containers:
        - name: connector
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          command: ["/bin/my-connector"]
```

The `profiles.arangodb.com` labels trigger the platform to inject the
integration sidecar into your pod.

### Standard Values

```yaml
# values.yaml
arangodb_platform:
  deployment:
    name: ""   # populated by ArangoPlatformService or helm --set
```

## Step 4: Deploy

There are two ways to deploy:

### Option A: Direct Helm Install

```bash
helm install my-connector ./my-connector \
  --namespace <namespace> \
  --set arangodb_platform.deployment.name=<deployment-name>
```

You manage the lifecycle (upgrades, rollbacks) yourself.

### Option B: ArangoPlatformService (Managed)

Upload the chart as an `ArangoPlatformChart`, then create an
`ArangoPlatformService`. The operator manages the deployment lifecycle,
including upgrades and health monitoring.

## Operational Notes

- **Crash handling**: If your connector crashes, the pod restarts
  automatically (Kubernetes restart policy). Jobs that were in Scheduled
  or Running state from the crashed handler will remain stuck until the
  handler heartbeat TTL (1 minute) expires. A future cleanup mechanism
  can return these jobs to Pending.
- **Multiple instances**: You can run multiple replicas. Job pickup is
  atomic (revision-based) — only one instance claims each job.
- **Local state**: Do not store state between jobs locally. Each job
  should be self-contained. Use the FileStore for persistent output.
- **Resource limits**: Standard Kubernetes resource limits apply to your
  container. Set them in the Deployment spec as needed.
- **Job runs forever?**: No. Each job should complete in bounded time.
  Use the `timeout` field to enforce a maximum duration. If your job
  exceeds the timeout, your connector should report it as Failed.

## Example

See the [sample AQL connector](https://github.com/arangodb/kube-arangodb-test/tree/master/tests/links/aql)
for a complete working example including the Go binary, Helm chart, and
integration test.
