# Amazon AWS Remarks

## Elastic Block Storage

Documentation:
- [AWS EBS Volume Types](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-volume-types.html)

Remarks:
- It is recommended to use at least GP2 (can be IO1) volume type for ArangoDeployment PV.
- GP2 Volume IOPS is mostly based on storage size. If bigger load is expected use bigger volumes.
- GP2 Volume supports burst mode. In case load in ArangoDeployment is expected only periodically you can use
smaller GP2 Volumes to save costs.
- AWS EBS support resizing of Volume. Volume size can be changed during lifetime, but it requires pod to be recreated.

## LoadBalancer

Documentation:
- [AWS LB Annotations](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/#load-balancers)

Remarks:
- AWS LB in TCP mode is able to resend request in case of timeout while waiting for response from Coordinator/DBServer.
This can break some POST requests, like data insertion. To change default value, set to 60s,
you can set annotation for ArangoDeployment LoadBalancer service.
```
kubectl annotate --overwrite service/<ArangoDeployment name>-ea service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout=<value is seconds, max 15 min>
```