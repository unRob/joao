# `joao`

A very wip configuration manager. Keeps config entries encoded as YAML in the filesystem, backs it up to 1Password, and syncs scrubbed copies to git. robots consume entries via 1Password Connect + Vault.

## Why

So I wanted to operate on my configuration mess...

- With a workflow something like [SOPS](https://github.com/mozilla/sops)',
- but that talks UNIX, like [go-config-yourself](https://github.com/unRob/go-config-yourself) (plus its later `bash` + `jq` + `yq` [re-implementation](https://github.com/unRob/nidito/tree/0812e0caf6d81dd06b740701c3e95a2aeabd86de/.milpa/commands/nidito/config)'s multi-storage improvements),
- That emulates [git-crypt](https://github.com/AGWA/git-crypt)'s sweet git filters,
- and plays nice with [1Password's neat ecosystem](https://developer.1password.com/),
- as well as Hashicorp's [Vault](https://vaultproject.io/),
- but is still just files, folders and git for all I care.

And thus, I set to write me, yet again, some configuration toolchain that:

- Allows the _structure_ of config trees to live happily **in the filesystem**: my home+cloud DC uses a lot of configuration spread over multiple files, one-off services don't really need the whole folder structure—I want a single tool to handle both.
- Prevents _secrets_ from ending up in **remote repositories**: I really dig `git-crypt`'s filters, not quite sure about how to safely operate them yet...
- Makes it **easy to edit** entries locally, as well as on the go: Easy for me to R/W, so YAML files, and 1Password's tools are pretty great for quick edits remotely.
- Is capable of bootstrapping other secret mangement processes: A single binary can talk to `op`'s CLI (hello, touch ID on macos!), to a 1password-connect server, and to vault as a plugin.

For a deeper dive on these points above, check out my [docs/letter-to-secret-santa.md](docs/letter-to-secret-santa.md).

---

## Configuration

Schema for configuration and non-secret values live along the code, and are pushed to remote origins. Secrets can optionally and temporally be flushed to disk for editing or other sorts of operations. Git filters are available to prevent secrets from being pushed to remotes. Secrets are grouped into files, and every file gets its own 1Password item.

Secret values are specified using the `!!secret` YAML tag.

The ideal workflow is:

1. configs are written to disk, temporarily
2. `joao flush --redact`es them to 1password, and removes secrets from disk
3. configuration values, secret or not, are read from:
  - `joao get` as needed by local processes. Mostly thinking of the human in the loop here, where `op` and suitable auth (i.e. touchid) workflows are available.
  - from 1Password Connect, for when vault is not configured or available (think during provisioning)
  - from Hashicorp Vault, for any automated process, after provisioning is complete.

---

`joao` operates on two modes, **repo** and **single-file**.

- **Repo** mode is useful to have multiple configuration files in a folder structure while configuring their 1Password mappings (vault and item names) in a single file.
- **Single-file** mode is useful when a single file contains all of the desired configuration, and its 1Password mapping is defined in that same file.

### Repo mode

Basically, configs are kept in a directory and their relative path maps to their 1Password item name. A `.joao.yaml` file must exist at the root configuration directory, specifying the 1Password vault to use, and optionally a prefix to prepend ot every item name

```yaml
# config/.joao.yaml
# the 1password vault to use as storage
vault: infra
# the optional nameTemplate is a go-template specifying the desired items' names
# turns config/host/juazeiro.yaml to host:juazeiro
nameTemplate: '{{ DirName }}:{{ FileName}}'
```

```yaml
# config/host/juazeiro.yaml => infra/host:juazeiro
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

### Single file mode

In single file mode, `joao` expects every file to have a `_joao: !!config` key with a vault name, and a name for the 1Password item.

```yaml
# src/git/config.yaml
_config: !!joao
  vault: bahianos
  name: service:git
smtp:
  server: smtp.example.org
  username: git@example.org
  password: !!secret quatro-paredes
  port: 587

```

## Usage

```sh
# NAME can be either a filesystem path or a colon delimited item name
# for example: config/host/juazeiro.yaml or [op-vault-name/]host:juazeiro

# DOT_DELIMITED_PATH is
# for example: tls.cert, roles.0, dc

# get a single value/tree from a single item/file
joao get NAME [--output|-o=(raw|json|yaml|op)] [--remote] [jq expr]
# set/update a single value in a single item/file
joao set NAME DOT_DELIMITED_PATH [--secret] [--flush] [--input=/path/to/input|<<<"value"]
# sync local changes upstream
joao flush NAME [--dry-run] [--redact]
# sync remote secrets to filesystem
joao fetch NAME [--dry-run]
# check for differences between local and remote items
joao diff PATH [--cache]
# initialize a new joao repo
joao repo init [PATH]
# list the item names within prefix
joao repo list [PREFIX]
# print the repo config root
joao repo root
#
joao repo status
joao repo filter clean FILE
joao repo filter diff PATH OLD_FILE OLD_SHA OLD_MODE NEW_FILE NEW_SHA NEW_MODE
joao repo filter smudge FILE
```
