# Maintenance

## Development on MacOS

This repo requires GNU command line tools instead BSD one (which are by default available on Mac).

Please add following to your `~/bashrc` or `~/.zshrc` file (it requires Hombebrew to be installed):

```shell
HOMEBREW_PREFIX=$(brew --prefix)
for d in ${HOMEBREW_PREFIX}/opt/*/libexec/gnubin; do export PATH=$d:$PATH; done
```

## ArangoDeployment

Maintenance on ArangoDeployment can be enabled using annotation.

Key: `deployment.arangodb.com/maintenance`
Value: `true`

To enable maintenance mode for ArangoDeployment kubectl command can be used:
`kubectl annotate arangodeployment deployment deployment.arangodb.com/maintenance=true`

To disable maintenance mode for ArangoDeployment kubectl command can be used:
`kubectl annotate --overwrite arangodeployment deployment deployment.arangodb.com/maintenance-`