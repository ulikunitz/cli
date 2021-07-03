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
	return &OptionError{
		Option: "unrecognized",
		Msg:    fmt.Sprintf("unrecognized option %s", arg),
	}
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
			return 1, &OptionError{Option: option,
				Msg: fmt.Sprintf(
					"option --%s requires no parameter",
					option)}
		}
		if err = found.SetValue(""); err != nil {
			return 1, &OptionError{Option: option,
				Msg: fmt.Sprintf(
					"error setting value for option --%s",
					option),
				Wrapped: err}
		}
		return 1, nil
	}

	var param string
	if k < 0 {
		if len(args) == 1 {
			return 1, &OptionError{Option: option,
				Msg: fmt.Sprintf("no parameter for option --%s",
					option),
			}
		}
		param = args[1]
		argsUsed = 2
	} else {
		argsUsed = 1
		param = arg[k+1:]
	}

	if err = found.SetValue(param); err != nil {
		return argsUsed, &OptionError{
			Option: option,
			Msg: fmt.Sprintf("error setting value %q for option --%s",
				param, option),
			Wrapped: err,
		}
	}

	return argsUsed, nil
}

func handleShortOptions(options []*Option, args []string) (argsUsed int, err error) {
	arg := args[0]
	i := 1
	for _, short := range arg[1:] {
		option := fmt.Sprintf("%c", short)
		var found *Option
		for _, o := range options {
			if o.Short == short {
				found = o
				break
			}
		}
		if found == nil {
			return i, unrecognizedOptionError(option)
		}
		if !found.HasParam {
			if err = found.SetValue(""); err != nil {
				return i, &OptionError{
					Option: option,
					Msg: fmt.Sprintf(
						"error setting value for"+
							" option -%s", option),
					Wrapped: err}
			}
			continue
		}
		if i >= len(args) {
			return i, &OptionError{
				Option: option,
				Msg: fmt.Sprintf(
					"option -%s lacks parameter", option),
			}
		}
		param := args[i]
		i++
		if err = found.SetValue(param); err != nil {
			return i, &OptionError{
				Option: option,
				Msg: fmt.Sprintf("error setting value %s for option %s",
					param, option),
				Wrapped: err,
			}
		}
	}
	return i, nil
}

type OptionError struct {
	Option  string
	Msg     string
	Wrapped error
}

func (err *OptionError) Error() string {
	msg := err.Msg
	if msg == "" {
		msg = fmt.Sprintf("option error for %s", msg)
	}
	if err.Wrapped != nil {
		return fmt.Sprintf("%s: %s", msg, err.Wrapped)
	}
	return msg
}

func (err *OptionError) Is(e error) bool {
	if oe, ok := e.(*OptionError); ok {
		return err.Option == oe.Option
	}
	return errors.Is(err.Wrapped, e)
}

func (err *OptionError) Unwrap() error { return err.Wrapped }

// errorList is represented by a slice of errors. It should be used if multiple
// errors should be returned by a function. It behaves itself as an error.
type errorList []error

// Flatten computes the error value to be returned from a function or method. If
// the error list is empty a nil error is computed, if the list has a single
// error this will be returned. Only if the list contains more than one element
// the list will be returned.
func (err errorList) Flatten() error {
	switch len(err) {
	case 0:
		return nil
	case 1:
		return err[0]
	default:
		return err
	}
}

func (err errorList) Unwrap() error {
	switch len(err) {
	case 0, 1:
		return nil
	default:
		return err[1:].Flatten()
	}
}

func (err errorList) Error() string {
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

func (err errorList) Is(e error) bool {
	if el, ok := e.(errorList); ok  {
		if len(err) != len(el) {
			return false
		}
		for i, cerr := range err {
			if !errors.Is(cerr, el[i]) {
				return false
			}
		}
		return true
	}
	if len(err) == 0 {
		return e == nil
	}
	return errors.Is(err[0], e)
}

// ParseOptions parses the flags and stops at first non-flag or '--'. It returns
// the number of args parsed.
func ParseOptions(w io.Writer, options []*Option, args []string) (n int, err error) {
	i := 0
	var errList errorList
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
				errList = append(errList, err)
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
				errList = append(errList, err)
			}
			continue
		}

		break
	}

	return i, errList.Flatten()
}
