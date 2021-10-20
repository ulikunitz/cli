package cli

import (
	"os"
)

// HelpCommand generates a help command that prints the documentation for the
// commands of the program.
func HelpCommand(root *Command) *Command {

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

	return cmd
}
