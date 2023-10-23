# Scaling

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
