package cli_test

import (
	"strings"
	"testing"

	"github.com/ulikunitz/cli"
)

func TestUsageOptions(t *testing.T) {
	var f bool
	var str string

	opts := []*cli.Option{
		cli.BoolOption(&f, "", 'f', "a boolean option"),
		cli.StringOption(&str, "str", 's', " a string option"),
	}

	var sb strings.Builder
	cli.UsageOptions(&sb, opts, "  ")

	s := sb.String()
	t.Logf("usage: %s", s)
}

func TestParseOptions(t *testing.T) {
	var f bool
	var str string

	opts := []*cli.Option{
		cli.BoolOption(&f, "flag", 'f', "a boolean option"),
		cli.StringOption(&str, "str", 's', " a string option"),
	}

	tests := []struct {
		args []string
		err  bool
		f    bool
		str  string
		n    int
	}{
		{args: []string{}, f: false, str: "", n: 0},
		{args: []string{"-h"}, err: true},
		{args: []string{"--str=foo", "bar"}, str: "foo", n: 1},
		{args: []string{"--s", "foo", "bar"}, str: "foo", n: 2},
		{args: []string{"--s", "foo", "--fl", "bar"}, str: "foo",
			f: true, n: 3},
		{args: []string{"-s"}, err: true},
		{args: []string{"--str"}, err: true},
		{args: []string{"-fs", "foo", "bar"},
			str: "foo", f: true, n: 2},
	}

	var sb strings.Builder
	for _, tc := range tests {
		sb.Reset()
		f = false
		str = ""
		n, err := cli.ParseOptions(&sb, opts, tc.args)
		if tc.err {
			if err == nil {
				t.Fatalf("ParsFlags(&sb, opts, %+v)"+
					" returns no error; want error",
					tc.args)
			}
			t.Logf("ParseFlags(&sb, opts, %+v) error %s",
				tc.args, err)
			continue
		}
		if err != nil {
			t.Fatalf("Parse error %s", err)
		}
		if n != tc.n {
			t.Fatalf(
				"ParseFlags(&s, opts, %+v) returned %d;"+
					" want %d", tc.args, n, tc.n)
		}
		if f != tc.f {
			t.Fatalf(
				"ParseFlags(&s, opts, %+v) has f=%t; want %t",
				tc.args, f, tc.f)
		}
		if str != tc.str {
			t.Fatalf(
				"ParseFlags(&s, opts, %+v) has str=%q; want %q",
				tc.args, str, tc.str)
		}
	}
}