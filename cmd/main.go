package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	filesync "github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/file"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm/full_sync"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm/iblt"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm/rcds"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/genSync"
)

const version = "0.2.0"

type cliConfig struct {
	host         string
	port         int
	algorithm    string
	mode         string
	items        string
	input        string
	output       string
	file         string
	expectedDiff int
	maxRetries   int
	once         bool
	freezeLocal  bool
	timeout      time.Duration
	chunkSize    int
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 1
	}

	switch args[0] {
	case "server":
		return runServer(args[1:], stdout, stderr)
	case "client":
		return runClient(args[1:], stdout, stderr)
	case "version":
		printVersion(stdout)
		return 0
	case "help", "--help", "-h":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "Unknown command: %s\n\n", args[0])
		printUsage(stderr)
		return 1
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "RCDS - Recursive Content-Dependent Shingling")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  rcds server [options]  Start a set or file sync server")
	fmt.Fprintln(w, "  rcds client [options]  Connect to a set or file sync server")
	fmt.Fprintln(w, "  rcds version           Print version information")
	fmt.Fprintln(w, "  rcds help              Print this help message")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Common Options:")
	fmt.Fprintln(w, "  --host <host>              Host address (default: 127.0.0.1)")
	fmt.Fprintln(w, "  --port <port>              TCP port (default: 8080)")
	fmt.Fprintln(w, "  --mode <set|file>          Sync mode (default: set)")
	fmt.Fprintln(w, "  --algorithm <algo>         Set algorithm: rcds, iblt, full (default: iblt)")
	fmt.Fprintln(w, "  --items <a,b,c>            Comma-separated set elements")
	fmt.Fprintln(w, "  --input <path>             Line-delimited set input, or file source for --mode file server")
	fmt.Fprintln(w, "  --output <path>            Write reconciled set, or file destination for --mode file client")
	fmt.Fprintln(w, "  --file <path>              File source/destination for --mode file")
	fmt.Fprintln(w, "  --expected-diff <n>        Expected symmetric difference for IBLT (default: 100)")
	fmt.Fprintln(w, "  --max-retries <n>          IBLT retry count (default: 3)")
	fmt.Fprintln(w, "  --timeout <duration>       File sync network timeout (default: 30s)")
	fmt.Fprintln(w, "  --chunk-size <bytes>       File sync chunk size (default: 65536)")
	fmt.Fprintln(w, "  --freeze-local             Do not apply remote set additions locally")
	fmt.Fprintln(w, "  --once                     Server handles one client then exits (default: true)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  rcds server --items server-only,shared --output server.out")
	fmt.Fprintln(w, "  rcds client --items client-only,shared --output client.out")
	fmt.Fprintln(w, "  rcds server --mode file --file ./source.bin")
	fmt.Fprintln(w, "  rcds client --mode file --file ./copy.bin")
}

func printVersion(w io.Writer) {
	fmt.Fprintf(w, "RCDS version %s\n", version)
	fmt.Fprintln(w, "Go implementation of Recursive Content-Dependent Shingling")
}

func parseCommandFlags(command string, args []string, stderr io.Writer) (*cliConfig, error) {
	config := &cliConfig{
		host:         "127.0.0.1",
		port:         8080,
		algorithm:    "iblt",
		mode:         "set",
		expectedDiff: 100,
		maxRetries:   3,
		once:         true,
		timeout:      30 * time.Second,
		chunkSize:    64 * 1024,
	}

	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&config.host, "host", config.host, "host address")
	fs.IntVar(&config.port, "port", config.port, "TCP port")
	fs.StringVar(&config.algorithm, "algorithm", config.algorithm, "set sync algorithm: rcds, iblt, full")
	fs.StringVar(&config.mode, "mode", config.mode, "sync mode: set or file")
	fs.StringVar(&config.items, "items", "", "comma-separated set elements")
	fs.StringVar(&config.input, "input", "", "line-delimited set input, or file source for file server")
	fs.StringVar(&config.output, "output", "", "write reconciled set, or file destination for file client")
	fs.StringVar(&config.file, "file", "", "file source or destination for file mode")
	fs.IntVar(&config.expectedDiff, "expected-diff", config.expectedDiff, "expected symmetric difference for IBLT")
	fs.IntVar(&config.maxRetries, "max-retries", config.maxRetries, "IBLT max retry count")
	fs.BoolVar(&config.once, "once", config.once, "server handles one client then exits")
	fs.BoolVar(&config.freezeLocal, "freeze-local", false, "do not apply remote set additions locally")
	fs.DurationVar(&config.timeout, "timeout", config.timeout, "file sync network timeout")
	fs.IntVar(&config.chunkSize, "chunk-size", config.chunkSize, "file sync chunk size")
	fs.Usage = func() {
		printUsage(stderr)
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if fs.NArg() > 0 {
		return nil, fmt.Errorf("unexpected positional arguments: %s", strings.Join(fs.Args(), " "))
	}
	if config.port < 1 || config.port > 65535 {
		return nil, fmt.Errorf("port number must be between 1 and 65535, got %d", config.port)
	}
	if config.mode != "set" && config.mode != "file" {
		return nil, fmt.Errorf("invalid mode %q. Valid options: set, file", config.mode)
	}
	if config.algorithm != "rcds" && config.algorithm != "iblt" && config.algorithm != "full" {
		return nil, fmt.Errorf("invalid algorithm %q. Valid options: rcds, iblt, full", config.algorithm)
	}
	if config.expectedDiff <= 0 {
		return nil, fmt.Errorf("expected-diff must be positive")
	}
	if config.maxRetries < 0 {
		return nil, fmt.Errorf("max-retries must be non-negative")
	}
	if config.chunkSize <= 0 {
		return nil, fmt.Errorf("chunk-size must be positive")
	}
	return config, nil
}

