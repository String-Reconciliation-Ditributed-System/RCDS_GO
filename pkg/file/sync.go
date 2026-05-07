package file

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const protocolVersion = 1
const defaultChunkSize = 64 * 1024
const defaultTimeout = 30 * time.Second

type SyncOptions struct {
	ChunkSize int
	Timeout   time.Duration
}

type SyncResult struct {
	FileSize          int64
	TotalChunks       int
	TransferredChunks int
	TransferredBytes  int64
	Checksum          string
}

type chunk struct {
	Hash string `json:"hash"`
	Size int    `json:"size"`
	Data []byte `json:"data,omitempty"`
}

type fileSyncRequest struct {
	Version int      `json:"version"`
	Hashes  []string `json:"hashes"`
}

type fileSyncResponse struct {
	Version  int     `json:"version"`
	Error    string  `json:"error,omitempty"`
	FileSize int64   `json:"file_size"`
	Checksum string  `json:"checksum"`
	Manifest []chunk `json:"manifest"`
	Chunks   []chunk `json:"chunks"`
}

func ServeOnce(ctx context.Context, host string, port int, sourcePath string, options SyncOptions) (SyncResult, error) {
	options = completeOptions(options)
	sourceChunks, fileSize, checksum, err := readChunks(sourcePath, options.ChunkSize)
	if err != nil {
		return SyncResult{}, err
	}

	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", net.JoinHostPort(defaultHost(host), strconv.Itoa(port)))
	if err != nil {
		return SyncResult{}, err
	}
	defer listener.Close()

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			listener.Close()
		case <-done:
		}
	}()

	conn, err := listener.Accept()
	if err != nil {
		return SyncResult{}, err
	}
	defer conn.Close()
	if options.Timeout > 0 {
		_ = conn.SetDeadline(time.Now().Add(options.Timeout))
	}

	var req fileSyncRequest
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		return SyncResult{}, err
	}
	if req.Version != protocolVersion {
		message := fmt.Sprintf("unsupported file sync protocol version %d", req.Version)
		_ = writeFileSyncError(conn, message)
		return SyncResult{}, fmt.Errorf("%s", message)
	}

	clientHashes := make(map[string]struct{}, len(req.Hashes))
	for _, h := range req.Hashes {
		clientHashes[h] = struct{}{}
	}

	manifest := make([]chunk, 0, len(sourceChunks))
	payloads := make([]chunk, 0)
	payloadSent := make(map[string]struct{})
	var transferred int64
	for _, c := range sourceChunks {
		manifest = append(manifest, chunk{Hash: c.Hash, Size: c.Size})
		if _, ok := clientHashes[c.Hash]; ok {
			continue
		}
		if _, ok := payloadSent[c.Hash]; ok {
			continue
		}
		payloads = append(payloads, c)
		payloadSent[c.Hash] = struct{}{}
		transferred += int64(len(c.Data))
	}

	resp := fileSyncResponse{
		Version:  protocolVersion,
		FileSize: fileSize,
		Checksum: checksum,
		Manifest: manifest,
		Chunks:   payloads,
	}
	if err := json.NewEncoder(conn).Encode(resp); err != nil {
		return SyncResult{}, err
	}

	return SyncResult{
		FileSize:          fileSize,
		TotalChunks:       len(manifest),
		TransferredChunks: len(payloads),
		TransferredBytes:  transferred,
		Checksum:          checksum,
	}, nil
}

func Pull(ctx context.Context, host string, port int, destinationPath string, options SyncOptions) (SyncResult, error) {
	options = completeOptions(options)
	localChunks, _, _, err := readChunksIfExists(destinationPath, options.ChunkSize)
	if err != nil {
		return SyncResult{}, err
	}

	localByHash := make(map[string][]byte, len(localChunks))
	hashes := make([]string, 0, len(localChunks))
	for _, c := range localChunks {
		if _, ok := localByHash[c.Hash]; !ok {
			localByHash[c.Hash] = c.Data
			hashes = append(hashes, c.Hash)
		}
	}

	dialer := net.Dialer{Timeout: options.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(defaultHost(host), strconv.Itoa(port)))
	if err != nil {
		return SyncResult{}, err
	}
	defer conn.Close()
	if options.Timeout > 0 {
		_ = conn.SetDeadline(time.Now().Add(options.Timeout))
	}

	req := fileSyncRequest{Version: protocolVersion, Hashes: hashes}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return SyncResult{}, err
	}

	var resp fileSyncResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return SyncResult{}, err
	}
	if resp.Error != "" {
		return SyncResult{}, fmt.Errorf("server rejected file sync: %s", resp.Error)
	}
	if resp.Version != protocolVersion {
		return SyncResult{}, fmt.Errorf("unsupported file sync protocol version %d", resp.Version)
	}
	if err := validateResponseManifest(resp); err != nil {
		return SyncResult{}, err
	}

	var transferred int64
	for _, c := range resp.Chunks {
		data := append([]byte(nil), c.Data...)
		localByHash[c.Hash] = data
		transferred += int64(len(data))
	}

	if err := writeManifest(destinationPath, resp.Manifest, localByHash, resp.Checksum); err != nil {
		return SyncResult{}, err
	}

	return SyncResult{
		FileSize:          resp.FileSize,
		TotalChunks:       len(resp.Manifest),
		TransferredChunks: len(resp.Chunks),
		TransferredBytes:  transferred,
		Checksum:          resp.Checksum,
	}, nil
}

