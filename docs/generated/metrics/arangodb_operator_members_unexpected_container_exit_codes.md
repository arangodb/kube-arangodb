# arangodb_operator_members_unexpected_container_exit_codes (Counter)

## Description

Counter of unexpected restarts in pod (Containers/InitContainers/EphemeralContainers)

## Labels

|     Label      | Description                                |
|:--------------:|:-------------------------------------------|
|   namespace    | Deployment Namespace                       |
|      name      | Deployment Name                            |
|     member     | Member ID                                  |
|   container    | Container Name                             |
| container_type | Container/InitContainer/EphemeralContainer |
|      code      | ExitCode                                   |
|     reason     | Reason                                     |
