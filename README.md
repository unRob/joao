# `joao`

A very wip configuration manager. Keeps config entries encoded as YAML in the filesystem, backs it up to 1Password, and syncs scrubbed copies to git. robots consume entries via 1Password Connect + Vault.

## Usage

```sh
# PATH refers to a filesystem path
# examples: config/host/juazeiro.yaml, service/gitea/config.joao.yaml

# QUERY refers to a sequence of keys delimited by dots
# examples: tls.cert, roles.0, dc, . (literal dot meaning the whole thing)

# there's better help available within each command, try:
joao get --help

# get a single value/tree from a single item/file
joao get [--output|-o=(raw|json|yaml|op)] [--remote] PATH [QUERY]
# set/update a single value in a single item/file
joao set [--secret] [--flush] [--input=/path/to/input|<<<"value"] PATH QUERY
# sync local changes upstream
joao flush [--dry-run] [--redact] PATH
# sync remote secrets to filesystem
joao fetch [--dry-run] PATH
# check for differences between local and remote items
joao diff [--cache] PATH

# show information on the git integration
joao git-filter

# show information on the vault integration
joao vault-plugin --help
```

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

## git integration

In order to store configuration files within a git repository while keeping secrets off remote copies, `joao` provides git filters.

To install them, **every collaborator** would need to run:

```sh
# setup filters in your local copy of the repo:
# this runs when you check in a file (i.e. about to commit a config file)
# it will flush secrets to 1password before removing secrets from the file on disk
git config filter.joao.clean "joao git-filter clean --flush %f"
# this step runs after checkout (i.e. pulling changes)
# it simply outputs the file as-is on disk
git config filter.joao.smudge cat
# let's enforce these filters
git config filter.joao.required true

# optionally, configure a diff filter to show changes as would be commited to git
# this does not modify the original file on disk
git config diff.joao.textconv "joao git-filter diff"
```

Then, **only once**, we need to specify which files to apply the filters and diff commands to:

```sh
# adds diff and filter attributes for config files ending with .joao.yaml
echo '**/*.joao.yaml filter=joao diff=joao' >> .gitattributes
# finally, commit and push these attributes
git add .gitattributes
git commit -m "installing joao attributes"
git push origin main
```

See:
  - https://git-scm.com/docs/gitattributes#_filter
  - https://git-scm.com/docs/gitattributes#_diff

## vault integration

`joao` can run as a plugin to Hashicorp Vault, and make whole configuration entries available—secrets and all—through the Vault API.

To install, download `joao` to the machine running `vault` at the `plugin_directory`, as specified by vault's config. The installed `joao` executable needs to be executable for the user running vault only.

### Configuration
```sh
export VAULT_PLUGIN_DIR=/var/lib/vault/plugins
chmod 700 "$VAULT_PLUGIN_DIR/joao"
export PLUGIN_SHA="$(openssl dgst -sha256 -hex "$VAULT_PLUGIN_DIR/joao" | awk '{print $2}')"
export VERSION="$($VAULT_PLUGIN_DIR/joao --version)"

# register
vault plugin register -sha256="$PLUGIN_SHA" -command=joao -args="vault-plugin" -version="$VERSION" secret joao

# configure, add `vault` to set a default vault for querying
vault write config/1password "host=$OP_CONNECT_HOST" "token=$OP_CONNECT_TOKEN" # vault=my-default-vault

if !vault plugin list secret | grep -c -m1 '^joao ' >/dev/null; then
  # first time, let's enable the secrets backend
  vault secrets enable --path=config joao
else
  # updating from a previous version
  vault secrets tune -plugin-version="$VERSION" config/
  vault plugin reload -plugin joao
fi
```

### Vault API

```sh
# VAULT is optional if configured with a default `vault`. See above

# vault read config/tree/[VAULT/]ITEM
vault read config/tree/service:api
vault read config/tree/prod/service:api

# vault list config/trees/[VAULT/]
vault list config/trees
vault list config/trees/prod
```

See:
  - https://developer.hashicorp.com/vault/docs/plugins
