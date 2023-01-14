// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"os"

	"git.rob.mx/nidito/chinampa/pkg/command"
	"git.rob.mx/nidito/joao/pkg/config"
)

var GitFilters = []*command.Command{
	FilterDiff,
	FilterClean,
	FilterGroup,
}

func redactedData(cmd *command.Command) error {
	path := cmd.Arguments[0].ToValue().(string)

	flush := false
	if opt, ok := cmd.Options["flush"]; ok {
		flush = opt.ToValue().(bool)
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	cfg, err := config.FromYAML(contents)
	if err != nil {
		return err
	}

	if flush {
		name, vault, err := config.VaultAndNameFrom(path, contents)
		if err != nil {
			return err
		}

		cfg.Name = name
		cfg.Vault = vault
	}

	res, err := cfg.AsYAML(config.OutputModeRedacted)
	if err != nil {
		return err
	}

	_, err = cmd.Cobra.OutOrStdout().Write(res)
	return err
}

var FilterGroup = &command.Command{
	Path:    []string{"git-filter"},
	Summary: "Subcommands used by `git` as filters",
	Description: `In order to store configuration files within a git repository while keeping secrets off remote copies, ﹅joao﹅ provides git filters.

To install them, **every collaborator** would need to run:

﹅﹅﹅sh
# setup filters in your local copy of the repo:
# this runs when you check in a file (i.e. about to commit a config file)
# it will flush secrets to 1password before removing secrets from the file on disk
git config filter.joao.clean "joao git-filter clean --flush %f"
# this step runs after checkout (i.e. pulling changes)
# it simply outputs the file as-is on disk
git config filter.joao.smudge cat
# let's enforce these filters
git config filter.joao.required true

# optionally, configure a diff filter to show changes as would be committed to git
# this does not modify the original file on disk
git config diff.joao.textconv "joao git-filter diff"
﹅﹅﹅

Then, **only once**, we need to specify which files to apply the filters and diff commands to:

﹅﹅﹅sh
# adds diff and filter attributes for config files ending with .joao.yaml
echo '**/*.joao.yaml filter=joao diff=joao' >> .gitattributes
# finally, commit and push these attributes
git add .gitattributes
git commit -m "installing joao attributes"
git push origin main
﹅﹅﹅

See:
  - https://git-scm.com/docs/gitattributes#_filter
  - https://git-scm.com/docs/gitattributes#_diff`,
	Arguments: command.Arguments{},
	Options:   command.Options{},
	Action: func(cmd *command.Command) error {
		data, err := cmd.ShowHelp(command.Root.Options, os.Args)
		if err != nil {
			return err
		}
		_, err = cmd.Cobra.OutOrStderr().Write(data)
		return err
	},
}

var FilterDiff = &command.Command{
	Path:        []string{"git-filter", "diff"},
	Summary:     "a filter for git to call during `git diff`",
	Description: `see ﹅joao git-filter﹅ for instructions to install this filter`,
	Arguments: command.Arguments{
		{
			Name:        "path",
			Description: "The git staged path to read from",
			Required:    true,
			Values: &command.ValueSource{
				Files: &[]string{"yaml", "yml"},
			},
		},
	},
	Options: command.Options{},
	Action:  redactedData,
}

var FilterClean = &command.Command{
	Path:    []string{"git-filter", "clean"},
	Summary: "a filter for git to call when a file is checked in",
	Description: `see ﹅joao git-filter﹅ for instructions to install this filter

Use ﹅--flush﹅ to save changes to 1password before redacting file.`,
	Arguments: command.Arguments{
		{
			Name:        "path",
			Description: "The git staged path to read from",
			Required:    true,
			Values: &command.ValueSource{
				Files: &[]string{"yaml", "yml"},
			},
		},
	},
	Options: command.Options{
		"flush": {
			Description: "Save to 1Password after before redacting",
			Type:        "bool",
		},
	},
	Action: redactedData,
}
