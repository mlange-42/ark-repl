package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/mlange-42/ark-repl/internal/monitor"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to Ark REPL.")
	serverReader := bufio.NewReader(conn)
	clientReader := bufio.NewScanner(os.Stdin)

	// Read initial greeting and first prompt
	for {
		line, err := serverReader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection closed.")
			return
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == ">" {
			break // first prompt received, stop greeting phase
		}
		fmt.Print(line)
	}

	for {
		// Show local prompt
		fmt.Print("> ")
		if !clientReader.Scan() {
			break
		}
		input := clientReader.Text()

		if input == "monitor" {
			_ = monitor.New(&monitor.RemoteConnection{Conn: conn})
			continue
		}

		// Send command to server
		fmt.Fprintln(conn, input)

		// Read response until next prompt
		for {
			line, err := serverReader.ReadString('\n')
			if err != nil {
				fmt.Println("Connection closed.")
				return
			}
			trimmed := strings.TrimSpace(line)
			if trimmed == ">" {
				break // prompt received, ready for next input
			}
			fmt.Print(line)
		}
	}
}