func runServer(args []string, stdout, stderr io.Writer) int {
	config, err := parseCommandFlags("server", args, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if config.mode == "file" {
		return runFileServer(ctx, config, stdout, stderr)
	}
	return runSetServer(ctx, config, stdout, stderr)
}

func runClient(args []string, stdout, stderr io.Writer) int {
	config, err := parseCommandFlags("client", args, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if config.mode == "file" {
		return runFileClient(ctx, config, stdout, stderr)
	}
	return runSetClient(config, stdout, stderr)
}

func runSetServer(ctx context.Context, config *cliConfig, stdout, stderr io.Writer) int {
	syncer, err := newSetSyncer(config)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	if err := addConfiguredElements(syncer, config); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	for {
		fmt.Fprintf(stdout, "Listening for set sync on %s:%d using %s\n", config.host, config.port, config.algorithm)
		if err := syncer.SyncServer(config.host, config.port); err != nil {
			if ctx.Err() != nil {
				return 0
			}
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
		if err := writeSetOutput(syncer, config.output); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
		printSetResult(stdout, syncer)
		if config.once {
			return 0
		}
	}
}

func runSetClient(config *cliConfig, stdout, stderr io.Writer) int {
	syncer, err := newSetSyncer(config)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	if err := addConfiguredElements(syncer, config); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if err := syncer.SyncClient(config.host, config.port); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	if err := writeSetOutput(syncer, config.output); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	printSetResult(stdout, syncer)
	return 0
}

func runFileServer(ctx context.Context, config *cliConfig, stdout, stderr io.Writer) int {
	sourcePath := firstNonEmpty(config.file, config.input)
	if sourcePath == "" {
		fmt.Fprintln(stderr, "Error: file mode server requires --file or --input")
		return 1
	}

	for {
		fmt.Fprintf(stdout, "Listening for file sync on %s:%d from %s\n", config.host, config.port, sourcePath)
		result, err := filesync.ServeOnce(ctx, config.host, config.port, sourcePath, fileOptions(config))
		if err != nil {
			if ctx.Err() != nil {
				return 0
			}
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
		printFileResult(stdout, result)
		if config.once {
			return 0
		}
	}
}

func runFileClient(ctx context.Context, config *cliConfig, stdout, stderr io.Writer) int {
	destinationPath := firstNonEmpty(config.file, config.output)
	if destinationPath == "" {
		fmt.Fprintln(stderr, "Error: file mode client requires --file or --output")
		return 1
	}

	result, err := filesync.Pull(ctx, config.host, config.port, destinationPath, fileOptions(config))
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	printFileResult(stdout, result)
	return 0
}

func newSetSyncer(config *cliConfig) (genSync.GenSync, error) {
	switch config.algorithm {
	case "full":
		return full_sync.NewFullSetSync()
	case "iblt":
		return iblt.NewIBLTSetSync(
			iblt.WithSymmetricSetDiff(config.expectedDiff),
			iblt.WithMaxSyncRetries(config.maxRetries),
		)
	case "rcds":
		return rcds.NewRCDSSetSync()
	default:
		return nil, fmt.Errorf("invalid algorithm %q", config.algorithm)
	}
}

func addConfiguredElements(syncer genSync.GenSync, config *cliConfig) error {
	syncer.SetFreezeLocal(config.freezeLocal)
	elements, err := configuredElements(config)
	if err != nil {
		return err
	}
	for _, elem := range elements {
		if err := syncer.AddElement([]byte(elem)); err != nil {
			return err
		}
	}
	return nil
}

func configuredElements(config *cliConfig) ([]string, error) {
	var elements []string
	for _, elem := range strings.Split(config.items, ",") {
		elem = strings.TrimSpace(elem)
		if elem != "" {
			elements = append(elements, elem)
		}
	}
	if config.input == "" || config.mode == "file" {
		return elements, nil
	}

	data, err := os.ReadFile(config.input)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if line != "" {
			elements = append(elements, line)
		}
	}
	return elements, nil
}

func writeSetOutput(syncer genSync.GenSync, outputPath string) error {
	if outputPath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	elements := sortedSetElements(syncer)
	content := strings.Join(elements, "\n")
	if len(elements) > 0 {
		content += "\n"
	}
	return os.WriteFile(outputPath, []byte(content), 0644)
}

func sortedSetElements(syncer genSync.GenSync) []string {
	localSet := syncer.GetLocalSet()
	elements := make([]string, 0, localSet.Len())
	for elem := range *localSet {
		elements = append(elements, fmt.Sprint(elem))
	}
	sort.Strings(elements)
	return elements
}

func fileOptions(config *cliConfig) filesync.SyncOptions {
	return filesync.SyncOptions{
		ChunkSize: config.chunkSize,
		Timeout:   config.timeout,
	}
}

func printSetResult(w io.Writer, syncer genSync.GenSync) {
	fmt.Fprintf(w, "Set sync complete: elements=%d additions=%d sent_bytes=%d received_bytes=%d total_bytes=%d\n",
		syncer.GetLocalSet().Len(),
		syncer.GetSetAdditions().Len(),
		syncer.GetSentBytes(),
		syncer.GetReceivedBytes(),
		syncer.GetTotalBytes(),
	)
}

func printFileResult(w io.Writer, result filesync.SyncResult) {
	fmt.Fprintf(w, "File sync complete: size=%d chunks=%d transferred_chunks=%d transferred_bytes=%d checksum=%s\n",
		result.FileSize,
		result.TotalChunks,
		result.TransferredChunks,
		result.TransferredBytes,
		result.Checksum,
	)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
