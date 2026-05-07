package iblt

import (
	"crypto"
	"encoding/binary"
	mathrand "math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/set"
)

var rand = newTestRand()

type testRand struct {
	source *mathrand.Rand
}

func newTestRand() *testRand {
	return &testRand{source: mathrand.New(mathrand.NewSource(1))}
}

func (r *testRand) Seed(seed int64) {
	r.source.Seed(seed)
}

func (r *testRand) IntnRange(min, max int) int {
	if max <= min {
		return min
	}
	return min + r.source.Intn(max-min)
}

func (r *testRand) String(length int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	buf := make([]byte, length)
	for i := range buf {
		buf[i] = alphabet[r.source.Intn(len(alphabet))]
	}
	return string(buf)
}

func TestIBLTUsesFullChecksumInSerializedTables(t *testing.T) {
	syncer, err := NewIBLTSetSync(WithSymmetricSetDiff(1), WithDataLen(4))
	require.NoError(t, err)

	require.NoError(t, syncer.AddElement([]byte("abcd")))

	ibltSyncer, ok := syncer.(*ibltSync)
	require.True(t, ok)

	serialized, err := ibltSyncer.Table.Serialize()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(serialized), 8)

	assert.Equal(t, uint16(ibltChecksumBytes), binary.BigEndian.Uint16(serialized[4:6]))
}

func TestWithDataLen(t *testing.T) {
	rand.Seed(100)
	tests := []struct {
		serverSetSize    int
		clientSetSize    int
		intersectionSize int
		dataLen          int
	}{
		{
			serverSetSize:    5,
			intersectionSize: 4,
			clientSetSize:    5,
			dataLen:          200,
		},
		{
			serverSetSize:    400,
			clientSetSize:    400,
			intersectionSize: 350,
			dataLen:          300,
		},
		{
			serverSetSize:    5000,
			clientSetSize:    4000,
			intersectionSize: 3001,
			dataLen:          20,
		},
	}
	for _, tt := range tests {
		t.Logf("New Pair test with %+v", tt)
		diffNum := tt.serverSetSize - tt.intersectionSize
		diffNum += tt.clientSetSize - tt.intersectionSize

		server, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum), WithDataLen(tt.dataLen), WithMaxSyncRetries(2))
		require.NoError(t, err)

		client, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum), WithDataLen(tt.dataLen), WithMaxSyncRetries(2))
		require.NoError(t, err)

		expectedSet := set.New()
		expectedClientExtra := set.New()
		expectedServerExtra := set.New()
		for i := 0; i < tt.intersectionSize; i++ {
			td := []byte(rand.String(tt.dataLen))
			err = server.AddElement(td)
			require.NoError(t, err)
			err = client.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		for i := 0; i < tt.clientSetSize-tt.intersectionSize; i++ {
			td := []byte(rand.String(tt.dataLen))
			err = client.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
			expectedClientExtra.InsertKey(td)
		}

		for i := 0; i < tt.serverSetSize-tt.intersectionSize; i++ {
			td := []byte(rand.String(tt.dataLen))
			err = server.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
			expectedServerExtra.InsertKey(td)
		}

		port := freePort(t)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			err := client.SyncServer("127.0.0.1", port)
			assert.NoError(t, err)
			wg.Done()
		}()
		err = server.SyncClient("127.0.0.1", port)
		assert.NoError(t, err)
		wg.Wait()

		assert.Len(t, *client.GetSetAdditions(), tt.serverSetSize-tt.intersectionSize)
		assert.Len(t, *server.GetSetAdditions(), tt.clientSetSize-tt.intersectionSize)

		assert.EqualValues(t, *expectedClientExtra, *server.GetSetAdditions())
		assert.EqualValues(t, *expectedServerExtra, *client.GetSetAdditions())

		assert.EqualValues(t, *server.GetLocalSet(), *client.GetLocalSet())
		assert.Equal(t, server.GetTotalBytes(), client.GetTotalBytes())
	}
}

func TestIBLTSyncRejectsNonByteElements(t *testing.T) {
	syncer, err := NewIBLTSetSync(WithSymmetricSetDiff(1))
	require.NoError(t, err)

	assert.Error(t, syncer.AddElement("not bytes"))
	assert.Error(t, syncer.DeleteElement("not bytes"))
}

func TestNewIBLTSetSyncRejectsInvalidHashFunc(t *testing.T) {
	assert.NotPanics(t, func() {
		_, err := NewIBLTSetSync(WithSymmetricSetDiff(1), WithHashFunc(crypto.Hash(0)))
		assert.Error(t, err)
	})
}

