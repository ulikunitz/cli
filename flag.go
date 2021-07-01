package cli

import (
	"fmt"
	"io"
	"sort"
)

type Flag struct {
	Name        string
	Short       rune
	Description string
	HasArg      bool
	Default     string
	Parse       func(arg string) error
}

// BoolFlag initializes a boolean flag. The argument f will be set to false.
func BoolFlag(f *bool, name string, short rune, description string) *Flag {
	*f = false
	return &Flag{
		Name:        name,
		Short:       short,
		Description: description,
		HasArg:      false,
		Default:     "",
		Parse: func(arg string) error {
			*f = true
			return nil
		},
	}
}

// StringFLag creates a string flag. The default value is the value that s has
// when Parse is called.
func StringFlag(s *string, name string, short rune, description string) *Flag {
	return &Flag{
		Name:        name,
		Short:       short,
		Description: description,
		HasArg:      true,
		Default:     *s,
		Parse: func(arg string) error {
			*s = arg
			return nil
		},
	}
}

func findFlag(flags []*Flag, name string) (f *Flag, ok bool) {
	for _, f := range flags {
		if f.Name == name {
			return f, true
		}
		if string(f.Short) == name {
			return f, true
		}
	}
	return nil, false
}

func UsageFlags(w io.Writer, flags []*Flag, indent string) {
	names := make([]string, 0, len(flags))
	for _, f := range flags {
		if f.Short != 0 {
			names = append(names, string(f.Short))
		} else {
			names = append(names, f.Name)
		}

	}
	sort.Strings(names)
	for _, s := range names {
		f, ok := findFlag(flags, s)
		if !ok {
			panic("we should know the string")
		}
		switch {
		case f.Short != 0 && f.Name != "":
			fmt.Fprintf(w, "%s-%c, --%s:\t%s", indent, f.Short,
				f.Name, f.Description)
		case f.Short != 0:
			fmt.Fprintf(w, "%s-%c:\t%s", indent, f.Short,
				f.Description)
		case f.Name != "":
			fmt.Fprintf(w, "%s--%s:\t%s", indent, f.Name,
				f.Description)
		default:
			continue
		}
		if f.Default != "" {
			fmt.Fprintf(w, " (%s)", f.Default)
		}
		fmt.Fprintln(w)
	}
}

// ParseFlags parses the flags and stops at first non-flag.
func ParseFlags(w io.Writer, flags []*Flag, args []string) error {
	panic("TODO")
}
