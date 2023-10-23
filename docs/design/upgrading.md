# Upgrade procedure

## Upgrading ArangoDB single to another version

The process for upgrading an existing ArangoDB single server
to another version is as follows:

- Set CR state to `Upgrading`
- Remove the server Pod (keep persistent volume)
- Create a new server Pod with new version
- Wait until server is ready before continuing
- Set CR state to `Ready`

## Upgrading ArangoDB cluster to another version

The process for upgrading an existing ArangoDB cluster
to another version is as follows:

- Set CR state to `Upgrading`
- For each agent:
  - Remove the agent Pod (keep persistent volume)
  - Create new agent Pod with new version
  - Wait until agent is ready before continuing
- For each dbserver:
  - Remove the dbserver Pod (keep persistent volume)
  - Create new dbserver Pod with new version
  - Wait until dbserver is ready before continuing
- For each coordinator:
  - Remove the coordinator Pod (keep persistent volume)
  - Create new coordinator Pod with new version
  - Wait until coordinator is ready before continuing
- Set CR state to `Ready`
