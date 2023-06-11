package cli

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// Option represents a single option.
type Option struct {
	// long name of the option; must be used with prefix '--'
	Name string
	// short option; must be in [A-Za-z0-9]
	Short rune
	// long names additional to name
	Names []string
	// short runes additional to Names
	Shorts []rune
	// Custom usage information about options
	UsageInfo string
	// description of the option
	Description string
	// HasParam defines whether the option has parameter
	HasParam bool
	// OptionalParam
	OptionalParam bool
	// ParamType describes the type of the parameter
	ParamType string
	// Default param value.
	Default string
	// SetValue set the value to the parameter string given and informs
	// whether there was a parameter or not.
	SetValue func(name string, param string, noParam bool) error
	// ResetValue can be used to reset the value. If it is nil then
	// opt.SetValue(opt.Default, false) will be called.
	ResetValue func()
}

// AllShorts returns all short option names in lexicographic order.
func (opt *Option) AllShorts() []rune {
	n := len(opt.Shorts)
	if opt.Short != 0 {
		n++
	}
	s := make([]rune, 0, n)
	if opt.Short != 0 {
		s = append(s, opt.Short)
	}
	s = append(s, opt.Shorts...)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	return s
}

// AllNames returns all long option names in lexicographic order.
func (opt *Option) AllNames() []string {
	n := len(opt.Names)
	if opt.Name != "" {
		n++
	}
	s := make([]string, 0, n)
	if opt.Name != "" {
		s = append(s, opt.Name)
	}
	s = append(s, opt.Names...)
	sort.Strings(s)
	return s
}

func (opt *Option) hasShortString(n string) bool {
	if n == "" || n == "\x00" {
		return false
	}
	if string(opt.Short) == n {
		return true
	}
	for _, r := range opt.Shorts {
		s := string(r)
		if s == n {
			return true
		}
	}
	return false
}

func (opt *Option) hasName(name string) bool {
	if name == "" {
		return false
	}
	if name == opt.Name {
		return true
	}
	for _, n := range opt.Names {
		if n == name {
			return true
		}
	}
	return false
}

func verifyShort(s rune) error {
	if s == 0 {
		return nil
	}
	if 'a' <= s && s <= 'z' {
		return nil
	}
	if 'A' <= s && s <= 'Z' {
		return nil
	}
	if '0' <= s && s <= '9' {
		return nil
	}
	return fmt.Errorf("short option %q (%U) not supported", s, s)
}

func validShort(s rune) {
	err := verifyShort(s)
	if err != nil {
		panic(err)
	}
}

const resetName = "<reset>"

// Reset calls ResetValue if defined or SetValue with with the default argument.
func (opt *Option) Reset() error {
	if opt.ResetValue != nil {
		opt.ResetValue()
		return nil
	}
	return opt.SetValue(resetName, opt.Default, false)
}

// BoolOption initializes a boolean flag. The argument f will be set to false.
func BoolOption(f *bool, name string, short rune, description string) *Option {
	validShort(short)
	*f = false
	return &Option{
		Name:        name,
		Short:       short,
		Description: description,
		HasParam:    false,
		Default:     "",
		SetValue: func(name, arg string, noParam bool) error {
			*f = true
			return nil
		},
		ResetValue: func() { *f = false },
	}
}

// StringOption creates a string flag. The default value is the value that s has
// when Parse is called.
func StringOption(s *string, name string, short rune, description string) *Option {
	validShort(short)
	return &Option{
		Name:        name,
		Short:       short,
		Description: description,
		HasParam:    true,
		ParamType:   "string",
		Default:     *s,
		SetValue: func(name, arg string, noParam bool) error {
			*s = arg
			return nil
		},
	}
}

// IntOption creates an integer flag. The default value is the value of n when
// this function is called. Integers in the form of 0b101, 0xf5 or 0234 are
// supported.
func IntOption(n *int, name string, short rune, description string) *Option {
	validShort(short)
	const intSize = 32 << (^uint(0) >> 63)
	var def string
	if *n != 0 {
		def = fmt.Sprintf("%d", *n)
	} else {
		def = ""
	}
	return &Option{
		Name:        name,
		Short:       short,
		Description: description,
		HasParam:    true,
		ParamType:   "int",
		Default:     def,
		SetValue: func(name, arg string, noParam bool) error {
			i, err := strconv.ParseInt(arg, 0, intSize)
			if err != nil {
				return err
			}
			*n = int(i)
			return nil
		},
	}
}

// Float64Option creates a flag with a floating point value. The default value
// is the value of f when called. All forms of floating point numbers valid in
// the Go language are supported.
func Float64Option(f *float64, name string, short rune, description string) *Option {
	validShort(short)
	var def string
	if *f != 0 {
		def = fmt.Sprintf("%g", *f)
	} else {
		def = ""
	}
	return &Option{
		Name:        name,
		Short:       short,
		Description: description,
		HasParam:    true,
		ParamType:   "float64",
		Default:     def,
		SetValue: func(name, arg string, noParam bool) error {
			x, err := strconv.ParseFloat(arg, 64)
			if err != nil {
				return err
			}
			*f = x
			return nil
		},
	}

}

func findOption(flags []*Option, name string) (f *Option, ok bool) {
	for _, f := range flags {
		if f.hasName(name) {
			return f, true
		}
		if f.hasShortString(name) {
			return f, true
		}
	}
	return nil, false
}

