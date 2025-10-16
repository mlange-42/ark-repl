package repl

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func parseInput(input string, commandRegistry map[string]commandEntry) (Command, bool, error) {
	if script, ok := parseScript(input); ok {
		return script, false, nil
	}
	tokens := strings.Fields(input)
	if len(tokens) < 1 {
		return nil, false, fmt.Errorf("no command provided")
	}

	cmdName := tokens[0]
	cmdStruct, ok := commandRegistry[cmdName]
	if !ok {
		return nil, false, fmt.Errorf("unknown command: %s", cmdName)
	}

	originalVal := reflect.ValueOf(cmdStruct.command)
	cmdVal := reflect.New(reflect.TypeOf(cmdStruct.command)).Elem()
	cmdVal.Set(originalVal)

	if len(tokens) == 1 {
		if err := setDefaults(cmdVal); err != nil {
			return nil, false, err
		}
		cmd, ok := cmdVal.Interface().(Command)
		if !ok {
			return nil, false, fmt.Errorf("command %s does not implement interface Command", cmdName)
		}
		return cmd, false, nil
	}

	if cmdVal.Type() == reflect.TypeFor[help]() {
		cmd, _, err := parseInput(strings.Join(tokens[1:], " "), commandRegistry)
		return cmd, true, err
	}

	// Parse subcommand
	i := 1
	for i < len(tokens) {
		kv := strings.SplitN(tokens[i], "=", 2)
		if len(kv) > 1 {
			break
		}
		subcmdName := tokens[i]
		subcmdField := cmdVal.FieldByNameFunc(func(s string) bool { return strings.ToLower(s) == subcmdName })
		if !subcmdField.IsValid() {
			return nil, false, fmt.Errorf("unknown subcommand or bool option: %s", subcmdName)
		}
		if subcmdField.Kind() == reflect.Bool {
			break
		}
		if subcmdField.Kind() != reflect.Struct {
			return nil, false, fmt.Errorf("unknown subcommand: %s", subcmdName)
		}
		cmdVal = subcmdField
		i++
	}

	// Fill defaults
	if err := setDefaults(cmdVal); err != nil {
		return nil, false, err
	}

	// Parse args
	for i < len(tokens) {
		kv := strings.SplitN(tokens[i], "=", 2)
		cmdName = kv[0]
		field := cmdVal.FieldByNameFunc(func(s string) bool { return strings.ToLower(s) == cmdName })
		if !field.IsValid() || !field.CanSet() {
			return nil, false, fmt.Errorf("invalid option: %s", cmdName)
		}

		if len(kv) != 2 && field.Kind() != reflect.Bool {
			return nil, false, fmt.Errorf("invalid option syntax: %s", tokens[i])
		}

		if err := setField(field, kv); err != nil {
			return nil, false, err
		}
		i++
	}

	exec, ok := cmdVal.Interface().(Command)
	if !ok {
		return nil, false, fmt.Errorf("command %s does not implement interface Command", cmdName)
	}
	return exec, false, nil
}

func setDefaults(cmdVal reflect.Value) error {
	for i := range cmdVal.NumField() {
		typeField := cmdVal.Type().Field(i)
		value, ok := typeField.Tag.Lookup("default")
		if !ok {
			continue
		}
		field := cmdVal.Field(i)

		if err := setField(field, []string{typeField.Name, value}); err != nil {
			return err
		}
	}
	return nil
}

