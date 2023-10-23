# arangodb_operator_engine_assertions (Counter)

## Description

Number of assertions invoked during Operator runtime

## Labels

| Label | Description   |
|:-----:|:--------------|
|  key  | Assertion Key |


## Alerting

| Priority |                       Query                        | Description                                 |
|:--------:|:--------------------------------------------------:|:--------------------------------------------|
| Warning  | irate(arangodb_operator_engine_assertions[1m]) &gt; 1 | Trigger an alert if OPS attention is needed |
