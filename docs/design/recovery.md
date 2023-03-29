# Manual Recovery

## Overview
Let's consider a situation where we had a ArangoDeployment in Cluster mode (3 DbServers, 3 Coordinators, 3 Agents)
with Local storage attached (only one K8s Node in the K8s cluster).

Due to some reason the ArangoDeployment was deleted (e.g. ETCD storage has been wiped out) and we want to recover it.
Fortunately, we have a backup of the data on the disk.

To recover the ArangoDeployment we need to:
1. Create PV and PVC for each member with persistent storage (agent, dbservers, single)
2. Create a new ArangoDeployment with the same members IDs

## Local storage data

We have a members (Agents & DbServers) data in the following directories:
```bash
> ls -1 /var/data/
f9rs2htwc9e0bzme
fepwdnnbf0keylgx
gqnkahucthoaityt
vka6ic19qcl1y3ec
rhlf8vixbsbewefo
rlzl467vfgsdpofu
```

To find out the name of the members to which data should be attached,
we need to check the `UUID` file content in each directory:
```bash
> cat /var/data/*/UUID
AGNT-pntg5yc8
AGNT-kfyuj8ow
AGNT-bv5rofcz
PRMR-9xztmg4t
PRMR-l1pp19yl
PRMR-31akmzrp
```

## Initial ArangoDeployment

Here is an example of the initial ArangoDeployment before deletion:
```yaml
cat <<EOF | kubectl apply -f -
apiVersion: "database.arangodb.com/v1"
kind: "ArangoDeployment"
metadata:
  name: "cluster"
spec:
  externalAccess:
    type: NodePort
  mode: Cluster
  agents:
    volumeClaimTemplate:
      spec:
        storageClassName: my-local-ssd
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        volumeMode: Filesystem
  dbservers:
    volumeClaimTemplate:
      spec:
        storageClassName: my-local-ssd
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        volumeMode: Filesystem
EOF
```

## Create PV and PVC

1. We need to create ArangoLocalStorage first:
    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: "storage.arangodb.com/v1alpha"
    kind: "ArangoLocalStorage"
    metadata:
      name: "local-storage"
    spec:
      storageClass:
        name: my-local-ssd
        isDefault: true
      localPath:
      - /mnt/data
    EOF
    ```
2. Now create PV and PVC for every directory listed above
   - Agents - here is an example for `AGNT-pntg5yc8`(`f9rs2htwc9e0bzme` directory)
     - PV
        ```yaml
        cat <<EOF | kubectl apply -f -
        apiVersion: "v1"
        kind: PersistentVolume
        metadata:
          labels:
            arango_deployment: cluster
            role: agent
          name: agent-pntg5yc8-f9rs2htwc9e0bzme
        spec:
          accessModes:
          - ReadWriteOnce
          capacity:
            storage: 1Gi
          local:
            path: /mnt/data/f9rs2htwc9e0bzme
          persistentVolumeReclaimPolicy: Retain
          storageClassName: my-local-ssd
          volumeMode: Filesystem
          nodeAffinity:
            required:
              nodeSelectorTerms:
              - matchExpressions:
                - key: kubernetes.io/hostname
                  operator: In
                  values:
                  - minikube
        EOF
        ```
     - PVC
         ```yaml
         cat <<EOF | kubectl apply -f -
         apiVersion: v1
         kind: PersistentVolumeClaim
         metadata:
           labels:
             app: arangodb
             arango_deployment: cluster
             role: agent
           name: agent-pntg5yc8
         spec:
           accessModes:
           - ReadWriteOnce
           resources:
             requests:
               storage: 1Gi
           storageClassName: my-local-ssd
           volumeMode: Filesystem
           volumeName: agent-pntg5yc8-f9rs2htwc9e0bzme
         EOF
         ```
   - DbServers - here is an example for `PRMR-9xztmg4t` (`vka6ic19qcl1y3ec` directory)
     - PV
        ```yaml
        cat <<EOF | kubectl apply -f -
        apiVersion: "v1"
        kind: PersistentVolume
        metadata:
          labels:
            arango_deployment: cluster
            role: dbserver
          name: dbserver-9xztmg4t-vka6ic19qcl1y3ec
        spec:
          accessModes:
          - ReadWriteOnce
          capacity:
            storage: 1Gi
          local:
            path: /mnt/data/vka6ic19qcl1y3ec
          persistentVolumeReclaimPolicy: Retain
          storageClassName: my-local-ssd
          volumeMode: Filesystem
          nodeAffinity:
            required:
              nodeSelectorTerms:
              - matchExpressions:
                - key: kubernetes.io/hostname
                  operator: In
                  values:
                  - minikube
        EOF
        ```
     - PVC
         ```yaml
         cat <<EOF | kubectl apply -f -
         apiVersion: v1
         kind: PersistentVolumeClaim
         metadata:
           labels:
             app: arangodb
             arango_deployment: cluster
             role: dbserver
           name: dbserver-9xztmg4t
         spec:
           accessModes:
           - ReadWriteOnce
           resources:
             requests:
               storage: 1Gi
           storageClassName: my-local-ssd
           volumeMode: Filesystem
           volumeName: dbserver-9xztmg4t-vka6ic19qcl1y3ec
         EOF
         ```

### Create ArangoDeployment with previously created PVC

Now we can create ArangoDeployment with previously created PVCs:
```yaml
cat <<EOF | kubectl apply -f -
apiVersion: "database.arangodb.com/v1"
kind: "ArangoDeployment"
metadata:
  name: "cluster"
spec:
  externalAccess:
    type: NodePort
  mode: Cluster
  agents:
    volumeClaimTemplate:
      spec:
        storageClassName: my-local-ssd
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        volumeMode: Filesystem
  dbservers:
    volumeClaimTemplate:
      spec:
        storageClassName: my-local-ssd
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        volumeMode: Filesystem
status:
  agency:
    ids:
      - AGNT-pntg5yc8
      - AGNT-kfyuj8ow
      - AGNT-bv5rofcz
    size: 3
  members:
    agents:
      - id: AGNT-pntg5yc8
        persistentVolumeClaim:
          name: agent-pntg5yc8
        persistentVolumeClaimName: agent-pntg5yc8
      - id: AGNT-kfyuj8ow
        persistentVolumeClaim:
          name: agent-kfyuj8ow
        persistentVolumeClaimName: agent-kfyuj8ow
      - id: AGNT-bv5rofcz
        persistentVolumeClaim:
          name: agent-bv5rofcz
        persistentVolumeClaimName: agent-bv5rofcz
    dbservers:
      - id: PRMR-9xztmg4t
        persistentVolumeClaim:
          name: cluster-dbserver-9xztmg4t
        persistentVolumeClaimName: cluster-dbserver-9xztmg4t
      - id: PRMR-l1pp19yl
        persistentVolumeClaim:
          name: cluster-dbserver-l1pp19yl
        persistentVolumeClaimName: cluster-dbserver-l1pp19yl
      - id: PRMR-31akmzrp
        persistentVolumeClaim:
          name: cluster-dbserver-31akmzrp
        persistentVolumeClaimName: cluster-dbserver-31akmzrp
EOF
```

That's it! Now you can use ArangoDB with restored data.
