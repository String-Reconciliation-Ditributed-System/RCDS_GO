package file

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"os"
	"path/filepath"
	"strings"
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

func TestWriteManifestRejectsSizeMismatch(t *testing.T) {
	hash := sha256.Sum256([]byte("abc"))
	err := writeManifest(
		filepath.Join(t.TempDir(), "copy.txt"),
		[]chunk{{Hash: hex.EncodeToString(hash[:]), Size: 4}},
		map[string][]byte{hex.EncodeToString(hash[:]): []byte("abc")},
		hex.EncodeToString(hash[:]),
	)
	if err == nil || !strings.Contains(err.Error(), "size mismatch") {
		t.Fatalf("expected size mismatch error, got %v", err)
	}
}

func TestWriteManifestRejectsNegativeChunkSize(t *testing.T) {
	hash := sha256.Sum256([]byte("abc"))
	err := writeManifest(
		filepath.Join(t.TempDir(), "copy.txt"),
		[]chunk{{Hash: hex.EncodeToString(hash[:]), Size: -1}},
		map[string][]byte{hex.EncodeToString(hash[:]): []byte("abc")},
		hex.EncodeToString(hash[:]),
	)
	if err == nil || !strings.Contains(err.Error(), "negative size") {
		t.Fatalf("expected negative size error, got %v", err)
	}
}

func TestWriteManifestRejectsChunkHashMismatch(t *testing.T) {
	actualHash := sha256.Sum256([]byte("abc"))
	declaredHash := sha256.Sum256([]byte("xyz"))
	err := writeManifest(
		filepath.Join(t.TempDir(), "copy.txt"),
		[]chunk{{Hash: hex.EncodeToString(declaredHash[:]), Size: 3}},
		map[string][]byte{hex.EncodeToString(declaredHash[:]): []byte("abc")},
		hex.EncodeToString(actualHash[:]),
	)
	if err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("expected hash mismatch error, got %v", err)
	}
}

func TestValidateResponseManifestRejectsFileSizeMismatch(t *testing.T) {
	err := validateResponseManifest(fileSyncResponse{
		FileSize: 4,
		Manifest: []chunk{
			{Hash: "a", Size: 2},
			{Hash: "b", Size: 1},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "manifest size mismatch") {
		t.Fatalf("expected manifest size mismatch, got %v", err)
	}
}

func TestValidateResponseManifestRejectsNegativeChunkSize(t *testing.T) {
	err := validateResponseManifest(fileSyncResponse{
		FileSize: 0,
		Manifest: []chunk{
			{Hash: "a", Size: -1},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "negative size") {
		t.Fatalf("expected negative size error, got %v", err)
	}
}

func TestReadChunksIfMissingUsesEmptyChecksum(t *testing.T) {
	_, size, checksum, err := readChunksIfExists(filepath.Join(t.TempDir(), "missing.bin"), 4)
	if err != nil {
		t.Fatal(err)
	}
	emptyHash := sha256.Sum256(nil)
	if size != 0 || checksum != hex.EncodeToString(emptyHash[:]) {
		t.Fatalf("unexpected missing file result: size=%d checksum=%s", size, checksum)
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
		if strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("localhost TCP bind is not permitted in this environment: %v", err)
		}
		t.Fatal(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func BenchmarkReadChunks1MiB(b *testing.B) {
	source := filepath.Join(b.TempDir(), "source.bin")
	content := make([]byte, 1<<20)
	for i := range content {
		content[i] = byte(i % 251)
	}
	if err := os.WriteFile(source, content, 0644); err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(content)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chunks, size, checksum, err := readChunks(source, defaultChunkSize)
		if err != nil {
			b.Fatal(err)
		}
		if size != int64(len(content)) || checksum == "" || len(chunks) != 16 {
			b.Fatalf("unexpected read result: size=%d checksum=%q chunks=%d", size, checksum, len(chunks))
		}
	}
}
