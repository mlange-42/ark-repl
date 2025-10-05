package repl

import (
	"encoding/json"
	"strings"

	"github.com/mlange-42/ark-repl/internal/monitor"
)

type localConnection struct {
	repl *Repl
}

func (s *localConnection) Get() (monitor.Stats, error) {
	out := strings.Builder{}
	s.repl.execCommand(getStats{}, &out)

	st := monitor.Stats{}
	if err := json.Unmarshal([]byte(out.String()), &st); err != nil {
		return st, err
	}
	return st, nil
}

func (s *localConnection) Exec(cmd string) error {
	out := strings.Builder{}
	command := s.repl.commands[cmd]
	s.repl.execCommand(command.command, &out)
	return nil
}
