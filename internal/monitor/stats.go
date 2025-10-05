package monitor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	arkstats "github.com/mlange-42/ark/ecs/stats"
)

type StatsGetter interface {
	Get() (Stats, error)
	Exec(any)
}

type Stats struct {
	Ticks int
	Stats *arkstats.World
}

type RemoteStats struct {
	Conn net.Conn
}

func (s *RemoteStats) Get() (Stats, error) {
	out := strings.Builder{}

	serverReader := bufio.NewReader(s.Conn)
	// Send command to server
	fmt.Fprintln(s.Conn, "stats-json")

	st := Stats{}
	// Read response
	for {
		line, err := serverReader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection closed.")
			return st, err
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == ">" {
			break // prompt received, ready for next input
		}
		fmt.Fprint(&out, line)
	}

	if err := json.Unmarshal([]byte(out.String()), &st); err != nil {
		return st, err
	}
	return st, nil
}

func (s *RemoteStats) Exec(cmd any) {
	if _, err := fmt.Fprintln(s.Conn, cmd.(string)); err != nil {
		fmt.Println("Connection closed.")
	}
}
