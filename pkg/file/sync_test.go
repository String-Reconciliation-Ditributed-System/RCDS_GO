package file

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadChunksAndWriteManifest(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.txt")
	content := []byte("abcdefghijklmnopqrstuvwxyz")
	if err := os.WriteFile(source, content, 0644); err != nil {
		t.Fatal(err)
	}

	chunks, size, checksum, err := readChunks(source, 5)
	if err != nil {
		t.Fatal(err)
	}
	if size != int64(len(content)) {
		t.Fatalf("expected size %d, got %d", len(content), size)
	}
	if len(chunks) != 6 {
		t.Fatalf("expected 6 chunks, got %d", len(chunks))
	}

	manifest := make([]chunk, 0, len(chunks))
	byHash := make(map[string][]byte, len(chunks))
	for _, c := range chunks {
		manifest = append(manifest, chunk{Hash: c.Hash, Size: c.Size})
		byHash[c.Hash] = c.Data
	}

	destination := filepath.Join(t.TempDir(), "nested", "copy.txt")
	if err := writeManifest(destination, manifest, byHash, checksum); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(content) {
		t.Fatalf("expected %q, got %q", string(content), string(data))
	}
}

func TestWriteManifestRejectsMissingChunk(t *testing.T) {
	err := writeManifest(
		filepath.Join(t.TempDir(), "copy.txt"),
		[]chunk{{Hash: "missing", Size: 3}},
		map[string][]byte{},
		"checksum",
	)
	if err == nil {
		t.Fatal("expected missing chunk error")
	}
}

func TestServeOnceAndPull(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "source.bin")
	destination := filepath.Join(tmp, "copy.bin")
	content := []byte("aaaabbbbccccddddeeeeffff")
	if err := os.WriteFile(source, content, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, []byte("aaaabbbb"), 0644); err != nil {
		t.Fatal(err)
	}

	port := freePort(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	serverResult := make(chan SyncResult, 1)
	serverErr := make(chan error, 1)
	go func() {
		result, err := ServeOnce(ctx, "127.0.0.1", port, source, SyncOptions{ChunkSize: 4, Timeout: 5 * time.Second})
		serverResult <- result
		serverErr <- err
	}()
	time.Sleep(50 * time.Millisecond)

	clientResult, err := Pull(ctx, "127.0.0.1", port, destination, SyncOptions{ChunkSize: 4, Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if err := <-serverErr; err != nil {
		t.Fatal(err)
	}
	server := <-serverResult

	data, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(content) {
		t.Fatalf("expected %q, got %q", string(content), string(data))
	}
	if clientResult.Checksum != server.Checksum {
		t.Fatalf("checksum mismatch: client %s server %s", clientResult.Checksum, server.Checksum)
	}
	if clientResult.TransferredChunks == 0 || clientResult.TransferredChunks >= clientResult.TotalChunks {
		t.Fatalf("expected partial transfer, got %+v", clientResult)
	}
}

func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}
