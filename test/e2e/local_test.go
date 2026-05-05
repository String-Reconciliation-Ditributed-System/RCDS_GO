//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLocalDeployment(t *testing.T) {
	binary := buildBinary(t)

	info, err := os.Stat(binary)
	require.NoError(t, err)
	require.False(t, info.IsDir())
	require.Greater(t, info.Size(), int64(0))
}

func TestServerStartStop(t *testing.T) {
	binary := buildBinary(t)
	port := freePort(t)

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, binary, "server", "--algorithm", "full", "--port", fmt.Sprint(port), "--items", "alpha")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	require.NoError(t, cmd.Start())

	time.Sleep(250 * time.Millisecond)
	cancel()
	_ = cmd.Wait()
}

func TestHealthCheck(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "version")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	require.Contains(t, string(output), "RCDS version")
}

func TestDataSynchronization(t *testing.T) {
	binary := buildBinary(t)
	port := freePort(t)
	tmp := t.TempDir()
	serverOut := filepath.Join(tmp, "server.out")
	clientOut := filepath.Join(tmp, "client.out")

	server := startServer(t, binary, port,
		"--algorithm", "full",
		"--items", "server-only,shared",
		"--output", serverOut,
	)

	client := exec.Command(binary,
		"client",
		"--algorithm", "full",
		"--port", fmt.Sprint(port),
		"--items", "client-only,shared",
		"--output", clientOut,
	)
	output, err := client.CombinedOutput()
	require.NoError(t, err, string(output))
	require.NoError(t, server.Wait(), serverOutput(server))

	expected := "client-only\nserver-only\nshared\n"
	assertFile(t, serverOut, expected)
	assertFile(t, clientOut, expected)
}

func TestFileSynchronization(t *testing.T) {
	binary := buildBinary(t)
	port := freePort(t)
	tmp := t.TempDir()
	source := filepath.Join(tmp, "source.bin")
	destination := filepath.Join(tmp, "destination.bin")

	content := []byte("aaaabbbbccccddddeeeeffff")
	require.NoError(t, os.WriteFile(source, content, 0644))
	require.NoError(t, os.WriteFile(destination, []byte("aaaabbbb"), 0644))

	server := startServer(t, binary, port,
		"--mode", "file",
		"--file", source,
		"--chunk-size", "4",
	)

	client := exec.Command(binary,
		"client",
		"--mode", "file",
		"--port", fmt.Sprint(port),
		"--file", destination,
		"--chunk-size", "4",
	)
	output, err := client.CombinedOutput()
	require.NoError(t, err, string(output))
	require.NoError(t, server.Wait(), serverOutput(server))

	data, err := os.ReadFile(destination)
	require.NoError(t, err)
	require.Equal(t, content, data)
}

func TestLargeDataset(t *testing.T) {
	binary := buildBinary(t)
	port := freePort(t)
	tmp := t.TempDir()
	serverInput := filepath.Join(tmp, "server.in")
	clientInput := filepath.Join(tmp, "client.in")
	serverOut := filepath.Join(tmp, "server.out")
	clientOut := filepath.Join(tmp, "client.out")

	var serverLines, clientLines []string
	for i := 0; i < 250; i++ {
		line := fmt.Sprintf("shared-%03d", i)
		serverLines = append(serverLines, line)
		clientLines = append(clientLines, line)
	}
	for i := 0; i < 50; i++ {
		serverLines = append(serverLines, fmt.Sprintf("server-%03d", i))
		clientLines = append(clientLines, fmt.Sprintf("client-%03d", i))
	}
	require.NoError(t, os.WriteFile(serverInput, []byte(strings.Join(serverLines, "\n")+"\n"), 0644))
	require.NoError(t, os.WriteFile(clientInput, []byte(strings.Join(clientLines, "\n")+"\n"), 0644))

	server := startServer(t, binary, port,
		"--algorithm", "full",
		"--input", serverInput,
		"--output", serverOut,
	)

	client := exec.Command(binary,
		"client",
		"--algorithm", "full",
		"--port", fmt.Sprint(port),
		"--input", clientInput,
		"--output", clientOut,
	)
	output, err := client.CombinedOutput()
	require.NoError(t, err, string(output))
	require.NoError(t, server.Wait(), serverOutput(server))

	serverData, err := os.ReadFile(serverOut)
	require.NoError(t, err)
	clientData, err := os.ReadFile(clientOut)
	require.NoError(t, err)
	require.Equal(t, string(serverData), string(clientData))
	require.Contains(t, string(clientData), "server-049")
	require.Contains(t, string(clientData), "client-049")
}

func TestRecoveryFromFailure(t *testing.T) {
	binary := buildBinary(t)
	port := freePort(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client := exec.CommandContext(ctx, binary, "client", "--algorithm", "full", "--port", fmt.Sprint(port), "--items", "orphan")
	output, err := client.CombinedOutput()
	require.Error(t, err)
	require.Contains(t, string(output), "Error:")
}

func buildBinary(t *testing.T) string {
	t.Helper()

	root, err := filepath.Abs("../..")
	require.NoError(t, err)

	binary := filepath.Join(t.TempDir(), "rcds")
	cmd := exec.Command("go", "build", "-o", binary, "./cmd")
	cmd.Dir = root
	cmd.Env = os.Environ()
	if os.Getenv("GOCACHE") == "" {
		cmd.Env = append(cmd.Env, "GOCACHE="+filepath.Join(t.TempDir(), "go-build"))
	}
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	return binary
}

func startServer(t *testing.T, binary string, port int, args ...string) *exec.Cmd {
	t.Helper()

	cmdArgs := append([]string{"server", "--port", fmt.Sprint(port)}, args...)
	cmd := exec.Command(binary, cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	})
	time.Sleep(250 * time.Millisecond)
	return cmd
}

func serverOutput(cmd *exec.Cmd) string {
	var parts []string
	if stdout, ok := cmd.Stdout.(*bytes.Buffer); ok {
		parts = append(parts, stdout.String())
	}
	if stderr, ok := cmd.Stderr.(*bytes.Buffer); ok {
		parts = append(parts, stderr.String())
	}
	return strings.Join(parts, "\n")
}

func assertFile(t *testing.T, path, expected string) {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, expected, string(data))
}

func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}
