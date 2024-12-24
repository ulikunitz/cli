// SPDX-FileCopyrightText: © 2021 Ulrich Kunitz
//
// SPDX-License-Identifier: BSD-3-Clause

package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// formatText is a generic text formatter. Text that starts with whitespace is
// formatted as verbatim text without reformatting. The output text is
// indented with the indent string. Don't use tabs in the indent because it will
// be counted as a single character.
func formatText(w io.Writer, s string, lineWidth int, indent string) (n int, err error) {
	const verbatimIndent = "  "
	if lineWidth <= 0 {
		lineWidth = 80
	}
	l := lex(strings.NewReader(s))
	column := 0
	var k int
	for {
		t, err := l.nextToken()
		if err != nil {
			if err != io.EOF {
				return n, err
			}
			if column > 0 {
				k, err = fmt.Fprint(w, "\n")
				n += k
				if err != nil {
					return n, err
				}
			}
			return n, nil
		}
		switch t.typ {
		case tParagraph:
			if column == 0 {
				k, err = fmt.Fprint(w, "\n")
			} else {
				k, err = fmt.Fprint(w, "\n\n")
			}
			n += k
			if err != nil {
				return n, err
			}
			column = 0
		case tVerbatim:
			if column > 0 {
				k, err := fmt.Fprint(w, "\n")
				n += k
				if err != nil {
					return n, err
				}
				column = 0
			}
			k, err = fmt.Fprintf(w, "%s%s%s\n", indent,
				verbatimIndent, t.val)
			n += k
			if err != nil {
				return n, err
			}
		case tWord:
			var (
				size       int
				wordIndent string
			)
			if column == 0 {
				wordIndent = indent
				size = len(indent)
			} else {
				wordIndent = " "
				size = 1
			}
			size += len(t.val)
			if column > 0 && column+size > lineWidth {
				k, err = fmt.Fprint(w, "\n")
				n += k
				if err != nil {
					return n, err
				}
				column = 0
				wordIndent = indent
				size = len(wordIndent) + len(t.val)
			}
			k, err = fmt.Fprint(w, wordIndent, t.val)
			n += k
			if err != nil {
				return n, err
			}
			column += size
		}
	}
}

// The remaining part of the file provides a lexer used by formatText. It uses
// the state as function approach introduced by Rob Pike.

type tokenType int

const (
	tError tokenType = iota
	tWord
	tParagraph
	tVerbatim
)

type token struct {
	typ tokenType
	val string
}

func shortenString(s string) string {
	if len(s) > 8 {
		return s[:8] + "…"
	}
	return s
}

func (t token) String() string {
	switch t.typ {
	case tError:
		return "error"
	case tParagraph:
		return "¶"
	case tVerbatim:
		s := fmt.Sprintf("%q", shortenString(t.val))
		return "`" + s[1:len(s)-1] + "`"
	}
	return fmt.Sprintf("%q", shortenString(t.val))
}

type lexer struct {
	r *bufio.Reader

	buf   []rune
	token token
	err   error
	state stateFn
}

func (l *lexer) readRune() rune {
	if l.err != nil {
		return 0
	}
	var r rune
	r, _, l.err = l.r.ReadRune()
	return r
}

func (l *lexer) unreadRune() {
	if l.err != nil {
		return
	}
	l.err = l.r.UnreadRune()
}

type stateFn func(l *lexer) stateFn

func startState(l *lexer) stateFn {
	r := l.readRune()
	if l.err != nil {
		return nil
	}
	if r == '\n' {
		return nl1State
	}
	if unicode.IsSpace(r) {
		return wsState
	}
	l.buf = append(l.buf, r)
	return wordState
}

func nl1State(l *lexer) stateFn {
	r := l.readRune()
	if l.err != nil {
		return nil
	}
	if r == '\n' {
		return paragraphState
	}
	if unicode.IsSpace(r) {
		return verbatim1State
	}
	l.buf = append(l.buf, r)
	return wordState
}

func wsState(l *lexer) stateFn {
	for {
		r := l.readRune()
		if l.err != nil {
			return nil
		}
		if r == '\n' {
			return nl1State
		}
		if unicode.IsSpace(r) {
			continue
		}
		l.buf = append(l.buf, r)
		return wordState
	}
}

func verbatim1State(l *lexer) stateFn {
	for {
		r := l.readRune()
		if l.err != nil {
			return nil
		}
		if r == '\n' {
			return nl1State
		}
		if unicode.IsSpace(r) {
			continue
		}
		l.buf = append(l.buf, r)
		return verbatimState
	}
}

func verbatimState(l *lexer) stateFn {
	for {
		r := l.readRune()
		if l.err != nil {
			l.token = token{typ: tVerbatim, val: string(l.buf)}
			return nil
		}
		if r == '\n' {
			l.token = token{typ: tVerbatim, val: string(l.buf)}
			return nl1State
		}
		l.buf = append(l.buf, r)
	}
}

func wordState(l *lexer) stateFn {
	for {
		r := l.readRune()
		if l.err != nil {
			l.token = token{typ: tWord, val: string(l.buf)}
			return nil
		}
		if r == '\n' || unicode.IsSpace(r) {
			l.unreadRune()
			l.token = token{typ: tWord, val: string(l.buf)}
			return nil
		}
		l.buf = append(l.buf, r)
	}
}

func paragraphState(l *lexer) stateFn {
	for {
		r := l.readRune()
		if l.err != nil {
			l.token = token{typ: tParagraph}
			return nil
		}
		if r == '\n' {
			continue
		}
		l.unreadRune()
		l.token = token{typ: tParagraph}
		return nl1State
	}
}

func (l *lexer) nextToken() (t token, err error) {
	if l.err != nil {
		return token{}, l.err
	}
	l.buf = l.buf[:0]
	l.token = token{typ: tError}
	if l.state == nil {
		l.state = startState
	}
	for {
		l.state = l.state(l)
		if l.token.typ != tError {
			return l.token, nil
		}
		if l.err != nil {
			return token{}, l.err
		}
		if l.state == nil {
			panic("nil state")
		}
	}
}

func lex(r io.Reader) *lexer {
	return &lexer{r: bufio.NewReader(r)}
}
