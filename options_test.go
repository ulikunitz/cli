// SPDX-FileCopyrightText: © 2021 Ulrich Kunitz
//
// SPDX-License-Identifier: BSD-3-Clause

package cli_test

import (
	"errors"
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
	cli.UsageOptions(&sb, opts, "  ", "  ")

	s := sb.String()
	t.Logf("usage:\n%s", s)
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
		err  *cli.OptionError
		f    bool
		str  string
		n    int
	}{
		{args: []string{}, f: false, str: "", n: 0},
		{args: []string{"-h"}, err: &cli.OptionError{Option: "unrecognized"}},
		{args: []string{"--str=foo", "bar"}, str: "foo", n: 1},
		{args: []string{"--s", "foo", "bar"}, str: "foo", n: 2},
		{args: []string{"--s", "foo", "--fl", "bar"}, str: "foo",
			f: true, n: 3},
		{args: []string{"-s"}, err: &cli.OptionError{Option: "s"}},
		{args: []string{"--str"}, err: &cli.OptionError{Option: "str"}},
		{args: []string{"-fs", "foo", "bar"},
			str: "foo", f: true, n: 2},
	}

	var sb strings.Builder
	for _, tc := range tests {
		sb.Reset()
		f = false
		str = ""
		n, err := cli.ParseOptions(opts, tc.args)
		if tc.err != nil {
			if err == nil {
				t.Fatalf("ParsFlags(&sb, opts, %+v)"+
					" returns no error; want error",
					tc.args)
			}
			if !errors.Is(err, tc.err) {
				t.Fatalf("ParsFlags(&sb, opts, %+v)"+
					" returns error %#v; want error %#v",
					tc.args, err, tc.err)

			}
			t.Logf("ParseFlags(&sb, opts, %+v) error %s; want %s",
				tc.args, err, tc.err)
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

func TestResetOptions(t *testing.T) {
	var f bool
	var str string

	opts := []*cli.Option{
		cli.BoolOption(&f, "flag", 'f', "a boolean option"),
		cli.StringOption(&str, "str", 's', " a string option"),
	}

	f = true
	str = "foobar"

	if err := cli.ResetOptions(opts); err != nil {
		t.Fatalf("ResetOptions error %s", err)
	}

	if f {
		t.Errorf("f is %t after reset; want %t", true, false)
	}
	if str != "" {
		t.Errorf("str is %q after reset; want %q", str, "")
	}
}
