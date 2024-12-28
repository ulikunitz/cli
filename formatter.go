// SPDX-FileCopyrightText: © 2021 Ulrich Kunitz
//
// SPDX-License-Identifier: BSD-3-Clause

package cli

import (
	"go/doc/comment"
	"io"
)

// formatText is an interface to the go doc formatter.
func formatText(w io.Writer, s string, lineWidth int, indent string) (n int, err error) {
	var p comment.Parser
	doc := p.Parse(s)
	var pr comment.Printer
	pr.TextWidth = lineWidth
	pr.TextPrefix = indent
	t := pr.Text(doc)
	return w.Write(t)
}
