// SPDX-FileCopyrightText: © 2021 Ulrich Kunitz
//
// SPDX-License-Identifier: BSD-3-Clause

package cli

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

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
		"\nEs ist gut. Es wird noch beser.\n",
	}

	for _, tc := range tests {
		var sb strings.Builder
		_, err := formatText(&sb, tc, 80, "    ")
		if err != nil {
			t.Fatalf("formatText(&sb, %q, %d, %q) error %s",
				tc, 80, "    ", err)
		}
		t.Logf("\n%s", sb.String())
	}
}

var hyphenIndentRegexp = regexp.MustCompile(`(?m)^(\s*)- .*$\n^(\s*)\S+$`)

func termLen(s string) int {
	n := 0
	for _, c := range s {
		if c == '\t' {
			n += 8 - n%8
			continue
		}
		n++
	}
	return n
}

func TestFormatText3(t *testing.T) {
	d, err := os.ReadFile("testdata/description.txt")
	if err != nil {
		t.Fatalf("ReadFile error %s", err)
	}
	s := string(d)
	var sb strings.Builder
	_, err = formatText(&sb, s, 80, "    ")
	if err != nil {
		t.Fatalf("formatText error %s", err)
	}
	o := sb.String()

	found, err := regexp.MatchString(`(?m)^$\n^$\n`, o)
	if err != nil {
		t.Fatalf("regexp.MatchString error %s", err)
	}
	if found {
		t.Fatalf("found 2 empty lines in sequence")
	}

	m := hyphenIndentRegexp.FindAllStringSubmatch(o, -1)
	for _, a := range m {
		if termLen(a[1])+2 != termLen(a[2]) {
			t.Fatalf(
				"indent error in %q; indent line 2 is %d; want %d",
				a[0], termLen(a[2]), termLen(a[1])+2)
		}
	}
}
