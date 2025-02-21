---
layout: page
parent: CRD reference
title: ArangoJob V1
---

# API Reference for ArangoJob V1

## Spec

### .spec.arangoDeploymentName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.45/pkg/apis/apps/v1/job_spec.go#L27)</sup>

ArangoDeploymentName holds the name of ArangoDeployment

***

### .spec.jobTemplate

Type: `batch.JobSpec` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.45/pkg/apis/apps/v1/job_spec.go#L33)</sup>

JobTemplate holds the Kubernetes Job Template

Links:
* [Kubernetes Documentation](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
* [Documentation of batch.JobSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#jobspec-v1-batch)

