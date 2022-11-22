package command

import (
	_c "git.rob.mx/nidito/joao/internal/constants"
	"git.rob.mx/nidito/joao/internal/runtime"
)

var Root = &Command{
	Summary:     "Helps organize config for roberto",
	Description: `﹅joao﹅ makes yaml, json, 1password and vault play along nicely.`,
	Path:        []string{"joao"},
	Options: Options{
		_c.HelpCommandName: &Option{
			ShortName:   "h",
			Type:        "bool",
			Description: "Display help for any command",
		},
		"verbose": &Option{
			ShortName:   "v",
			Type:        "bool",
			Default:     runtime.VerboseEnabled(),
			Description: "Log verbose output to stderr",
		},
		"no-color": &Option{
			Type:        "bool",
			Description: "Disable printing of colors to stderr",
			Default:     !runtime.ColorEnabled(),
		},
		"color": &Option{
			Type:        "bool",
			Description: "Always print colors to stderr",
			Default:     runtime.ColorEnabled(),
		},
		"silent": &Option{
			Type:        "bool",
			Description: "Silence non-error logging",
		},
		"skip-validation": &Option{
			Type:        "bool",
			Description: "Do not validate any arguments or options",
		},
	},
}
