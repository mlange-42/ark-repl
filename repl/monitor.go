package repl

import (
	"encoding/json"
	"strings"

	"github.com/mlange-42/ark-repl/internal/monitor"
)

type localStats struct {
	repl *Repl
}

func (s *localStats) Get() (monitor.Stats, error) {
	out := strings.Builder{}
	s.repl.execCommand(getStats{}, &out)

	st := monitor.Stats{}
	if err := json.Unmarshal([]byte(out.String()), &st); err != nil {
		return st, err
	}
	return st, nil
}

func (s *localStats) Exec(cmd any) {
	out := strings.Builder{}
	s.repl.execCommand(cmd.(Command), &out)
}
