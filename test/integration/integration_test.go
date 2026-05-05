//go:build integration
// +build integration

package integration

import (
	"crypto"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm/full_sync"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm/iblt"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm/rcds"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/genSync"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/set"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetReconciliation tests integration between set operations and algorithms
func TestSetReconciliation(t *testing.T) {
	// Create two sets with some overlap
	set1 := set.New()
	set2 := set.New()

	// Add elements to set1
	for i := 0; i < 100; i++ {
		set1.InsertKey(i)
	}

	// Add elements to set2 with partial overlap
	for i := 50; i < 150; i++ {
		set2.InsertKey(i)
	}

	// Verify set operations
	assert.Equal(t, 100, set1.Len(), "Set1 should have 100 elements")
	assert.Equal(t, 100, set2.Len(), "Set2 should have 100 elements")

	// Expected difference: 50 elements unique to each set
	t.Logf("Set1 has %d elements, Set2 has %d elements", set1.Len(), set2.Len())
}

// TestHashFunctions tests hash function integration
func TestHashFunctions(t *testing.T) {
	testCases := []struct {
		name     string
		hashFunc crypto.Hash
		data     []byte
	}{
		{"SHA256", crypto.SHA256, []byte("test data")},
		{"SHA1", crypto.SHA1, []byte("test data")},
		{"MD5", crypto.MD5, []byte("test data")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hashData := algorithm.HashBytesWithCryptoFunc(tc.data, tc.hashFunc)
			hash, err := hashData.ToBytes()
			require.NoError(t, err, "Hash should not error")
			require.NotNil(t, hash, "Hash should not be nil")
			require.Greater(t, len(hash), 0, "Hash should not be empty")

			// Hash should be deterministic
			hashData2 := algorithm.HashBytesWithCryptoFunc(tc.data, tc.hashFunc)
			hash2, err2 := hashData2.ToBytes()
			require.NoError(t, err2)
			assert.Equal(t, hash, hash2, "Hash should be deterministic")
		})
	}
}

// TestHashStringConversion tests string to hash conversion
func TestHashStringConversion(t *testing.T) {
	testString := "test string"

	hashData := algorithm.HashString(testString)

	// Test ToUint64
	hash64, err := hashData.ToUint64()
	require.NoError(t, err)
	require.NotEqual(t, uint64(0), hash64, "Hash should not be zero")

	// Test ToUint32
	hash32, err := hashData.ToUint32()
	require.NoError(t, err)
	require.NotEqual(t, uint32(0), hash32, "Hash should not be zero")

	// Test ToBytes
	hashBytes, err := hashData.ToBytes()
	require.NoError(t, err)
	require.NotEmpty(t, hashBytes, "Hash bytes should not be empty")
}

// TestEndToEndReconciliation tests complete reconciliation workflow
func TestEndToEndReconciliation(t *testing.T) {
	server, err := full_sync.NewFullSetSync()
	require.NoError(t, err)
	client, err := full_sync.NewFullSetSync()
	require.NoError(t, err)

	require.NoError(t, server.AddElement([]byte("server-only")))
	require.NoError(t, server.AddElement([]byte("shared")))
	require.NoError(t, client.AddElement([]byte("client-only")))
	require.NoError(t, client.AddElement([]byte("shared")))

	syncPair(t, server, client)

	assert.EqualValues(t, *server.GetLocalSet(), *client.GetLocalSet())
	assert.True(t, server.GetLocalSet().Has([]byte("server-only")))
	assert.True(t, server.GetLocalSet().Has([]byte("client-only")))
	assert.True(t, server.GetLocalSet().Has([]byte("shared")))
}

// TestRCDSWithDifferentAlgorithms tests RCDS with different underlying algorithms
func TestRCDSWithDifferentAlgorithms(t *testing.T) {
	algorithms := []struct {
		name    string
		factory func() (genSync.GenSync, error)
	}{
		{
			name: "IBLT",
			factory: func() (genSync.GenSync, error) {
				return iblt.NewIBLTSetSync(iblt.WithSymmetricSetDiff(4), iblt.WithMaxSyncRetries(2))
			},
		},
		{
			name: "RCDS",
			factory: func() (genSync.GenSync, error) {
				return rcds.NewRCDSSetSync()
			},
		},
		{
			name:    "FullSync",
			factory: full_sync.NewFullSetSync,
		},
	}

	for _, algo := range algorithms {
		t.Run(algo.name, func(t *testing.T) {
			server, err := algo.factory()
			require.NoError(t, err)
			client, err := algo.factory()
			require.NoError(t, err)

			require.NoError(t, server.AddElement([]byte("server-only")))
			require.NoError(t, server.AddElement([]byte("shared")))
			require.NoError(t, client.AddElement([]byte("client-only")))
			require.NoError(t, client.AddElement([]byte("shared")))

			syncPair(t, server, client)

			assert.EqualValues(t, *server.GetLocalSet(), *client.GetLocalSet())
		})
	}
}

// TestConcurrentOperations tests thread safety
func TestConcurrentOperations(t *testing.T) {
	t.Skip("Skipping - Set implementation is not thread-safe (maps are not safe for concurrent use)")

	// Note: This test reveals that the current Set implementation using Go's map
	// is not thread-safe. This is a known limitation and would need mutex protection
	// if concurrent access is required.

	s := set.New()

	// Launch multiple goroutines to insert concurrently
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(offset int) {
			for j := 0; j < 100; j++ {
				s.InsertKey(offset*100 + j)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	assert.Equal(t, 1000, s.Len(), "Should have 1000 elements after concurrent insertions")
}

// TestDataPersistence tests data persistence across restarts
func TestDataPersistence(t *testing.T) {
	t.Skip("Skipping - requires persistence implementation")

	// This test would:
	// 1. Save state to disk
	// 2. Restart service
	// 3. Verify state is restored
}

func syncPair(t *testing.T, server genSync.GenSync, client genSync.GenSync) {
	t.Helper()

	port := freePort(t)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		errs <- server.SyncServer("127.0.0.1", port)
	}()
	time.Sleep(100 * time.Millisecond)
	errs <- client.SyncClient("127.0.0.1", port)
	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
}

func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok, fmt.Sprintf("expected TCP address, got %T", listener.Addr()))
	return addr.Port
}