// Usage returns the one-line string for the option.
func (opt *Option) Usage() string {
	if opt.UsageInfo != "" {
		return opt.UsageInfo
	}
	var ptype string
	if opt.HasParam {
		ptype = opt.ParamType
		if ptype == "" {
			ptype = "param"
		}
	}
	var sb strings.Builder
	i := 0

	if shorts := opt.AllShorts(); len(shorts) > 0 {
		fmt.Fprint(&sb, "-")
		i++
		for _, r := range shorts {
			fmt.Fprintf(&sb, "%c", r)
		}
		if opt.HasParam {
			if opt.OptionalParam {
				fmt.Fprintf(&sb, " [%s]", ptype)
			} else {
				fmt.Fprintf(&sb, " %s", ptype)
			}
		}
	}

	if names := opt.AllNames(); len(names) > 0 {
		for _, n := range names {
			if i > 0 {
				fmt.Fprintf(&sb, ", ")
			}
			fmt.Fprintf(&sb, "--%s", n)
			i++
			if opt.HasParam {
				if opt.OptionalParam {
					fmt.Fprintf(&sb, "[=%s]", ptype)
				} else {
					fmt.Fprintf(&sb, "=%s", ptype)
				}
			}
		}
	}
	if opt.Default != "" {
		fmt.Fprintf(&sb, " (default %s)", opt.Default)
	}
	return sb.String()
}

// UsageOptions returns a textual list of all options sorted by alphabet. Usage
// information for an option will be preceded by indent1 and the description by
// indent1+indent2 formatted on 80 character lines.
func UsageOptions(w io.Writer, opts []*Option, indent1, indent2 string) (n int, err error) {
	names := make([]string, 0, len(opts)+32)
	for _, f := range opts {
		shorts := f.AllShorts()
		if len(shorts) > 0 {
			names = append(names, string(shorts[0]))
			continue
		}
		fNames := f.AllNames()
		if len(fNames) > 0 {
			names = append(names, fNames[0])
		}
	}
	sort.Strings(names)
	for _, s := range names {
		f, ok := findOption(opts, s)
		if !ok {
			panic("we should know the string")
		}
		k, err := fmt.Fprint(w, indent1)
		n += k
		if err != nil {
			return n, err
		}
		k, err = fmt.Fprint(w, f.Usage())
		n += k
		if err != nil {
			return n, err
		}
		k, err = fmt.Fprintln(w)
		n += k
		if err != nil {
			return n, err
		}
		k, err = formatText(w, f.Description, 80, indent1+indent2)
		n += k
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

func unrecognizedOptionError(arg string) error {
	return &OptionError{
		Option: "unrecognized",
		Msg:    fmt.Sprintf("unrecognized option %s", arg),
	}
}

func handleLongOption(options []*Option, args []string) (argsUsed int, err error) {
	for i, a := range args[1:] {
		if len(a) > 0 && a[0] == '-' {
			args = args[:i+1]
			break
		}
	}
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
		for _, name := range o.AllNames() {
			if strings.HasPrefix(name, option) {
				if found != nil {
					return 0, unrecognizedOptionError(arg)
				}
				option = name
				found = o
			}
		}
	}
	if found == nil {
		return 1, unrecognizedOptionError(arg)
	}

	if !found.HasParam {
		if k >= 0 {
			return 1, &OptionError{Option: option,
				Msg: fmt.Sprintf(
					"option --%s requires no parameter",
					option)}
		}
		if err = found.SetValue(option, "", true); err != nil {
			return 1, &OptionError{Option: option,
				Msg: fmt.Sprintf(
					"error setting value for option --%s",
					option),
				Wrapped: err}
		}
		return 1, nil
	}

	var (
		param   string
		noParam bool
	)
	if k < 0 {
		if len(args) == 1 {
			if !found.OptionalParam {
				return 1, &OptionError{Option: option,
					Msg: fmt.Sprintf("no parameter for option --%s",
						option),
				}
			}
			noParam = true
			argsUsed = 1
		} else {
			param = args[1]
			argsUsed = 2
		}
	} else {
		param = arg[k+1:]
		argsUsed = 1
	}

	if err = found.SetValue(option, param, noParam); err != nil {
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
	for i, a := range args[1:] {
		if len(a) > 0 && a[0] == '-' {
			args = args[:i+1]
			break
		}
	}
	arg := args[0]
	i := 1
	for _, short := range arg[1:] {
		option := string(short)
		var found *Option
		for _, o := range options {
			if o.hasShortString(option) {
				found = o
				break
			}
		}
		if found == nil {
			return i, unrecognizedOptionError(option)
		}

		if !found.HasParam {
			if err = found.SetValue(option, "", true); err != nil {
				return i, &OptionError{
					Option: option,
					Msg: fmt.Sprintf(
						"error setting value for"+
							" option -%s", option),
					Wrapped: err}
			}
			continue
		}

		var (
			param   string
			noParam bool
		)
		if i >= len(args) {
			if !found.OptionalParam {
				return i, &OptionError{
					Option: option,
					Msg: fmt.Sprintf(
						"option -%s lacks parameter",
						option),
				}
			}
			noParam = true
		} else {
			param = args[i]
			i++
		}
		if err = found.SetValue(option, param, noParam); err != nil {
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

// OptionError provides information regarding an option error.
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

// Is checks whether the error is actually one of the errors provided.
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
	if el, ok := e.(errorList); ok {
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

// ResetOptions resets all options to the default. It may be useful before you
// are executing Parse a second time on an option set.
func ResetOptions(options []*Option) error {
	var errList errorList
	for _, o := range options {
		err := o.Reset()
		if err != nil {
			errList = append(errList, err)
		}
	}
	return errList.Flatten()
}

// ParseOptions parses the flags and stops at first non-flag or '--'. It returns
// the number of args parsed.
func ParseOptions(options []*Option, args []string) (n int, err error) {
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
				errList = append(errList, err)
			}
			continue
		}

		break
	}

	return i, errList.Flatten()
}
