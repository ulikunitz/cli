/*
Package cli supports the creation of command line appplications with
subcommands and help output.

A typical program will import this package, setup the root command and add a
help command.

  package main

  import (
	"log"
	"os"

	"github.com/ulikunitz/cli"
  )

  func main() {
	log.SetFlags(0)

	root := &cli.Command{
		Name:        "foo",
		Info:        "program to run compression benchmarks",
		Subcommands: []*cli.Command{subcommand()},
	}

	cli.AddHelpCommand(root)

	var args []string
	if len(os.Args) == 1 {
		args = []string{"help"}
	} else {
		args = os.Args[1:]
	}

	if err := cli.Run(root, args); err != nil {
		log.Fatal(err)
	}
  }
*/
package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"
)

// Command represents a command in the command tree. It may be the root of its
// own subcommand tree. The program itself will be represented by a root command
// with the name of the program.
type Command struct {
	// Name of command usually short (e.g. "list")
	Name string
	// Short description of the command (e.g. "list all config parameters")
	Info string
	// The usage string may have multiple lines.
	Usage string
	// Longer description that will be formatted.
	Description string
	// Options list. Note these options must immediately follow the
	// command in the command line and any non-option will stop the
	// processing of the options for this command.
	Options []*Option
	// List of all subcommands for this command.
	Subcommands []*Command
	// Function that executes the command.
	Exec func(args []string) error
}

func findCommand(commands []*Command, name string) (cmd *Command, ok bool) {
	for _, cmd := range commands {
		if cmd.Name == name {
			return cmd, true
		}
	}
	return nil, false
}

func maxLen(strings []string) int {
	n := 0
	for _, s := range strings {
		k := utf8.RuneCountInString(s)
		if k > n {
			n = k
		}
	}
	return n
}

// WriteDoc puts the documentation our on w. the style used is that of man
// files.
func (cmd *Command) WriteDoc(w io.Writer) (n int, err error) {
	const indent = "    "
	var k int
	i := 0
	if cmd.Name != "" || cmd.Info != "" {
		k, err = fmt.Fprintln(w, "NAME")
		n += k
		if err != nil {
			return n, err
		}
		i++
		switch {
		case cmd.Name != "" && cmd.Info != "":
			k, err = fmt.Fprintf(w, "%s%s - %s\n",
				indent, cmd.Name, cmd.Info)
		case cmd.Name != "":
			k, err = fmt.Fprintln(w, indent, cmd.Name)
		case cmd.Info != "":
			k, err = fmt.Fprintln(w, indent, cmd.Info)
		}
		n += k
		if err != nil {
			return n, err
		}
	}
	if cmd.Usage != "" {
		if i > 0 {
			k, err = fmt.Fprintln(w)
			n += k
			if err != nil {
				return n, err
			}
		}
		i++
		k, err = fmt.Fprintln(w, "USAGE")
		n += k
		if err != nil {
			return n, err
		}
		k, err = fmt.Fprintln(w, indent, cmd.Usage)
		n += k
		if err != nil {
			return n, err
		}
	}
	if cmd.Description != "" {
		if i > 0 {
			k, err = fmt.Fprintln(w)
			n += k
			if err != nil {
				return n, err
			}
		}
		i++
		k, err = fmt.Fprintln(w, "DESCRIPTION")
		n += k
		if err != nil {
			return n, err
		}
		k, err = formatText(w, cmd.Description, 80, indent)
		n += k
		if err != nil {
			return n, err
		}
	}
	if len(cmd.Options) > 0 {
		if i > 0 {
			k, err = fmt.Fprintln(w)
			n += k
			if err != nil {
				return n, err
			}
		}
		i++
		k, err = fmt.Fprintln(w, "OPTIONS")
		n += k
		if err != nil {
			return n, err
		}
		k, err = UsageOptions(w, cmd.Options, indent, indent)
		n += k
		if err != nil {
			return n, err
		}
	}
	if len(cmd.Subcommands) > 0 {
		if i > 0 {
			k, err = fmt.Fprintln(w)
			n += k
			if err != nil {
				return n, err
			}
		}
		i++
		k, err = fmt.Fprintln(w, "SUBCOMMANDS")
		n += k
		if err != nil {
			return n, err
		}
		names := make([]string, 0, len(cmd.Subcommands))
		for _, c := range cmd.Subcommands {
			if c.Name != "" {
				names = append(names, c.Name)
			}
		}
		sort.Strings(names)

		maxNameLen := maxLen(names)

		for _, name := range names {
			subcmd, ok := findCommand(cmd.Subcommands, name)
			if !ok {
				panic(fmt.Errorf("can't find %q", name))
			}
			if subcmd.Info != "" {
				k, err = fmt.Fprintf(w, "%s%-*s- %s\n",
					indent, maxNameLen+1, name, subcmd.Info)
			} else {
				k, err = fmt.Fprintf(w, "%s%s\n", indent, name)
			}
			n += k
			if err != nil {
				return n, err
			}
		}
	}
	if i > 0 {
		k, err = fmt.Fprintln(w)
		n += k
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

// CommandError might be generated during Command parsing.
type CommandError struct {
	Name    string
	Message string
	Wrapped error
}

// Unwrap returns the wrapped error.
func (err *CommandError) Unwrap() error { return err.Wrapped }

// Error prints the error message and appends the error string of the wrapped
// error.
func (err *CommandError) Error() string {
	var sb strings.Builder

	if err.Name != "" {
		fmt.Fprintf(&sb, "%s", err.Name)
		if err.Message != "" {
			fmt.Fprintf(&sb, ": %s", err.Message)
		}
	} else {
		fmt.Fprintf(&sb, "%s", err.Message)
	}

	if err.Wrapped != nil {
		fmt.Fprintf(&sb, ": %s", err.Wrapped)
	}

	return sb.String()
}

func unrecognizedCommand(arg string) *CommandError {
	return &CommandError{
		Name:    "unrecognized",
		Message: fmt.Sprintf("unrecognized command %s", arg),
	}
}

// Parse parses the argument list and determines the sequence of subcommands.
// The root command itself is not parsed but its flags. Out is used for error
// messages during parsing. The return value n provides the number of commands
// parsed.
func Parse(root *Command, args []string) (commands []*Command, n int, err error) {
	commands = make([]*Command, 0, 4)
	cmd := root
	for {
		commands = append(commands, cmd)
		if len(cmd.Options) > 0 {
			k, err := ParseOptions(cmd.Options, args[n:])
			n += k
			if err != nil {
				if cmd != root {
					err = &CommandError{
						Name:    cmd.Name,
						Message: "",
						Wrapped: err}
				}
				return commands, n, err
			}
		}
		if n < len(args) {
			arg := args[n]
			var found *Command
			for _, c := range cmd.Subcommands {
				if strings.HasPrefix(c.Name, arg) {
					if found != nil {
						err = unrecognizedCommand(arg)
						return commands, n, err
					}
					found = c
				}
			}
			if found == nil {
				return commands, n, nil
			}
			n++
			cmd = found
			continue
		}
		return commands, n, nil
	}
}

// Run parses the arguments and executes the exec command for the command
// identified. The call may return an error.
func Run(root *Command, args []string) error {
	commands, n, err := Parse(root, args)
	if err != nil {
		return err
	}
	cmd := commands[len(commands)-1]
	if cmd.Exec == nil {
		err := &CommandError{
			Name:    cmd.Name,
			Message: "couldn't find executable subcommand",
		}
		return err
	}
	args = args[n:]
	err = cmd.Exec(args)
	return err
}
