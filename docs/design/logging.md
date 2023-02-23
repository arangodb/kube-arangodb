# Logging configuration

## Operator logging

### Log level

To adjust logging level of the operator, you can use `operator.args` in chart template value 
as described in [Additional configuration](./additional_configuration.md).

For example, to set log level to `INFO` and `DEBUG` for `requests` package, you can use the following value:
```yaml
operator:
  args: ["--log.level=INFO", "--log.level=requests=DEBUG"]
```

### Log format

By default, operator logs in `pretty` format.

To switch logging format to the JSON, you can use `operator.args` in chart template value:
```yaml
operator:
  args: ["--log.format=pretty"]
```

## ArangoDeployment logging

By default, ArangoDeployment logs in `pretty` format.

To switch logging format to the JSON we need to pass `--log.use-json-format` argument to the ArangoDB server in the deployment:
```yaml
apiVersion: database.arangodb.com/v1
kind: ArangoDeployment
metadata:
  name: single
spec:
  mode: Single
  single:
    args:
      - --log.use-json-format
      - --log.level=INFO
      - --log.level=backup=TRACE
```
