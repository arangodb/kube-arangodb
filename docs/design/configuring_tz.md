# Configuring timezone

To set timezone for cluster components, mount the required timezone into container
by adjusting `spec.<group>` of ArangoDeployment resource:
```yaml
dbservers:
  volumeMounts:
    - mountPath: /etc/localtime
      name: timezone
  volumes:
    - hostPath:
        path: /usr/share/zoneinfo/Europe/Warsaw
        type: File
      name: timezone
```

If `/usr/share/zoneinfo` is not present on your host your probably have to install `tzdata` package. 

