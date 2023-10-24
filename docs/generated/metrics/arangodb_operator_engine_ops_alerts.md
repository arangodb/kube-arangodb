# arangodb_operator_engine_ops_alerts (Counter)

## Description

Counter for actions which requires ops attention

## Labels

|   Label   | Description          |
|:---------:|:---------------------|
| namespace | Deployment Namespace |
|   name    | Deployment Name      |


## Alerting

| Priority |                       Query                        | Description                                 |
|:--------:|:--------------------------------------------------:|:--------------------------------------------|
| Warning  | irate(arangodb_operator_engine_ops_alerts[1m]) &gt; 1 | Trigger an alert if OPS attention is needed |
