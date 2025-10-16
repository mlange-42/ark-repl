package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/mlange-42/ark-repl/internal/monitor"
)

// CLI arguments.
type CLI struct {
	Address string   `arg:"" help:"Server address to connect to ('host:port' or just ':port'). Default: localhost:9000" default:"localhost:9000"`
	Run     []string `help:"REPL commands to run on startup." short:"r" name:"run" placeholder:"COMMAND"`
}

func main() {
	var cli CLI
	kong.Parse(&cli)
	addr := normalizeAddress(cli.Address)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			panic(err)
		}
	}()

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

		var input string
		if len(cli.Run) > 0 {
			input = cli.Run[0]
			fmt.Println(input)
			cli.Run = cli.Run[1:]
		} else {
			if !clientReader.Scan() {
				break
			}
			input = clientReader.Text()
		}

		if input == "monitor" {
			_ = monitor.New(&monitor.RemoteConnection{Conn: conn})
			continue
		}

		if input == "$" {
			var blockLines []string
			blockLines = append(blockLines, "$") // include opening delimiter

			for {
				if !clientReader.Scan() {
					fmt.Println("Unexpected end of input during block.")
					return
				}
				blockLine := clientReader.Text()
				blockLines = append(blockLines, blockLine)

				if strings.TrimSpace(blockLine) == "$" {
					break // closing delimiter found
				}
			}

			// Send full block (including $...$) to server
			fullBlock := strings.Join(blockLines, "\n")
			fmt.Fprintln(conn, fullBlock)
		} else {
			// Send command to server
			if _, err := fmt.Fprintln(conn, input); err != nil {
				panic(err)
			}
		}

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

func normalizeAddress(input string) string {
	if strings.HasPrefix(input, ":") {
		return "localhost" + input
	}
	return input
}
