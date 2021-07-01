package cli_test

import (
	"strings"
	"testing"

	"github.com/ulikunitz/cli"
)

func TestFlagUsage(t *testing.T) {
	var f bool
	var str string

	flags := []*cli.Flag{
		cli.BoolFlag(&f, "", 'f', "a flag"),
		cli.StringFlag(&str, "str", 's', " a string flag"),
	}

	var sb strings.Builder
	cli.UsageFlags(&sb, flags, "  ")

	s := sb.String()
	t.Logf("usage: %s", s)
}