func completeOptions(options SyncOptions) SyncOptions {
	if options.ChunkSize <= 0 {
		options.ChunkSize = defaultChunkSize
	}
	if options.Timeout <= 0 {
		options.Timeout = defaultTimeout
	}
	return options
}

func defaultHost(host string) string {
	if host == "" {
		return "127.0.0.1"
	}
	return host
}

func readChunksIfExists(path string, chunkSize int) ([]chunk, int64, string, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			empty := sha256.Sum256(nil)
			return nil, 0, hex.EncodeToString(empty[:]), nil
		}
		return nil, 0, "", err
	}
	return readChunks(path, chunkSize)
}

func readChunks(path string, chunkSize int) ([]chunk, int64, string, error) {
	if chunkSize <= 0 {
		return nil, 0, "", fmt.Errorf("chunk size must be positive")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, 0, "", err
	}
	defer f.Close()

	capacity := 0
	if info, statErr := f.Stat(); statErr == nil && info.Size() > 0 {
		estimatedChunks := (info.Size() + int64(chunkSize) - 1) / int64(chunkSize)
		if estimatedChunks < 1_000_000 {
			capacity = int(estimatedChunks)
		}
	}

	fullHash := sha256.New()
	chunks := make([]chunk, 0, capacity)
	var size int64
	buf := make([]byte, chunkSize)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			data := append([]byte(nil), buf[:n]...)
			size += int64(n)
			if _, hashErr := fullHash.Write(data); hashErr != nil {
				return nil, 0, "", hashErr
			}
			sum := sha256.Sum256(data)
			chunks = append(chunks, chunk{
				Hash: hex.EncodeToString(sum[:]),
				Size: n,
				Data: data,
			})
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, "", err
		}
	}

	return chunks, size, hex.EncodeToString(fullHash.Sum(nil)), nil
}

func writeManifest(path string, manifest []chunk, chunks map[string][]byte, expectedChecksum string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".rcds-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	fullHash := sha256.New()
	for _, ref := range manifest {
		if ref.Size < 0 {
			_ = tmp.Close()
			return fmt.Errorf("chunk %s has negative size %d", ref.Hash, ref.Size)
		}
		data, ok := chunks[ref.Hash]
		if !ok {
			_ = tmp.Close()
			return fmt.Errorf("missing chunk %s", ref.Hash)
		}
		if len(data) != ref.Size {
			_ = tmp.Close()
			return fmt.Errorf("chunk %s size mismatch: expected %d, got %d", ref.Hash, ref.Size, len(data))
		}
		sum := sha256.Sum256(data)
		if actual := hex.EncodeToString(sum[:]); actual != ref.Hash {
			_ = tmp.Close()
			return fmt.Errorf("chunk %s hash mismatch: got %s", ref.Hash, actual)
		}
		if _, err := tmp.Write(data); err != nil {
			_ = tmp.Close()
			return err
		}
		if _, err := fullHash.Write(data); err != nil {
			_ = tmp.Close()
			return err
		}
	}

	if actual := hex.EncodeToString(fullHash.Sum(nil)); actual != expectedChecksum {
		_ = tmp.Close()
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actual)
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func validateResponseManifest(resp fileSyncResponse) error {
	var manifestSize int64
	for _, ref := range resp.Manifest {
		if ref.Size < 0 {
			return fmt.Errorf("chunk %s has negative size %d", ref.Hash, ref.Size)
		}
		manifestSize += int64(ref.Size)
	}
	if manifestSize != resp.FileSize {
		return fmt.Errorf("manifest size mismatch: response file_size=%d manifest_bytes=%d", resp.FileSize, manifestSize)
	}
	return nil
}

func writeFileSyncError(w io.Writer, message string) error {
	return json.NewEncoder(w).Encode(fileSyncResponse{
		Version: protocolVersion,
		Error:   message,
	})
}
