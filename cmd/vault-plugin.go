// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"os"

	"git.rob.mx/nidito/chinampa/pkg/command"
	"git.rob.mx/nidito/joao/internal/vault"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/plugin"
)

var Plugin = &command.Command{
	Path:    []string{"vault-plugin"},
	Summary: "Starts a vault-joao-plugin server",
	Description: `﹅joao﹅ can run as a plugin to Hashicorp Vault, and make whole configuration entries available—secrets and all—through the Vault API.

To install, download ﹅joao﹅ to the machine running ﹅vault﹅ at the ﹅plugin_directory﹅, as specified by vault's config. The installed ﹅joao﹅ executable needs to be executable for the user running vault only.

### Configuration
﹅﹅﹅sh
export VAULT_PLUGIN_DIR=/var/lib/vault/plugins
chmod 700 "$VAULT_PLUGIN_DIR/joao"
export PLUGIN_SHA="$(openssl dgst -sha256 -hex "$VAULT_PLUGIN_DIR/joao" | awk '{print $2}')"
export VERSION="$($VAULT_PLUGIN_DIR/joao --version)"

# register
vault plugin register -sha256="$PLUGIN_SHA" -command=joao -args="vault-plugin" -version="$VERSION" secret joao

# configure, add ﹅vault﹅ to set a default vault for querying
vault write config/1password "host=$OP_CONNECT_HOST" "token=$OP_CONNECT_TOKEN" # vault=my-default-vault

if !(vault plugin list secret | grep -c -m1 '^joao ' >/dev/null); then
  # first time, let's enable the secrets backend
  vault secrets enable -path=config joao
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
		"sigh0": {
			ShortName: "c",
			Default:   "",
		},
		"sigh1": {
			ShortName: "t",
			Default:   "",
		},
	},
	Action: func(cmd *command.Command) error {
		apiClientMeta := &api.PluginAPIClientMeta{}
		flags := apiClientMeta.FlagSet()
		flags.Parse(os.Args[2:])

		tlsConfig := apiClientMeta.GetTLSConfig()
		tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)
		return plugin.ServeMultiplex(&plugin.ServeOpts{
			BackendFactoryFunc: vault.Factory,
			TLSProviderFunc:    tlsProviderFunc,
		})
	},
}
