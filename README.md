# joao

a very wip configuration manager. keep config in the filesystem, back it up to 1password. Make it available to services via vault + 1password connect.

```yaml
# config/.joao.yaml
# the 1password vault to use as storage
vault: nidito
# the prefix to prepend to all configs from this directory
prefix: example
# think about single config or config in files 
```

```yaml
# config/host/juazeiro.yaml
address: 142.42.42.42
dc: bah0
mac: !!secret 00:11:22:33:44:55
tls:
  cert: !!secret |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
roles:
  - consul-client
  - nomad-client
  - http
token:
  bootstrap: !!secret 01234567-89ab-cdfe-0123-456789abcdef
```

```sh
# NAME can be either a filesystem path or a colon delimited item name
# for example: config/host/juazeiro.yaml or host:juazeiro

# DOT_DELIMITED_PATH is 
# for example: tls.cert, roles.0, dc

joao get NAME [--output|-o=(raw|json|yaml)] [--remote|--local] [jq expr]
joao set NAME [--from=/path/to/input] [--secret] [--flush] DOT_DELIMITED_PATH [<<<"value"]
joao flush NAME [--dry-run]
joao fetch NAME [--dry-run]
# PREFIX is a prefix to search for keys at, i.e. NAME[.DOT_DELIMITED_PATH]
joao list PREFIX
joao diff NAME [--cache]
joao repo init
joao repo status
joao repo filter clean FILE
joao repo filter diff PATH OLD_FILE OLD_SHA OLD_MODE NEW_FILE NEW_SHA NEW_MODE
joao repo filter smudge FILE
```
