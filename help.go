package cli

import (
	"os"
)

// AddHelpCommand adds a subcommand help to the root command if doesn't support
// a help command already.
func AddHelpCommand(root *Command) bool {
	for _, cmd := range root.Subcommands {
		if cmd.Name == "help" {
			return false
		}
	}

	f := func(args []string) error {
		commands, _, err := Parse(root, args)
		if err != nil {
			return err
		}
		cmd := commands[len(commands)-1]
		_, err = cmd.WriteDoc(os.Stdout)
		return err
	}

	cmd := &Command{
		Name:  "help",
		Info:  "prints help messages",
		Usage: root.Name + " help <commands>...",
		Exec:  f,
	}

	root.Subcommands = append(root.Subcommands, cmd)

	return true
}

var helpFlag = false

func helpOption() *Option {
	return &Option{
		Name:        "help",
		Short:       'h',
		Description: "prints help message for command",
		HasParam:    false,
		Default:     "",
		SetValue: func(arg string, noParam bool) error {
			helpFlag = true
			return nil
		},
		ResetValue: func() { helpFlag = false },
	}
}

// AddHelpOption adds a help option for the command if it doesn't have an option
// -h already. Note the Exec function must already been set.
func AddHelpOption(cmd *Command) bool {
	if cmd.Name == "help" {
		return false
	}
	if cmd.Exec == nil {
		return false
	}
	for _, o := range cmd.Options {
		if o.Short == 'h' {
			return false
		}
	}
	f := cmd.Exec
	newF := func(args []string) error {
		if helpFlag {
			_, err := cmd.WriteDoc(os.Stdout)
			return err
		}
		return f(args)
	}
	cmd.Options = append(cmd.Options, helpOption())
	cmd.Exec = newF
	return true
}

// AddHelpOptionToAll adds a help option to all subcommands that don't have the
// name help.
func AddHelpOptionToAll(cmd *Command) {
	AddHelpOption(cmd)
	for _, c := range cmd.Subcommands {
		AddHelpOptionToAll(c)
	}
}