func TestWithHashFunc(t *testing.T) {
	rand.Seed(10)
	tests := []struct {
		serverSetSize    int
		clientSetSize    int
		intersectionSize int
		hashFunc         crypto.Hash
	}{
		{
			serverSetSize:    5,
			intersectionSize: 4,
			clientSetSize:    5,
			hashFunc:         crypto.SHA512,
		},
		{
			serverSetSize:    400,
			clientSetSize:    400,
			intersectionSize: 350,
			hashFunc:         crypto.SHA256,
		},
		{
			serverSetSize:    5000,
			clientSetSize:    4000,
			intersectionSize: 3001,
			hashFunc:         crypto.SHA1,
		},
	}
	for _, tt := range tests {
		t.Logf("New Pair test with %+v", tt)
		diffNum := tt.serverSetSize - tt.intersectionSize
		diffNum += tt.clientSetSize - tt.intersectionSize

		server, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum), WithHashFunc(tt.hashFunc))
		require.NoError(t, err)

		client, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum), WithHashFunc(tt.hashFunc))
		require.NoError(t, err)

		expectedSet := set.New()
		for i := 0; i < tt.intersectionSize; i++ {
			td := []byte(rand.String(rand.IntnRange(1, 1000)))
			err = server.AddElement(td)
			require.NoError(t, err)
			err = client.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		for i := 0; i < tt.clientSetSize-tt.intersectionSize; i++ {
			td := []byte(rand.String(rand.IntnRange(1, 1000)))
			err = client.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		for i := 0; i < tt.serverSetSize-tt.intersectionSize; i++ {
			td := []byte(rand.String(rand.IntnRange(1, 1000)))
			err = server.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		port := freePort(t)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			err := client.SyncServer("127.0.0.1", port)
			assert.NoError(t, err)
			wg.Done()
		}()
		err = server.SyncClient("127.0.0.1", port)
		assert.NoError(t, err)
		wg.Wait()

		assert.EqualValues(t, *server.GetLocalSet(), *client.GetLocalSet())
		assert.Equal(t, server.GetTotalBytes(), client.GetTotalBytes())
	}
}

func TestNewIBLTSetSyncWithDifferentDestinations(t *testing.T) {
	rand.Seed(100)

	tests := []struct {
		serverSetSize    int
		clientSetSize    int
		intersectionSize int
	}{
		{
			serverSetSize:    5,
			intersectionSize: 4,
			clientSetSize:    5,
		},
		{
			serverSetSize:    400,
			clientSetSize:    400,
			intersectionSize: 350,
		},
		{
			serverSetSize:    5000,
			clientSetSize:    4000,
			intersectionSize: 3001,
		},
	}
	for _, tt := range tests {
		t.Logf("New Pair test with %+v", tt)
		diffNum := tt.serverSetSize - tt.intersectionSize
		diffNum += tt.clientSetSize - tt.intersectionSize

		server, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum))
		require.NoError(t, err)

		client1, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum))
		require.NoError(t, err)

		client2, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum))
		require.NoError(t, err)

		expectedSet := set.New()
		for i := 0; i < tt.intersectionSize; i++ {
			td := []byte(rand.String(rand.IntnRange(1, 1000)))
			err = server.AddElement(td)
			require.NoError(t, err)
			err = client1.AddElement(td)
			require.NoError(t, err)
			err = client2.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		for i := 0; i < tt.clientSetSize-tt.intersectionSize; i++ {
			td := []byte(rand.String(rand.IntnRange(1, 1000)))
			err = client1.AddElement(td)
			require.NoError(t, err)
			err = client2.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		for i := 0; i < tt.serverSetSize-tt.intersectionSize; i++ {
			td := []byte(rand.String(rand.IntnRange(1, 1000)))
			err = server.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		var wg sync.WaitGroup
		t.Log("syncing with client 1 in the first address")
		port1 := freePort(t)
		wg.Add(1)
		go func() {
			err := client1.SyncServer("127.0.0.1", port1)
			assert.NoError(t, err)
			wg.Done()
		}()
		wg.Add(1)
		go func() {
			err = server.SyncClient("127.0.0.1", port1)
			assert.NoError(t, err)
			wg.Done()
		}()
		wg.Wait()

		t.Log("syncing with client 2 in the second address")
		port2 := freePort(t)
		wg.Add(1)
		go func() {
			err := client2.SyncServer("127.0.0.1", port2)
			assert.NoError(t, err)
			wg.Done()
		}()
		wg.Add(1)
		go func() {
			err = server.SyncClient("127.0.0.1", port2)
			assert.NoError(t, err)
			wg.Done()
		}()
		wg.Wait()

		assert.EqualValues(t, *server.GetLocalSet(), *client1.GetLocalSet())
		assert.EqualValues(t, *server.GetLocalSet(), *client2.GetLocalSet())
	}
}

func TestIBLTSyncParameterMismatchReturnsErrors(t *testing.T) {
	server, err := NewIBLTSetSync(WithSymmetricSetDiff(1), WithDataLen(4))
	require.NoError(t, err)
	client, err := NewIBLTSetSync(WithSymmetricSetDiff(2), WithDataLen(4))
	require.NoError(t, err)

	require.NoError(t, server.AddElement([]byte("srvr")))
	require.NoError(t, client.AddElement([]byte("clnt")))

	port := freePort(t)
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.SyncServer("127.0.0.1", port)
	}()

	clientErr := client.SyncClient("127.0.0.1", port)
	require.Error(t, clientErr)
	assert.Contains(t, clientErr.Error(), "IBLT parameter mismatch")

	select {
	case err := <-serverErr:
		require.Error(t, err)
		assert.Contains(t, err.Error(), "IBLT parameter mismatch")
	case <-time.After(2 * time.Second):
		t.Fatal("server did not return after parameter mismatch")
	}
}