func setField(field reflect.Value, kv []string) error {
	switch field.Kind() {
	case reflect.Bool:
		if len(kv) == 1 {
			field.SetBool(true)
		} else {
			if b, err := strconv.ParseBool(kv[1]); err == nil {
				field.SetBool(b)
			} else {
				return fmt.Errorf("invalid value for bool option '%s': %s", kv[0], kv[1])
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v, err := strconv.Atoi(kv[1]); err == nil {
			field.SetInt(int64(v))
		} else {
			return fmt.Errorf("invalid value for int option '%s': %s", kv[0], kv[1])
		}
	case reflect.Float64, reflect.Float32:
		if v, err := strconv.ParseFloat(kv[1], 64); err == nil {
			field.SetFloat(v)
		} else {
			return fmt.Errorf("invalid value for float option '%s': %s", kv[0], kv[1])
		}
	case reflect.String:
		field.SetString(kv[1])
	case reflect.Slice:
		elemType := field.Type().Elem()
		rawValues := strings.Split(kv[1], ",")
		slice := reflect.MakeSlice(field.Type(), 0, len(rawValues))
		for _, raw := range rawValues {
			var val reflect.Value
			switch elemType.Kind() {
			case reflect.String:
				val = reflect.ValueOf(raw)
			case reflect.Bool:
				b, err := strconv.ParseBool(raw)
				if err != nil {
					return fmt.Errorf("invalid value for []bool option '%s': %s", kv[0], raw)
				}
				val = reflect.ValueOf(b)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				i, err := strconv.Atoi(raw)
				if err != nil {
					return fmt.Errorf("invalid value for []int option '%s': %s", kv[0], raw)
				}
				val = reflect.ValueOf(i)
			case reflect.Float64, reflect.Float32:
				f, err := strconv.ParseFloat(raw, 64)
				if err != nil {
					return fmt.Errorf("invalid value for []float option '%s': %s", kv[0], raw)
				}
				val = reflect.ValueOf(f)
			default:
				return fmt.Errorf("unsupported slice element type: %s", elemType.Kind())
			}
			slice = reflect.Append(slice, val)
		}
		field.Set(slice)
	default:
		return fmt.Errorf("unsupported argument type %s for option '%s'", field.Kind().String(), kv[0])
	}
	return nil
}

func extractHelp(cmd Command, out *strings.Builder) error {
	commands := []string{}
	cmdHelp := []string{}
	options := []string{}

	cmdVal := reflect.ValueOf(cmd)

	for i := range cmdVal.NumField() {
		field := cmdVal.Field(i)
		typeField := cmdVal.Type().Field(i)

		if field.Kind() == reflect.Struct {
			cmdName := strings.ToLower(typeField.Name)
			commands = append(commands, cmdName)
			interf, ok := field.Interface().(Command)
			if !ok {
				return fmt.Errorf("command %s does not implement interface Command", cmdName)
			}
			out := strings.Builder{}
			interf.Help(&out)
			parts := strings.SplitN(out.String(), "\n", 2)
			var helpText string
			if len(parts) > 0 {
				helpText = parts[0]
			}
			cmdHelp = append(cmdHelp, helpText)
			continue
		}

		var kind string
		switch field.Kind() {
		case reflect.Bool, reflect.String:
			kind = field.Kind().String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			kind = "int"
		case reflect.Float32, reflect.Float64:
			kind = "float"
		case reflect.Slice:
			elemType := field.Type().Elem()
			switch elemType.Kind() {
			case reflect.Bool, reflect.String:
				kind = elemType.Kind().String() + "s"
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				kind = "ints"
			case reflect.Float32, reflect.Float64:
				kind = "floats"
			default:
				kind = "unknowns"
			}
		default:
			kind = "unknown"
		}

		help, ok := typeField.Tag.Lookup("help")
		if ok {
			help += " "
		}
		defaultValue, ok := typeField.Tag.Lookup("default")
		if ok {
			defaultValue = "Default: " + defaultValue
		}

		options = append(options,
			fmt.Sprintf("%-14s%-7s  %s%s",
				strings.ToLower(typeField.Name),
				kind, help, defaultValue,
			))
	}

	cmd.Help(out)
	if len(commands) > 0 {
		fmt.Fprintln(out, "\nCommands:")
		for i, c := range commands {
			fmt.Fprintf(out, "  %-12s %s\n", c, cmdHelp[i])
		}
	}
	if len(options) > 0 {
		fmt.Fprintln(out, "\nOptions:")
		for _, o := range options {
			fmt.Fprintf(out, "  %s\n", o)
		}
	}

	return nil
}
