# Scaling

The ArangoDB operator supports up and down scaling of
the number of dbservers & coordinators.

Q: Agents as well?

The scale up or down, change the number of servers in the custom
resource and apply the updated resource using:

```bash
kubectl apply -f yourCustomResourceFile.yaml
```

## Internal process

The internal process followed by the ArangoDB operator
when scaling up is as follows:

- Set CR state to `Scaling`
- Create an additional server Pod
- Wait until server is ready before continuing
- Set CR state to `Ready`

The internal process followed by the ArangoDB operator
when scaling down a dbserver is as follows:

- Set CR state to `Scaling`
- Drain the dbserver (TODO fill in procedure)
- Shutdown the dbserver such that it removes itself from the agency
- Remove the dbserver Pod
- Set CR state to `Ready`

The internal process followed by the ArangoDB operator
when scaling down a coordinator is as follows:

- Set CR state to `Scaling`
- Shutdown the coordinator such that it removes itself from the agency
- Remove the coordinator Pod
- Set CR state to `Ready`

Note: Scaling is always done 1 server at a time.
