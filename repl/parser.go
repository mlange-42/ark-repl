package repl

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func parseInput(input string, commandRegistry map[string]command) (command, bool, error) {
	tokens := strings.Fields(input)
	if len(tokens) < 1 {
		return nil, false, fmt.Errorf("no command provided")
	}

	cmdName := tokens[0]
	cmdStruct, ok := commandRegistry[cmdName]
	if !ok {
		return nil, false, fmt.Errorf("unknown command: %s", cmdName)
	}

	cmdVal := reflect.New(reflect.TypeOf(cmdStruct)).Elem()

	if len(tokens) == 1 {
		cmd, ok := cmdVal.Interface().(command)
		if !ok {
			return nil, false, fmt.Errorf("command %s does not implement interface", cmdName)
		}
		return cmd, false, nil
	}

	if cmdVal.Type() == reflect.TypeFor[hlp]() {
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
			return nil, false, fmt.Errorf("unknown subcommand: %s", subcmdName)
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

		switch field.Kind() {
		case reflect.Bool:
			if len(kv) == 1 {
				field.SetBool(true)
			} else {
				if b, err := strconv.ParseBool(kv[1]); err == nil {
					field.SetBool(b)
				} else {
					return nil, false, fmt.Errorf("invalid value for bool option: %s", kv[1])
				}
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if v, err := strconv.Atoi(kv[1]); err == nil {
				field.SetInt(int64(v))
			} else {
				return nil, false, fmt.Errorf("invalid value for int option: %s", kv[1])
			}
		case reflect.Float64, reflect.Float32:
			if v, err := strconv.ParseFloat(kv[1], 64); err == nil {
				field.SetFloat(v)
			} else {
				return nil, false, fmt.Errorf("invalid value for int option: %s", kv[1])
			}
		case reflect.String:
			field.SetString(kv[1])
		default:
			return nil, false, fmt.Errorf("unsupported argument type %s", field.Kind().String())
		}
		i++
	}

	exec, ok := cmdVal.Interface().(command)
	if !ok {
		return nil, false, fmt.Errorf("command %s does not implement interface", cmdName)
	}
	return exec, false, nil
}
