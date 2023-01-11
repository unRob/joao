// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"git.rob.mx/nidito/chinampa/pkg/command"
	"git.rob.mx/nidito/joao/internal/vault"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/plugin"
)

var Plugin = &command.Command{
	Path:    []string{"vault", "server"},
	Summary: "Starts a vault-joao-plugin server",
	Description: `Runs ﹅joao﹅ as a vault plugin.

You'll need to install ﹅joao﹅ in the machine running ﹅vault﹅ to ﹅plugin_directory﹅ as specified by vault's config. The installed ﹅joao﹅ executable needs to be executable for the user running vault only.

### Configuration
﹅﹅﹅sh
export VAULT_PLUGIN_DIR=/var/lib/vault/plugins
chmod 700 "$VAULT_PLUGIN_DIR/joao"
export PLUGIN_SHA="$(openssl dgst -sha256 -hex "$VAULT_PLUGIN_DIR/joao" | awk '{print $2}')"
export VERSION="$($VAULT_PLUGIN_DIR/joao --version)"

# register
vault plugin register -sha256="$PLUGIN_SHA" -command=joao -args="vault,server" -version="$VERSION" secret joao

# configure, add ﹅vault﹅ to set a default vault for querying
vault write config/1password "host=$OP_CONNECT_HOST" "token=$OP_CONNECT_TOKEN" # vault=my-default-vault

if !vault plugin list secret | grep -c -m1 '^joao ' >/dev/null; then
  # first time, let's enable the secrets backend
  vault secrets enable --path=config joao
else
  # updating from a previous version
  vault secrets tune -plugin-version="$VERSION" config/
  vault plugin reload -plugin joao
fi
﹅﹅﹅

### Vault API

﹅﹅﹅sh
# VAULT is optional if configured with a default ﹅vault﹅. See above

# vault read config/tree/[VAULT/]ITEM
vault read config/tree/service:api
vault read config/tree/prod/service:api

# vault list config/trees/[VAULT/]
vault list config/trees
vault list config/trees/prod
﹅﹅﹅

See:
  - https://developer.hashicorp.com/vault/docs/plugins
`,
	Options: command.Options{
		"ca-cert": {
			Type:        command.ValueTypeString,
			Description: "See https://pkg.go.dev/github.com/hashicorp/vault/api#TLSConfig",
		},
		"ca-path": {
			Type:        command.ValueTypeString,
			Description: "See https://pkg.go.dev/github.com/hashicorp/vault/api#TLSConfig",
		},
		"client-cert": {
			Type:        command.ValueTypeString,
			Description: "See https://pkg.go.dev/github.com/hashicorp/vault/api#TLSConfig",
		},
		"client-key": {
			Type:        command.ValueTypeString,
			Description: "See https://pkg.go.dev/github.com/hashicorp/vault/api#TLSConfig",
		},
		"tls-skip-verify": {
			Type:        command.ValueTypeBoolean,
			Description: "See https://pkg.go.dev/github.com/hashicorp/vault/api#TLSConfig",
			Default:     false,
		},
	},
	Action: func(cmd *command.Command) error {
		return plugin.ServeMultiplex(&plugin.ServeOpts{
			BackendFactoryFunc: vault.Factory,
			TLSProviderFunc: api.VaultPluginTLSProvider(&api.TLSConfig{
				CACert:        cmd.Options["ca-cert"].ToString(),
				CAPath:        cmd.Options["ca-path"].ToString(),
				ClientCert:    cmd.Options["client-cert"].ToString(),
				ClientKey:     cmd.Options["client-key"].ToString(),
				TLSServerName: "",
				Insecure:      cmd.Options["tls-skip-verify"].ToValue().(bool),
			}),
		})
	},
}
