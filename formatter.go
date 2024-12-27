// SPDX-FileCopyrightText: © 2021 Ulrich Kunitz
//
// SPDX-License-Identifier: BSD-3-Clause

package cli

import (
	"go/doc/comment"
	"io"
)

// formatText is a generic text formatter. Text that starts with whitespace is
// formatted as verbatim text without reformatting. The output text is
// indented with the indent string. Don't use tabs in the indent because it will
// be counted as a single character.
func formatText(w io.Writer, s string, lineWidth int, indent string) (n int, err error) {
	var p comment.Parser
	doc := p.Parse(s)
	var pr comment.Printer
	pr.TextWidth = lineWidth
	pr.TextPrefix = indent
	t := pr.Text(doc)
	return w.Write(t)
}
