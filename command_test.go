package cli_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ulikunitz/cli"
)

func TestParse(t *testing.T) {
	var (
		dir string
		/*
			forceDelete bool
			forceCreate bool
			listAll     bool
		*/
	)

	var sb strings.Builder

	versionCmd := &cli.Command{
		Name:  "version",
		Info:  "prints version information",
		Usage: "foo version",
		Description: "The command version prints the version of" +
			" program foo",
		Exec: func(args []string) error {
			fmt.Fprintf(&sb, "command version\n")
			return nil
		},
	}

	root := &cli.Command{
		Name:        "foo",
		Info:        "test program",
		Usage:       "foo [Options] <subcommand>",
		Description: "The program foo provides several subcommands.",
		Options: []*cli.Option{
			cli.StringOption(&dir, "dir", 'd', "directory"),
		},
		Subcommands: []*cli.Command{versionCmd},
		Exec: func(args []string) error {
			fmt.Fprintf(&sb, "command root\n")
			fmt.Fprintf(&sb, "dir %q\n", dir)
			return nil
		},
	}

	tests := []struct {
		args   []string
		output string
	}{
		{args: []string{"version"}, output: "command version\n"},
		{args: []string{"--dir=foo"},
			output: "command root\ndir \"foo\"\n"},
	}

	for _, tc := range tests {
		sb.Reset()
		err := cli.Run(root, tc.args)
		if err != nil {
			t.Fatalf("Run %s error %s", tc.args, err)
		}
	}

	sb.Reset()
	_, err := root.WriteDoc(&sb)
	if err != nil {
		t.Fatalf("root.WriteDoc error %s", err)
	}
	doc := sb.String()
	t.Logf("doc:\n%s", doc)
}
