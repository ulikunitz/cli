package cli

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

type Option struct {
	Name        string
	Short       rune
	Description string
	HasParam    bool
	Default     string
	SetValue    func(arg string) error
}

// BoolOption initializes a boolean flag. The argument f will be set to false.
func BoolOption(f *bool, name string, short rune, description string) *Option {
	*f = false
	return &Option{
		Name:        name,
		Short:       short,
		Description: description,
		HasParam:    false,
		Default:     "",
		SetValue: func(arg string) error {
			*f = true
			return nil
		},
	}
}

// StringOption creates a string flag. The default value is the value that s has
// when Parse is called.
func StringOption(s *string, name string, short rune, description string) *Option {
	return &Option{
		Name:        name,
		Short:       short,
		Description: description,
		HasParam:    true,
		Default:     *s,
		SetValue: func(arg string) error {
			*s = arg
			return nil
		},
	}
}

func findOption(flags []*Option, name string) (f *Option, ok bool) {
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

func UsageOptions(w io.Writer, opts []*Option, indent string) {
	names := make([]string, 0, len(opts))
	for _, f := range opts {
		if f.Short != 0 {
			names = append(names, string(f.Short))
		} else {
			names = append(names, f.Name)
		}

	}
	sort.Strings(names)
	for _, s := range names {
		f, ok := findOption(opts, s)
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

func unrecognizedOptionError(arg string) error {
	return fmt.Errorf("unregognized option %s", arg)
}

func handleLongOption(options []*Option, args []string) (argsUsed int, err error) {
	var option string
	arg := args[0]
	k := strings.IndexByte(arg, '=')
	if k >= 0 {
		option = arg[2:k]
	} else {
		option = arg[2:]
	}
	if option == "" {
		return 1, unrecognizedOptionError(arg)
	}
	var found *Option
	for _, o := range options {
		if strings.HasPrefix(o.Name, option) {
			if found != nil {
				return 0, unrecognizedOptionError(arg)
			}
			found = o
		}
	}
	if found == nil {
		return 1, unrecognizedOptionError(arg)
	}

	option = found.Name

	if !found.HasParam {
		if k >= 0 {
			return 1, fmt.Errorf(
				"option --%s requires no parameter", option)
		}
		if err = found.SetValue(""); err != nil {
			return 1, fmt.Errorf(
				"error setting value for option --%s: %w",
				option, err)
		}
		return 1, nil
	}

	var param string
	if k < 0 {
		if len(args) == 1 {
			return 1, fmt.Errorf("no parameter for option --%s",
				option)
		}
		param = args[1]
		argsUsed = 2
	} else {
		argsUsed = 1
		param = arg[k+1:]
	}

	if err = found.SetValue(param); err != nil {
		return argsUsed, fmt.Errorf(
			"error setting value %q for option --%s",
			param, option)
	}

	return argsUsed, nil
}

func handleShortOptions(options []*Option, args []string) (argsUsed int, err error) {
	arg := args[0]
	i := 1
	for _, option := range arg[1:] {
		var found *Option
		for _, o := range options {
			if o.Short == option {
				found = o
				break
			}
		}
		if found == nil {
			return i, fmt.Errorf("unrecognized option -%c", option)
		}
		if !found.HasParam {
			if err = found.SetValue(""); err != nil {
				return i, fmt.Errorf(
					"error setting value for option -%c: %w",
					option, err)
			}
			continue
		}
		if i >= len(args) {
			return i, fmt.Errorf("option -%c lacks parameter", option)
		}
		param := args[i]
		i++
		if err = found.SetValue(param); err != nil {
			return i, fmt.Errorf(
				"error setting value %s for option -%c: %w",
				param, option, err)
		}
	}
	return i, nil
}

type ParseOptionsErrors []error

func (err ParseOptionsErrors) Error() string {
	if len(err) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, e := range err[:len(err)-1] {
		fmt.Fprintln(&sb, e)
	}
	fmt.Fprint(&sb, err[len(err)-1])
	return sb.String()
}

func (err ParseOptionsErrors) Is(e error) bool {
	if terr, ok := e.(ParseOptionsErrors); ok {
		if len(err) != len(terr) {
			return false
		}
		for i, cerr := range err {
			if !errors.Is(cerr, terr[i]) {
				return false
			}
		}
		return true
	}
	for _, cerr := range err {
		if errors.Is(cerr, e) {
			return true
		}
	}
	return false
}

// ParseOptions parses the flags and stops at first non-flag or '--'. It returns
// the number of args parsed.
func ParseOptions(w io.Writer, options []*Option, args []string) (n int, err error) {
	i := 0
	var pferr ParseOptionsErrors
	for i < len(args) {
		a := args[i]
		if strings.HasPrefix(a, "--") {
			if a == "--" {
				return i + 1, nil
			}
			argsUsed, err := handleLongOption(options, args[i:])
			i += argsUsed
			if err != nil {
				fmt.Fprintln(w, err)
				pferr = append(pferr, err)
			}
			continue
		}

		if strings.HasPrefix(a, "-") {
			if a == "-" {
				return i, nil
			}

			argsUsed, err := handleShortOptions(options, args[i:])
			i += argsUsed
			if err != nil {
				fmt.Fprintln(w, err)
				pferr = append(pferr, err)
			}
			continue
		}

		break
	}

	switch len(pferr) {
	case 0:
		return i, nil
	case 1:
		return i, pferr[0]
	default:
		return i, pferr
	}

}
