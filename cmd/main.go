package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "server":
		runServer()
	case "client":
		runClient()
	case "version":
		printVersion()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("RCDS - Recursive Content-Dependent Shingling")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  rcds server [options]  - Start RCDS server")
	fmt.Println("  rcds client [options]  - Start RCDS client")
	fmt.Println("  rcds version           - Print version information")
	fmt.Println("  rcds help              - Print this help message")
	fmt.Println()
	fmt.Println("Server Options:")
	fmt.Println("  --host <host>          - Server host address (default: 127.0.0.1)")
	fmt.Println("  --port <port>          - Server port (default: 8080)")
	fmt.Println("  --algorithm <algo>     - Sync algorithm: rcds, iblt, full (default: iblt)")
	fmt.Println()
	fmt.Println("Client Options:")
	fmt.Println("  --host <host>          - Server host address (default: 127.0.0.1)")
	fmt.Println("  --port <port>          - Server port (default: 8080)")
	fmt.Println("  --algorithm <algo>     - Sync algorithm: rcds, iblt, full (default: iblt)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  rcds server --port 8080")
	fmt.Println("  rcds client --host 127.0.0.1 --port 8080")
}

func printVersion() {
	fmt.Println("RCDS version 0.1.0")
	fmt.Println("Go implementation of Recursive Content-Dependent Shingling")
}

// networkConfig holds network configuration parsed from command-line arguments
type networkConfig struct {
	host      string
	port      int
	algorithm string
}

// parseNetworkFlags parses common network flags (--host, --port, --algorithm) from command-line arguments
func parseNetworkFlags() (*networkConfig, error) {
	config := &networkConfig{
		host:      "127.0.0.1",
		port:      8080,
		algorithm: "iblt",
	}

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--host":
			if i+1 < len(args) {
				config.host = args[i+1]
				i++
			}
		case "--port":
			if i+1 < len(args) {
				_, err := fmt.Sscanf(args[i+1], "%d", &config.port)
				if err != nil {
					return nil, fmt.Errorf("invalid port number '%s': %v", args[i+1], err)
				}
				if config.port < 1 || config.port > 65535 {
					return nil, fmt.Errorf("port number must be between 1 and 65535, got %d", config.port)
				}
				i++
			}
		case "--algorithm":
			if i+1 < len(args) {
				config.algorithm = args[i+1]
				if config.algorithm != "rcds" && config.algorithm != "iblt" && config.algorithm != "full" {
					return nil, fmt.Errorf("invalid algorithm '%s'. Valid options: rcds, iblt, full", config.algorithm)
				}
				i++
			}
		}
	}

	return config, nil
}

func runServer() {
	config, err := parseNetworkFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting RCDS server...\n")
	fmt.Printf("  Host: %s\n", config.host)
	fmt.Printf("  Port: %d\n", config.port)
	fmt.Printf("  Algorithm: %s\n", config.algorithm)
	fmt.Println("\nNote: Full server implementation coming soon.")
	fmt.Println("For now, please use the library directly in your Go code.")
	fmt.Println("See README.md for usage examples.")
}

func runClient() {
	config, err := parseNetworkFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting RCDS client...\n")
	fmt.Printf("  Host: %s\n", config.host)
	fmt.Printf("  Port: %d\n", config.port)
	fmt.Printf("  Algorithm: %s\n", config.algorithm)
	fmt.Println("\nNote: Full client implementation coming soon.")
	fmt.Println("For now, please use the library directly in your Go code.")
	fmt.Println("See README.md for usage examples.")
}
