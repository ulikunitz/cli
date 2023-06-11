// SPDX-FileCopyrightText: © 2021 Ulrich Kunitz
//
// SPDX-License-Identifier: BSD-3-Clause

package cli

import (
	"io"
	"strings"
	"testing"
)

func tokenize(s string) (a []token, err error) {
	l := lex(strings.NewReader(s))
	for {
		t, err := l.nextToken()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			return a, nil
		}
		a = append(a, t)
	}
}

func TestLexer(t *testing.T) {
	tests := []struct {
		s string
		t []token
	}{
		{
			s: "foo bar\nfoo bar\n\n x := 3\n  x := 4\n",
			t: []token{
				{typ: tWord, val: "foo"},
				{typ: tWord, val: "bar"},
				{typ: tWord, val: "foo"},
				{typ: tWord, val: "bar"},
				{typ: tParagraph},
				{typ: tVerbatim, val: "x := 3"},
				{typ: tVerbatim, val: "x := 4"},
			},
		},
	}

	for _, tc := range tests {
		tokens, err := tokenize(tc.s)
		if err != nil {
			t.Fatalf("tokenize(%q) error %s", tc.s, err)
		}
		for i, tok := range tokens {
			if tok != tc.t[i] {
				t.Fatalf("[%d] got token %+v; want %+v",
					i, tok, tc.t[i])
			}
		}
	}
}

func TestFormatText(t *testing.T) {
	const s = `
Im Wald fühlte er sich wohl. Die Bäume warfen Schatten und schützten ihn vor der Sonne.
Es fühlte sich an als ob sie ihn auch von den E-Mails und Telekonferenzen schützten.

Hier beginnt der nächste Paragraph.
  x := 3
  x := 4
`
	var sb strings.Builder
	_, err := formatText(&sb, s, 80, "    ")
	if err != nil {
		t.Fatalf("formatText error %s", err)
	}
	t.Logf("\n%s", sb.String())
}

func TestFormatText2(t *testing.T) {
	tests := []string{
		"a boolean option",
	}

	for _, tc := range tests {
		var sb strings.Builder
		_, err := formatText(&sb, tc, 80, "    ")
		if err != nil {
			t.Fatalf("formatText(&sb, %q, %d, %q) error %s",
				tc, 80, "    ", err)
		}
	}
}
