package monitor

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/goccy/go-json"
	arkstats "github.com/mlange-42/ark/ecs/stats"
)

// Connection interface.
type Connection interface {
	Get() (Stats, error)
	Exec(string) error
}

// Stats for the monitor.
type Stats struct {
	Ticks int
	Stats *arkstats.World
}

// RemoteConnection implements Connection.
type RemoteConnection struct {
	Conn net.Conn
}

// Get stats.
func (s *RemoteConnection) Get() (Stats, error) {
	out := strings.Builder{}

	serverReader := bufio.NewReader(s.Conn)
	st := Stats{}

	// Send command to server
	if _, err := fmt.Fprintln(s.Conn, "stats-json"); err != nil {
		return st, err
	}

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

// Exec a command.
func (s *RemoteConnection) Exec(cmd string) error {
	serverReader := bufio.NewReader(s.Conn)
	if _, err := fmt.Fprintln(s.Conn, cmd); err != nil {
		return err
	}

	// Read response
	for {
		line, err := serverReader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection closed.")
			return err
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == ">" {
			break // prompt received, ready for next input
		}
	}
	return nil
}
