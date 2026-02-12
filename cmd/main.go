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

func runServer() {
	host := "127.0.0.1"
	port := 8080
	algorithm := "iblt"

	// Parse server options
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--host":
			if i+1 < len(args) {
				host = args[i+1]
				i++
			}
		case "--port":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &port)
				i++
			}
		case "--algorithm":
			if i+1 < len(args) {
				algorithm = args[i+1]
				i++
			}
		}
	}

	fmt.Printf("Starting RCDS server...\n")
	fmt.Printf("  Host: %s\n", host)
	fmt.Printf("  Port: %d\n", port)
	fmt.Printf("  Algorithm: %s\n", algorithm)
	fmt.Println("\nNote: Full server implementation coming soon.")
	fmt.Println("For now, please use the library directly in your Go code.")
	fmt.Println("See README.md for usage examples.")
}

func runClient() {
	host := "127.0.0.1"
	port := 8080
	algorithm := "iblt"

	// Parse client options
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--host":
			if i+1 < len(args) {
				host = args[i+1]
				i++
			}
		case "--port":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &port)
				i++
			}
		case "--algorithm":
			if i+1 < len(args) {
				algorithm = args[i+1]
				i++
			}
		}
	}

	fmt.Printf("Starting RCDS client...\n")
	fmt.Printf("  Host: %s\n", host)
	fmt.Printf("  Port: %d\n", port)
	fmt.Printf("  Algorithm: %s\n", algorithm)
	fmt.Println("\nNote: Full client implementation coming soon.")
	fmt.Println("For now, please use the library directly in your Go code.")
	fmt.Println("See README.md for usage examples.")
}