func TestIbltSync_SuccessRate(t *testing.T) {
	samples := 10
	maxSetSize := 1000
	maxElementSize := 1000
	failed := 0
	for index := 1; index <= samples; index++ {
		rand.Seed(int64(index))
		dataLen := rand.IntnRange(1, maxElementSize)
		serverSetSize := rand.IntnRange(0, maxSetSize)
		clientSetSize := rand.IntnRange(0, maxSetSize)
		intersectionSize := rand.IntnRange(0, func() int {
			if serverSetSize == 0 || clientSetSize == 0 {
				return 1
			}
			if serverSetSize > clientSetSize {
				return clientSetSize
			}
			return serverSetSize
		}())
		diffNum := serverSetSize - intersectionSize
		diffNum += clientSetSize - intersectionSize
		if diffNum == 0 {
			diffNum = 1
		}

		server, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum), WithDataLen(dataLen))
		require.NoError(t, err)

		client, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum), WithDataLen(dataLen))
		require.NoError(t, err)

		expectedSet := set.New()
		for i := 0; i < intersectionSize; i++ {
			td := []byte(rand.String(dataLen))
			err = server.AddElement(td)
			require.NoError(t, err)
			err = client.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		for i := 0; i < clientSetSize-intersectionSize; i++ {
			td := []byte(rand.String(dataLen))
			err = client.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		for i := 0; i < serverSetSize-intersectionSize; i++ {
			td := []byte(rand.String(dataLen))
			err = server.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		if !syncIBLTPairForSuccessRate(t, server, client) {
			failed++
			continue
		}
		diff := server.GetLocalSet().Difference(client.GetLocalSet())
		if diff.Len() != 0 {
			failed++
		}
	}
	t.Logf("IBLT success rate is %v", float32(samples-failed)/float32(samples))
}

func TestIbltSync_SuccessRateWithRetries(t *testing.T) {
	samples := 10
	maxSetSize := 1000
	maxElementSize := 1000
	retries := 5
	failed := 0
	for index := 1; index <= samples; index++ {
		rand.Seed(int64(index))
		dataLen := rand.IntnRange(1, maxElementSize)
		serverSetSize := rand.IntnRange(0, maxSetSize)
		clientSetSize := rand.IntnRange(0, maxSetSize)
		intersectionSize := rand.IntnRange(0, func() int {
			if serverSetSize == 0 || clientSetSize == 0 {
				return 1
			}
			if serverSetSize > clientSetSize {
				return clientSetSize
			}
			return serverSetSize
		}())
		diffNum := serverSetSize - intersectionSize
		diffNum += clientSetSize - intersectionSize
		if diffNum == 0 {
			diffNum = 1
		}

		server, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum), WithDataLen(dataLen), WithMaxSyncRetries(retries))
		require.NoError(t, err)

		client, err := NewIBLTSetSync(WithSymmetricSetDiff(diffNum), WithDataLen(dataLen), WithMaxSyncRetries(retries))
		require.NoError(t, err)

		expectedSet := set.New()
		for i := 0; i < intersectionSize; i++ {
			td := []byte(rand.String(dataLen))
			err = server.AddElement(td)
			require.NoError(t, err)
			err = client.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		for i := 0; i < clientSetSize-intersectionSize; i++ {
			td := []byte(rand.String(dataLen))
			err = client.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		for i := 0; i < serverSetSize-intersectionSize; i++ {
			td := []byte(rand.String(dataLen))
			err = server.AddElement(td)
			require.NoError(t, err)
			expectedSet.InsertKey(td)
		}

		if !syncIBLTPairForSuccessRate(t, server, client) {
			failed++
			continue
		}
		diff := server.GetLocalSet().Difference(client.GetLocalSet())
		if diff.Len() != 0 {
			t.Logf("Sync failed with %d difference.", diff.Len())
			failed++
		}
	}
	t.Logf("IBLT success rate with %d retries is %v", retries, float32(samples-failed)/float32(samples))
}

func syncIBLTPairForSuccessRate(t *testing.T, server, client interface {
	SyncClient(string, int) error
	SyncServer(string, int) error
}) bool {
	t.Helper()

	port := freePort(t)
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- client.SyncServer("127.0.0.1", port)
	}()
	clientErr := server.SyncClient("127.0.0.1", port)
	var remoteErr error
	select {
	case remoteErr = <-serverErr:
	case <-time.After(5 * time.Second):
		t.Log("sync server did not return")
		return false
	}
	if clientErr != nil || remoteErr != nil {
		t.Logf("sync transport failed: client=%v server=%v", clientErr, remoteErr)
		return false
	}
	return true
}

func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}
