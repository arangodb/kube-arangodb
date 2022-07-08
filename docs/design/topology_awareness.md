# Topology awareness

## Table of contents
1. [Overview](#1)
2. [Requirements](#2)
3. [Enable/Disable topology](#3)
4. [Check topology](#4)

## Overview <a name="1"></a>

Topology awareness is responsible for the even distribution of groups of pods across nodes in the cluster.
A distribution should be done by the zone, so thanks to that if one of the zone fails there are other working pods
in different zones. For the time being, there are 3 groups of pods that can be distributed evenly
(coordinators, agents, DB servers). For each of these groups, the Kube-ArangoDB operator
tries to distribute them in different zones in a cluster, so there can not
be a situation where many pods of the same group exist in one zone and there are no
pods in other zones. It would lead to many issues when a zone with many pods failed.
When Kube-ArangoDB operator is going to add a new pod, but all zones already contain a pod of this group,
it will choose the zone with the fewest number of pods of this group.

#### Example
Let's say we have two zones (uswest-1, uswest-2) and we would like to distribute ArangoDB cluster 
with 3 coordinators, 3 agents, and 3 DB servers. First coordinator, agent, and DB server would go to random zone (e.g. uswest-1).
Second coordinator must be assigned to the `uswest-2` zone, because the zone `uswest-1` already contains one coordinator.
The same happens for the second agent and the second DB server. Third coordinator can be placed randomly 
because each of the zone contains exactly one coordinator, so after this operation one of the zone should have 2 coordinators 
and second zone should have 1 coordinator. The same applies to agents and DB servers.

According to the above example we can see that:
- coordinators should not be assigned to the same zone with other coordinators, unless ALL zones contain coordinators.
- agents should not be placed in the same zone with other agents, unless ALL zones contain agents.
- DB servers should not be placed in the same zone with other DB servers, unless ALL zones contain DB servers.

## Requirements <a name="2"></a>

- It does not work in a `Single` mode of a deployment.
  The `spec.mode` of the Kubernetes resource ArangoDeployment can not be set to `Single`.
- Kube-ArangoDB version should be at least 1.2.10 and enterprise version.

## How to enable/disable topology awareness for the ArangoDeployment <a name="3"></a>

Enable topology:
```yaml
spec:
  topology:
    enabled: true
    label: string # A node's label which will be considered as distribution affinity. By default: 'topology.kubernetes.io/zone' 
    zones: int # How many zones will be used to assign pods there. It must be higher than 0.
```

Disable topology:
```yaml
spec:
  topology:
    enable: false
```
or remove `spec.topology` object.

## How to check which ArangoDB members are assigned to the topology <a name="4"></a>

#### Topology aware

Each member should be topology aware, and it can be checked in list of conditions here `status.members.[agents|coordinators|dbservers].conditions`. 
Example:
```yaml
status:
  ...
  members:
    agents:
    - conditions:                  
         reason:  Topology awareness enabled
         status:  True
         type:    TopologyAware
```

If `status` for the condition's type `TopologyAware` is set to `false` then it is required to replace ArangoDB member.
To do so we need to set pod's annotation `deployment.arangodb.com/replace` to `true`, starting from all
coordinators which are not assigned to any zone. This situation usually happens when 
topology was enabled on an existing ArangoDeployment resource.

#### Member topology
Each member's status should have topology, and it can be checked here `status.members.[agents|coordinators|dbservers].topology` and here `status.topology`. 
Example:
```yaml
status:
  ...
  members:
    agents:
    - id: AGNT-2shphs7a
      topology:
        id:     35a61527-9d2b-49df-8a31-e62417fcd7e6
        label:  eu-central-1c
        rack:   0
  ...
  topology:
    id:     35a61527-9d2b-49df-8a31-e62417fcd7e6
    label:  topology.kubernetes.io/zone
    size:   3
    zones:
      - id: 0
        labels:
          - eu-central-1c
        members:
          agnt:
            - AGNT-2shphs7a
          ...
      - ...
  ...
```
which means that `AGNT-2shphs7a` is assigned to `eu-central-1c`.

#### Pod's labels

A pod which belongs to the member should have two new labels. 
Example:
```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    deployment.arangodb.com/topology: 35a61527-9d2b-49df-8a31-e62417fcd7e6
    deployment.arangodb.com/zone: "0"
```

#### Pod anti-affinity

A pod which belongs to the member should have a new pod anti affinity rules. 
Example:
```yaml
spec:
  affinity:
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
          - key: deployment.arangodb.com/topology
            operator: In
            values:
            - 35a61527-9d2b-49df-8a31-e62417fcd7e6
          - ...
          - key: deployment.arangodb.com/zone
            operator: In
            values:
            - "1"
            - "2"
          - ...
        topologyKey: topology.kubernetes.io/zone
      - ...
```
which means that pod can not be assigned to zone `1` and `2`.

#### Node affinity

A pod which belongs to the member can have a node affinity rules. If a pod does not have it then it will have pod affinities. 
Example:
```yaml
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
            - key: topology.kubernetes.io/zone
              operator: In
              values:
              - eu-central-1c
            - ...
        - matchExpressions:
            - key: topology.kubernetes.io/zone
              operator: NotIn
              values:
              - eu-central-1a
              - eu-central-1b
            - ...
```

#### Pod affinity

A pod which belongs to the member can have a pod affinity rules. If a pod does not have it then it will have node affinity.
Example:
```yaml
spec:
  affinity:
    podAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
          - key: deployment.arangodb.com/topology
            operator: In
            values:
            - 35a61527-9d2b-49df-8a31-e62417fcd7e6
          - key: deployment.arangodb.com/zone
            operator: In
            values:
            - "1"
          - ...
        topologyKey: topology.kubernetes.io/zone 
```
