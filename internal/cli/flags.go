package cli

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var ErrHelp = errors.New("help requested")

type FlagSet struct {
	name  string
	flags map[string]*flagDef
	short map[string]string
	Usage func()
}

type flagDef struct {
	long    string
	isBool  bool
	strVal  *string
	boolVal *bool
	intVal  *int
	defStr  string
	usage   string
}

func NewFlagSet(name string) *FlagSet {
	return &FlagSet{
		name:  name,
		flags: make(map[string]*flagDef),
		short: make(map[string]string),
	}
}

func (fs *FlagSet) String(long, def, usage string) *string {
	val := def
	fd := &flagDef{long: long, strVal: &val, defStr: def, usage: usage}
	fs.flags[long] = fd
	return fd.strVal
}

func (fs *FlagSet) Bool(long string, def bool, usage string) *bool {
	val := def
	fd := &flagDef{long: long, isBool: true, boolVal: &val, usage: usage}
	fs.flags[long] = fd
	return fd.boolVal
}

func (fs *FlagSet) Int(long string, def int, usage string) *int {
	val := def
	fd := &flagDef{long: long, intVal: &val, usage: usage}
	fs.flags[long] = fd
	return fd.intVal
}

func (fs *FlagSet) Short(long, short, def, usage string) *string {
	ptr := fs.String(long, def, usage)
	fs.short[short] = long
	return ptr
}

func (fs *FlagSet) Parse(args []string) ([]string, error) {
	var remaining []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			remaining = append(remaining, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			remaining = append(remaining, args[i:]...)
			break
		}

		if arg == "-h" || arg == "--help" {
			if fs.Usage != nil {
				fs.Usage()
			}
			return nil, ErrHelp
		}

		long, inline, err := splitFlagArg(arg)
		if err != nil {
			return nil, err
		}

		fd, err := fs.lookup(long)
		if err != nil {
			return nil, err
		}

		if fd.isBool {
			if inline != nil {
				v, err := strconv.ParseBool(*inline)
				if err != nil {
					return nil, fmt.Errorf("invalid boolean value for --%s: %q", fd.long, *inline)
				}
				*fd.boolVal = v
			} else {
				*fd.boolVal = true
			}
			continue
		}

		var raw string
		switch {
		case inline != nil:
			raw = *inline
		case i+1 < len(args):
			i++
			raw = args[i]
		default:
			return nil, fmt.Errorf("flag --%s requires a value", fd.long)
		}

		if fd.strVal != nil {
			*fd.strVal = raw
			continue
		}
		if fd.intVal != nil {
			n, err := strconv.Atoi(raw)
			if err != nil {
				return nil, fmt.Errorf("invalid integer value for --%s: %q", fd.long, raw)
			}
			*fd.intVal = n
		}
	}
	return remaining, nil
}

func splitFlagArg(arg string) (long string, inline *string, err error) {
	switch {
	case strings.HasPrefix(arg, "--"):
		name := arg[2:]
		if name == "" {
			return "", nil, fmt.Errorf("empty flag name")
		}
		if idx := strings.Index(name, "="); idx >= 0 {
			v := name[idx+1:]
			return name[:idx], &v, nil
		}
		return name, nil, nil
	case strings.HasPrefix(arg, "-") && len(arg) == 2:
		return arg[1:], nil, nil
	case strings.HasPrefix(arg, "-"):
		name := strings.TrimPrefix(arg, "-")
		if idx := strings.Index(name, "="); idx >= 0 {
			name = name[:idx]
		}
		return "", nil, fmt.Errorf("unknown flag %q (long flags use --%s)", arg, name)
	default:
		return "", nil, fmt.Errorf("invalid flag %q", arg)
	}
}

func (fs *FlagSet) lookup(name string) (*flagDef, error) {
	if len(name) == 1 {
		if long, ok := fs.short[name]; ok {
			name = long
		}
	}
	fd, ok := fs.flags[name]
	if !ok {
		if len(name) == 1 {
			return nil, fmt.Errorf("unknown flag -%s", name)
		}
		return nil, fmt.Errorf("unknown flag --%s", name)
	}
	return fd, nil
}

func (fs *FlagSet) FlagLines() []string {
	names := make([]string, 0, len(fs.flags))
	for name := range fs.flags {
		names = append(names, name)
	}
	sort.Strings(names)

	lines := make([]string, 0, len(names))
	for _, name := range names {
		fd := fs.flags[name]
		line := "  --" + fd.long
		if fd.defStr != "" {
			line += fmt.Sprintf("  (default %q)", fd.defStr)
		} else if fd.isBool && fd.boolVal != nil && !*fd.boolVal {
			line += "  (default false)"
		} else if fd.intVal != nil {
			line += fmt.Sprintf("  (default %d)", *fd.intVal)
		}
		if fd.usage != "" {
			line += "  " + fd.usage
		}
		lines = append(lines, line)
	}
	return lines
}
